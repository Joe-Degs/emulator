// emulator logic - loads and maps files into memory and starts the
// fetch-decode-execute loop
package main

import (
	"debug/elf"
	"fmt"
	"os"
)

// Emulator keeps the state of the emulated system
type Emulator struct {
	*Mmu
	registers [33]uint64
}

// FileHeader holds elf file header information
type ElfBinary struct {
	filename string
	entry    uint64
	segments []elf.ProgHeader
}

func NewEmulator(size uint) *Emulator {
	return &Emulator{
		NewMmu(size), [33]uint64{},
	}
}

// Set the start of the program loaded into memory
func (e *Emulator) SetProgramStart(addr uint64) {
	e.programStart = VirtAddr(addr)
	e.SetReg(Pc, addr)
}

// Fork an emulator
func (e Emulator) Fork() *Emulator {
	return &Emulator{e.Mmu.Fork(), [33]uint64{}}
}

func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}

// Load executable binary into the emulator's address space
func (e *Emulator) LoadSegments(hdr ElfBinary) error {
	fileContents, err := os.ReadFile(hdr.filename)
	if err != nil {
		return err
	}

	for _, seg := range hdr.segments {
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

	// set the entry point of the program in memory
	e.SetProgramStart(hdr.entry)
	return nil
}

func (e *Emulator) LoadFile(path string) error {
	elfBinary, err := elf.Open(path)
	if err != nil {
		return err
	}

	bin := ElfBinary{
		filename: path,
		entry:    elfBinary.FileHeader.Entry,
		segments: make([]elf.ProgHeader, 0, len(elfBinary.Progs)),
	}
	for _, hdr := range elfBinary.Progs {
		typ := hdr.ProgHeader.Type
		if typ == elf.PT_LOAD || typ == elf.PT_PHDR {
			bin.segments = append(bin.segments, hdr.ProgHeader)
		}
	}
	return e.LoadSegments(bin)
}

func (e *Emulator) SetReg(reg Register, val uint64) {
	if reg == Zero {
		return
	}
	e.registers[reg] = val
}

func (e Emulator) Reg(reg Register) uint64 { return e.registers[reg] }

func (e *Emulator) IncPc() { e.SetReg(Pc, e.Reg(Pc)+4) }

// Given a register pointing into executable memory, this function
// reads a 32 bit unsigned value from that address
func (e *Emulator) ReadFromRegister(reg Register) (inst uint32, err error) {
	addr := VirtAddr(e.Reg(reg))
	return ReadIntoValPerms(e.Mmu, addr, inst, PERM_EXEC)
}

// NextInstAndOpcode gets the next instruction from memory and gets the opcode
func (e Emulator) NextInstAndOpcode() (inst uint32, opcode uint8, err error) {
	inst, err = e.ReadFromRegister(Pc)
	opcode = uint8(inst & 0b1111111)
	return
}

type EmuExit struct {
    cause error
    opcode uint8
    pc uint64
}

func (e EmuExit) Error() string {
    return fmt.Sprintf(
        "EmuExit {\n\t%s,\n\tpc: %#x,\n\topcode: %#08b\n}\n",
        e.cause.Error(), e.pc, e.opcode,
    )
}

// Run is the fetch - decode - execute loop
func (e *Emulator) Run() (err error) {
	for {
		inst, opcode, err := e.NextInstAndOpcode()
		if err != nil {
			return err
		}
		pc := e.Reg(Pc)

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
			target := e.Reg(inst.rs1) + uint64(int64(inst.imm&^1))
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
				if rs1 < rs2 {
					simm := uint64(int64(inst.imm)) + pc
					e.SetReg(Pc, simm)
					continue
				}
			case 0x4:
				// BGE
				if rs1 >= rs2 {
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
					simm := uint64(int64(inst.imm)) + pc
					e.SetReg(Pc, uint64(simm))
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
