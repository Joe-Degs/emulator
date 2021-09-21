package main

import (
	"debug/elf"
	"fmt"
	"os"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
)

// Emulator keeps the state of the emulated system
type Emulator struct {
	*Mmu
	registers [33]uint64
}

// Section represents elf binary loadable program segments
type Section struct {
	fileOffset  uint
	memSize     uint
	fileSize    uint
	virtualAddr VirtAddr
	permissions Perm
}

// FileHeader holds elf file header information
type FileHeader struct {
	filename string
	entry    uint64
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
func (e *Emulator) LoadSections(hdr FileHeader, sections []Section) error {
	fileContents, err := os.ReadFile(hdr.filename)
	if err != nil {
		return err
	}

	for _, section := range sections {
		// set memory as writable
		e.SetPermissions(section.virtualAddr, section.memSize, PERM_WRITE)

		// write file contents into memory
		e.WriteFrom(section.virtualAddr,
			fileContents[section.fileOffset:section.fileOffset+section.fileSize])

		// write in any pads with zeros
		if section.memSize > section.fileSize {
			pad := make([]uint8, section.memSize-section.fileSize)
			e.WriteFrom(VirtAddr(uint(section.virtualAddr)+section.fileSize),
				pad)
		}
		// demote permissions to originals
		e.SetPermissions(section.virtualAddr, section.memSize, section.permissions)

		// update curAlloc beyond all sections and 16-byte align it
		e.curAlloc = VirtAddr(max(
			uint(e.curAlloc),
			(uint(section.virtualAddr)+section.memSize+0xf)&^0xf),
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

	// parse the file header containing the path to the program to run and
	// the entry point of the program
	fileHdr := FileHeader{
		filename: path,
		entry:    elfBinary.FileHeader.Entry,
	}

	loadableSecs := make([]Section, 0, len(elfBinary.Progs))
	fmt.Println(loadableSecs)

	for _, hdr := range elfBinary.Progs {
		// Get all the loadable sections of the program header
		if hdr.ProgHeader.Type == elf.PT_LOAD {
			// get program header
			ph := hdr.ProgHeader

			// parse the needed sections
			sec := Section{
				fileOffset:  uint(ph.Off),
				memSize:     uint(ph.Memsz),
				fileSize:    uint(ph.Filesz),
				virtualAddr: VirtAddr(ph.Vaddr),
			}

			// parsing program header permissions
			if ph.Flags&elf.PF_R == elf.PF_R {
				// add readable perms to section
				sec.permissions |= PERM_READ
			}

			if ph.Flags&elf.PF_W == elf.PF_W {
				// add writable perms to section
				sec.permissions |= PERM_WRITE
			}

			if ph.Flags&elf.PF_X == elf.PF_X {
				// add executable perms to section
				sec.permissions |= PERM_EXEC
			}

			// append to sections
			loadableSecs = append(loadableSecs, sec)
		}
	}

	// still need the entry point and the stack and all othe other stuff too

	e.LoadSections(fileHdr, loadableSecs)
	return nil
}

func (e *Emulator) SetReg(reg Register, val uint64) {
	if reg == Zero {
		return
	}
	e.registers[reg] = val
}

func (e Emulator) Reg(reg Register) uint64 { return e.registers[reg] }

func (e *Emulator) IncRegPc() { e.SetReg(Pc, e.Reg(Pc)+4) }

// Given a register pointing to into executable memory, this function
// reads a 32 bit unsigned value from that address
func (e *Emulator) ReadFromRegister(reg Register) (inst uint32, err error) {
	buf := make([]byte, 4)
	addr := VirtAddr(e.Reg(reg))
	if addr < e.programStart {
		//e.SetReg(reg, uint64(addr|0x1000))
		//addr = VirtAddr(e.Reg(reg))
	}
	err = e.ReadIntoPerms(addr, buf, PERM_EXEC)
	if err == nil {
		inst = *(*uint32)(unsafe.Pointer(&buf[0]))
	}
	return
}

// Run runs in a loop interpreting and running risc-v code
func (e *Emulator) Run() (err error) {
	for {
		// get the next instruction
		inst, err := e.ReadFromRegister(Pc)
		pc := e.Reg(Pc)
		opcode := inst & 0b1111111

		fmt.Printf("opcode: %#b -- pc: %#x\n", inst&0b1111111, pc)

		if err != nil {
			return err
		}

		switch opcode {
		case 0b0110011:
			// Rtype
			// register arithmetic
			inst := Decode(inst, Rtype{}).(Rtype)
			rs1 := e.Reg(inst.rs1)
			rs2 := e.Reg(inst.rs2)

			switch inst.funct3 | inst.funct7 {
			case 0x0:
				// ADD
				e.SetReg(inst.rd, rs1+rs2)
			case 0x20:
				// SUB
				e.SetReg(inst.rd, rs1-rs2)
			case 0x4:
				// XOR
				e.SetReg(inst.rd, rs1^rs2)
			case 0x6:
				// OR
				e.SetReg(inst.rd, rs1|rs2)
			case 0x7:
				// AND
				e.SetReg(inst.rd, rs1&rs2)
			case 0x1:
				// SLL
				shamt := rs2 & 0b111111
				e.SetReg(inst.rd, rs1<<shamt)
			case 0x5:
				// SRL
				shamt := rs2 & 0b111111
				e.SetReg(inst.rd, rs1>>shamt)
			case 0x5 | 0x20:
				// SRA
				shamt := rs2 & 0b111111
				e.SetReg(inst.rd, uint64(int64(rs1)>>shamt))
			case 0x2:
				// SLT
				if int64(rs1) < int64(rs2) {
					e.SetReg(inst.rd, 1)
				} else {
					e.SetReg(inst.rd, 0)
				}
			case 0x3:
				// SLTU
				if rs1 < rs2 {
					e.SetReg(inst.rd, 1)
				} else {
					e.SetReg(inst.rd, 0)
				}
			}
		case 0b0010011:
			// Itype
			// immediate arithmetic
			inst := Decode(inst, Itype{}).(Itype)
			rs1 := e.Reg(inst.rs1)
			imm := uint64(int64(inst.imm))
			switch inst.funct3 {
			case 0x0:
				// ADDI
				e.SetReg(inst.rd, rs1+imm)
			case 0x4:
				// XORI
				e.SetReg(inst.rd, rs1^imm)
			case 0x6:
				// ORI
				e.SetReg(inst.rd, rs1|imm)
			case 0x7:
				// ANDI
				e.SetReg(inst.rd, rs1&imm)
			case 0x1:
				// SLLI
				funct7 := (inst.imm >> 6) & 0b111111
				if funct7 == 0x0 {
					shamt := inst.imm & 0b111111
					e.SetReg(inst.rd, rs1<<shamt)
				} else {
					panic("unreachable slli")
				}
			case 0x5:
				funct7 := (inst.imm >> 5) & 0b111111
				shamt := inst.imm & 0b111111
				if funct7 == 0x0 {
					// SRLI
					e.SetReg(inst.rd, rs1>>shamt)
				} else if funct7 == 0x16 {
					// SRAI
					e.SetReg(inst.rd, uint64(int64(rs1)>>shamt))
				} else {
					panic("unreachable srai")
				}
			case 0x2:
				// SLTI
				if int64(rs1) < int64(imm) {
					e.SetReg(inst.rd, 1)
				} else {
					e.SetReg(inst.rd, 0)
				}
			case 0x3:
				// SLTIU
				if rs1 < imm {
					e.SetReg(inst.rd, 1)
				} else {
					e.SetReg(inst.rd, 0)
				}
			default:
				panic(fmt.Errorf("uimplemented Itype with opcode: %b\n", opcode))
			}

		case 0b0000011:
			// Itype
			// loads
			inst := Decode(inst, Itype{}).(Itype)
			addr := VirtAddr(e.Reg(inst.rs1) + uint64(int64(inst.imm)))

			switch inst.funct3 {
			case 0x0:
				// LB
				var val int8
				err = e.ReadIntoVal(addr, &val)
				if err != nil {
					return err
				}
				e.SetReg(inst.rd, uint64(int64(val)))
			case 0x1:
				// LH
				var val int16
				err = e.ReadIntoVal(addr, &val)
				if err != nil {
					return err
				}
				e.SetReg(inst.rd, uint64(int64(val)))
			case 0x2:
				// LW
				var val int32
				err = e.ReadIntoVal(addr, &val)
				if err != nil {
					return err
				}
				e.SetReg(inst.rd, uint64(int64(val)))
			case 0x3:
				// LD
				var val int64
				err = e.ReadIntoVal(addr, &val)
				if err != nil {
					return err
				}
				e.SetReg(inst.rd, uint64(int64(val)))
			case 0x4:
				// LBU
				var val uint8
				err = e.ReadIntoVal(addr, &val)
				if err != nil {
					return err
				}
				e.SetReg(inst.rd, uint64(int64(val)))
			case 0x5:
				// LHU
				var val uint16
				err = e.ReadIntoVal(addr, &val)
				if err != nil {
					return err
				}
				e.SetReg(inst.rd, uint64(int64(val)))
			case 0x6:
				// LWU
				var val uint32
				err = e.ReadIntoVal(addr, &val)
				if err != nil {
					return err
				}
				e.SetReg(inst.rd, uint64(int64(val)))
			}
		case 0b0100011:
			// Stype
			// stores
			inst := Decode(inst, Stype{}).(Stype)
			addr := VirtAddr(e.Reg(inst.rs1) + uint64(int64(inst.imm)))
			val := int64(e.Reg(inst.rs2))

			switch inst.funct3 {
			case 0x0:
				// SB
				err = e.WriteFromVal(addr, int8(val))
				if err != nil {
					return err
				}
			case 0x1:
				// SH
				err = e.WriteFromVal(addr, int16(val))
				if err != nil {
					return err
				}
			case 0x2:
				// SW
				err = e.WriteFromVal(addr, int32(val))
				if err != nil {
					return err
				}
			case 0x3:
				// SD
				fmt.Printf("addr: %x %x\n", addr, val)
				spew.Dump(inst)
				err = e.WriteFromVal(addr, int64(val))
				if err != nil {
					return err
				}
			}
		case 0b0110111:
			// Utype
			// LUI
			inst := Decode(inst, Utype{}).(Utype)
			e.SetReg(inst.rd, uint64(int64(inst.imm)))
		case 0b0010111:
			// Utype
			// AUIPC
			inst := Decode(inst, Utype{}).(Utype)
			e.SetReg(inst.rd, uint64(int64(inst.imm))+pc)
		case 0b1101111:
			// Jtype
			// JAL
			inst := Decode(inst, Jtype{}).(Jtype)
			e.SetReg(inst.rd, pc+4)
			e.SetReg(Pc, pc+uint64(int64(inst.imm)))
			continue
		case 0b1100111:
			// Itype
			// JALR
			inst := Decode(inst, Itype{}).(Itype)
			target := e.Reg(inst.rs1) + uint64(int64(inst.imm))
			// TODO(Joe)
			// this does not look right. Some jalr's are jumping beyond the
			// the program start `0x11190` in memory and this is the only sane
			// to bring them some sense.
			e.SetReg(inst.rd, pc+4)
			e.SetReg(Pc, (target &^ 1))
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
				if int64(rs1) == int64(rs2) {
					simm := pc + uint64(int64(inst.imm))
					e.SetReg(Pc, simm)
					continue
				}
			case 0x1:
				// BNE
				if int64(rs1) != int64(rs2) {
					simm := pc + uint64(int64(inst.imm))
					e.SetReg(Pc, simm)
					continue
				}
			case 0x2:
				// BLT
				if int64(rs1) < int64(rs2) {
					simm := pc + uint64(int64(inst.imm))
					e.SetReg(Pc, simm)
					continue
				}
			case 0x4:
				// BGE
				if int64(rs1) >= int64(rs2) {
					simm := pc + uint64(int64(inst.imm))
					e.SetReg(Pc, simm)
					continue
				}
			case 0x6:
				// BLTU
				if rs1 < rs2 {
					simm := pc + uint64(int64(inst.imm))
					e.SetReg(Pc, simm)
					continue
				}
			case 0x7:
				// BGEU
				if rs1 >= rs2 {
					simm := pc + uint64(int64(inst.imm))
					e.SetReg(Pc, simm)
					continue
				}
			}
		case 0b0011011:
			// Itype
			// 32-bit arithmetic
			inst := Decode(inst, Itype{}).(Itype)
			rs1 := uint32(e.Reg(inst.rs1))
			imm := uint32(inst.imm)

			switch inst.funct3 {
			case 0x0:
				// ADDIW
				e.SetReg(inst.rd, uint64(int64(int32(rs1+imm))))
			case 0x1:
				// SLLIW
				funct7 := (inst.imm >> 5) & 0b1111111
				if funct7 == 0x0 {
					shamt := inst.imm & 0b11111
					e.SetReg(inst.rd, uint64(int64(int32(rs1<<shamt))))
				} else {
					panic("unreachable slli")
				}
			case 0x5:
				funct7 := (inst.imm >> 5) & 0b1111111
				shamt := inst.imm & 0b11111
				if funct7 == 0x0 {
					// SRLIW
					e.SetReg(inst.rd, uint64(int64(int32(rs1>>shamt))))
				} else if funct7 == 0x16 {
					// SRAIW
					e.SetReg(inst.rd, uint64(int64(int32(rs1)>>shamt)))
				} else {
					panic("unreachable srai")
				}
			}
		case 0b0111011:
			// Rtype
			// register arithmetic
			inst := Decode(inst, Rtype{}).(Rtype)
			rs1 := uint32(e.Reg(inst.rs1))
			rs2 := uint32(e.Reg(inst.rs2))

			switch inst.funct3 | inst.funct7 {
			case 0x0:
				// ADDW
				e.SetReg(inst.rd, uint64(int64(int32(rs1+rs2))))
			case 0x20:
				// SUBW
				e.SetReg(inst.rd, uint64(int64(int32(rs1-rs2))))
			case 0x1:
				// SLLW
				shamt := rs2 & 0b11111
				e.SetReg(inst.rd, uint64(int64(int32(rs1<<shamt))))
			case 0x5:
				// SRLW
				shamt := rs2 & 0b11111
				e.SetReg(inst.rd, uint64(int64(int32(rs1>>shamt))))
			case 0x5 | 0x20:
				// SRAW
				shamt := rs2 & 0b11111
				e.SetReg(inst.rd, uint64(int64(int32(rs1)>>shamt)))
			}
		case 0b0001111:
			// FENCE
			return fmt.Errorf("fence\n")
		case 0b1110011:
			if inst == 0b00000000000000000000000001110011 {
				// ECALL
				return fmt.Errorf("ecall\n")
			} else if inst == 0b0000000000010000000000000001110011 {
				// EBREAK
				return fmt.Errorf("ebreak\n")
			}
		default:
			return fmt.Errorf("unhandled opcode: %#b", opcode)
		}

		npc := e.Reg(Pc)
		e.SetReg(Pc, npc+4)
		continue
	}
	return nil
}
