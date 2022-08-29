// emulator logic - loads and maps files into memory and starts the
// fetch-decode-execute loop
package main

import (
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

// Emulator keeps the state of the emulated system in this case a machine of
// RV64I architecture
type Emulator struct {
	*Mmu
	file      ElfBinary
	registers [33]uint64
}

// ElfBinary holds data necessary to succefully prepare program for execution.
type ElfBinary struct {
	path, name string
	args       []string
	entry      uint64
	segments   []elf.ProgHeader
}

// create a new emulator
func NewEmulator(size uint) *Emulator {
	return &Emulator{
		Mmu: NewMmu(size),
	}
}

// Set the address at which to start program execution
func (e *Emulator) setPC(addr uint64) {
	e.programStart = VirtAddr(addr)
	e.SetReg(Pc, addr)
}

// create an identical copy of the emulator
func (e Emulator) Fork() *Emulator {
	return &Emulator{
		Mmu:  e.Mmu.Fork(),
		file: e.file,
	}
}

func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}

// Load executable binary into the emulator's memory unit for execution.
func (e *Emulator) loadSegments() error {
	fileContents, err := os.ReadFile(e.file.path)
	if err != nil {
		return err
	}

	for _, seg := range e.file.segments {
		// set memory as writable
		alignedSize := (seg.Memsz + seg.Align) &^ seg.Align
		e.SetPermissions(VirtAddr(seg.Vaddr), uint(alignedSize), PERM_WRITE)

		// write file contents into memory
		if err := e.WriteFrom(VirtAddr(seg.Vaddr), fileContents[seg.Off:seg.Off+seg.Filesz]); err != nil {
			return err
		}

		// fill-in any pads with zeros
		if seg.Memsz > seg.Filesz {
			pad := make([]uint8, seg.Memsz-seg.Filesz)
			if err := e.WriteFrom(VirtAddr(seg.Vaddr+seg.Filesz), pad); err != nil {
				return err
			}
		}
		// demote permissions to originals
		e.SetPermissions(VirtAddr(seg.Vaddr), uint(alignedSize), Perm(seg.Flags))

		// update curAlloc beyond all sections and 16-byte align it
		e.curAlloc = VirtAddr(
			max(uint(e.curAlloc), uint(seg.Vaddr+seg.Memsz+seg.Align)&^uint(seg.Align)),
		)
	}
	e.setPC(e.file.entry)
	return nil
}

// reserve space in memory for static and dynamic objects (stack and heap).
func (e *Emulator) allocStackAndHeap() {
	// stack starts at a 16-byte address 255 steps away from last address.
	e.setStack(VirtAddr((e.Len() - 0xff) &^ 0xf))
	e.SetReg(Sp, uint64(e.Stack()))
	// calculate what to add to sp to get to end of memory.
	end := uint(e.Len()-int(e.Stack()-STACK_SIZE)) - 1
	e.SetPermissions(e.Stack()-STACK_SIZE, end, PERM_READ|PERM_WRITE)
	e.setHeap(e.Allocate(HEAP_SIZE))
}

// This is what a program looks like in memory
//
//	SP<-+---------------+->0xff...(end)
//     |               |
//     |     stack     |
//     |               |
//     +---------------+
//     |               |
//     |    (unused)   |
//     |               |
//     +---------------+
//     |               |
//     |     heap      |
//     |               |
//     +---------------+
//     |               |
//     |  static data  |
//     |               |
//     +---------------+
//     |               |
//     | section .text |
//     |  code segment |
//     |               |
//     +---------------+->0x0(start)
//
// MapProgram maps the executable elf file into memory to create a process image.
// It also sets the heap and stack start points basically setting everything up
// for execution to begin. it does the work of execv
// int execv(const char *pathname, char *const argv[]);
func (e *Emulator) MapProgram(path string, args []string) error {
	elfBinary, err := elf.Open(path)
	if err != nil {
		return err
	}
	_, name := filepath.Split(path)

	bin := ElfBinary{
		name:     name,
		path:     path,
		args:     args,
		entry:    elfBinary.FileHeader.Entry,
		segments: make([]elf.ProgHeader, 0, len(elfBinary.Progs)),
	}
	for _, hdr := range elfBinary.Progs {
		typ := hdr.ProgHeader.Type
		if typ == elf.PT_LOAD {
			bin.segments = append(bin.segments, hdr.ProgHeader)
		}
	}
	e.file = bin
	if err = e.loadSegments(); err != nil {
		return err
	}
	e.allocStackAndHeap()

	// insert name of executable as first argument in vector
	args = append(args[:0], append([]string{e.file.name}, args[0:]...)...)
	vec := nullTerminateArgs(args)
	size := (len(vec) + 0xf) &^ 0xf
	argv := VirtAddr(e.Reg(Sp) - uint64(size))
	if err := e.WriteFrom(argv, vec); err != nil {
		return err
	}
	// set up the stack for the main function
	// int main(int argc, char *argv[], char *envp[])
	err = push(e, uint64(0))          // evnp
	err = push(e, uint64(argv))       // argv
	err = push(e, int32(len(args)+1)) // argc
	return err
}

// push is a routine for pushing values onto the stack
func push[T Primitive](emu *Emulator, val T) error {
	size := unsafe.Sizeof(val)
	sp := emu.Reg(Sp) - uint64(size)
	if err := WriteFromVal(emu.Mmu, VirtAddr(sp), val); err != nil {
		return err
	}
	emu.SetReg(Sp, sp)
	return nil
}

// null terminate arguments variables for the program
func nullTerminateArgs(args []string) []byte {
	var strs []byte
	for _, str := range args {
		s := append([]byte(str), 0)
		strs = append(strs, s...)
	}
	return strs
}

// Set the specified registers value
func (e *Emulator) SetReg(reg Register, val uint64) {
	if reg == Zero {
		return
	}
	e.registers[reg] = val
}

// Reg returns the value in the specified register.
func (e Emulator) Reg(reg Register) uint64 { return e.registers[reg] }

// IncPc moves the program counter to the next instruction
func (e *Emulator) IncPc() { e.SetReg(Pc, e.Reg(Pc)+4) }

// Given a register pointing into executable memory, this function
// reads a 32 bit unsigned value from that address
func (e *Emulator) ReadFromRegister(reg Register) (inst uint32, err error) {
	addr := VirtAddr(e.Reg(reg))
	return ReadIntoValPerms(e.Mmu, addr, inst, PERM_EXEC)
}

// NextInstAndOpcode gets the next instruction and opcode from memory
func (e Emulator) NextInstAndOpcode() (inst uint32, opcode uint8, err error) {
	inst, err = e.ReadFromRegister(Pc)
	opcode = uint8(inst & 0b1111111)
	return
}

// EmuExit signals a pause or end of execution by the emulator
type EmuExit struct {
	cause  error
	opcode uint8
	pc     uint64
}

func (e EmuExit) Error() string {
	return fmt.Sprintf(
		"EmuExit {\n\t%s,\n\tpc: %#x,\n\topcode: %#08b\n}\n",
		e.cause.Error(), e.pc, e.opcode,
	)
}

// Done signals the emulator when the program executing exits succesfully
type Done struct{ status int }

func (d Done) Error() string {
	return fmt.Sprintf("exited with %d", d.status)
}

// Run is the fetch - decode - execute loop (it gets the next instruction,
// decodes it and performs the operations encoded into the instruction)
func (e *Emulator) Run() (err error) {
	for {
		inst, opcode, err := e.NextInstAndOpcode()
		if err != nil {
			return err
		}
		pc := e.Reg(Pc)

		if VERBOSE_PC_OPCODE {
			fmt.Printf("opcode: %#08b, pc: %#x\n", opcode, pc)
		}

		switch opcode {
		case 0b0110011:
			// rtype - register - register arithmetic
			e.decodeRtypeArith(inst)
		case 0b0010011:
			// itype - register - immediate arithmetic
			e.decodeItypeImmArith(inst)
		case 0b0000011:
			// itype - memory loads
			if err := e.decodeItypeLoads(inst); err != nil {
				return EmuExit{err, opcode, pc}
			}
		case 0b0100011:
			// stype - memory stores
			e.decodeStypeStore(inst)
		case 0b0110111:
			// Utype
			// LUI
			inst := Decode(inst, Utype{}).(Utype)
			e.SetReg(inst.rd, uint64(int64(inst.imm<<12)))
		case 0b0010111:
			// Utype
			// AUIPC
			inst := Decode(inst, Utype{}).(Utype)
			e.SetReg(inst.rd, uint64(int64(inst.imm<<12))+pc)
		case 0b1101111:
			// Jtype
			// JAL
			inst := Decode(inst, Jtype{}).(Jtype)
			e.SetReg(inst.rd, pc+4)
			e.SetReg(Pc, uint64(int64(inst.imm))+pc)
			continue
		case 0b1100111:
			// Itype
			// JALR
			inst := Decode(inst, Itype{}).(Itype)
			target := e.Reg(inst.rs1) + uint64(int64(inst.imm))
			e.SetReg(inst.rd, pc+4)
			e.SetReg(Pc, target)
			continue
		case 0b1100011:
			// Btype
			// conditional branches
			inst := Decode(inst, Btype{}).(Btype)
			rs1 := e.Reg(inst.rs1)
			rs2 := e.Reg(inst.rs2)

			switch inst.funct3 {
			case 0x0:
				// BEQ
				if rs1 == rs2 {
					simm := uint64(int64(inst.imm)) + pc
					e.SetReg(Pc, simm)
					continue
				}
			case 0x1:
				// BNE
				if rs1 != rs2 {
					simm := uint64(int64(inst.imm)) + pc
					e.SetReg(Pc, simm)
					continue
				}
			case 0x2:
				// BLT
				if int64(rs1) < int64(rs2) {
					simm := uint64(int64(inst.imm)) + pc
					e.SetReg(Pc, simm)
					continue
				}
			case 0x4:
				// BGE
				if int64(rs1) >= int64(rs2) {
					simm := uint64(int64(inst.imm)) + pc
					e.SetReg(Pc, simm)
					continue
				}
			case 0x6:
				// BLTU
				if rs1 < rs2 {
					simm := uint64(int64(inst.imm)) + pc
					e.SetReg(Pc, uint64(simm))
					continue
				}
			case 0x7:
				// BGEU
				if rs1 >= rs2 {
					simm := uint64(inst.imm) + pc
					e.SetReg(Pc, simm)
					continue
				}
			}
		case 0b0011011:
			// itype 32bit register-immediate arithmetic
			e.decodeItype32bitArith(inst)
		case 0b0111011:
			// rtype register-register arithmetic
			e.decodeRtype32RegArith(inst)
		case 0b0001111:
			// FENCE
			return EmuExit{fmt.Errorf("fence\n"), opcode, pc}
		case 0b1110011:
			if inst == 0b00000000000000000000000001110011 {
				// ECALL
				if err := e.TrapIntoSystem(); err != nil {
					return EmuExit{err, opcode, pc}
				}
			} else if inst == 0b0000000000010000000000000001110011 {
				// EBREAK
				return EmuExit{fmt.Errorf("ebreak\n"), opcode, pc}
			}
		default:
			return fmt.Errorf("unhandled opcode: %#b", opcode)
		}
		e.IncPc()
	}
	return nil
}
