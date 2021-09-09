package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

// Emulator keeps the state of the emulated system
type Emulator struct {
	*Mmu
	registers [32]uint64
}

// Section represents elf binary section
type Section struct {
	fileOffset  uint
	memSize     uint
	fileSize    uint
	virtualAddr VirtAddr
	permissions Perm
}

func NewEmulator(size uint) *Emulator {
	return &Emulator{
		NewMmu(size), [32]uint64{},
	}
}

// Fork an emulator
func (e Emulator) Fork() *Emulator {
	return &Emulator{e.Mmu.Fork(), [32]uint64{}}
}

func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}

// Load executable binary into the emulator's address space
func (e *Emulator) Load(path string, sections []Section) error {
	fileContents, err := os.ReadFile(path)
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

// Run runs in a loop interpreting and running risc-v code
func (e *Emulator) Run() (err error) {
	for i := 0; i < 5; i++ {
		// get the next instruction
		inst, err := e.ReadInto32(VirtAddr(e.Reg(Pc)), PERM_EXEC)

		if err != nil {
			return err
		}

		switch opcode := inst & 0b1111111; opcode {
		case 0b0110011:
			// Rtype
			// register arithmetic
			inst := DecodeInstruction(inst, Rtype{}).(Rtype)
			spew.Dump(inst)
			goto NextInst
		case 0b0010011:
			// Itype
			// immediate arithmetic
			inst := DecodeInstruction(inst, Itype{}).(Itype)
			spew.Dump(inst)
			goto NextInst
		case 0b0000011:
			// Itype
			// loads
			inst := DecodeInstruction(inst, Itype{}).(Itype)
			spew.Dump(inst)
			goto NextInst
		case 0b0100011:
			// Stype
			// stores
			inst := DecodeInstruction(inst, Stype{}).(Stype)
			spew.Dump(inst)
			goto NextInst
		case 0b0110111:
			// Utype
			// LUI
			inst := DecodeInstruction(inst, Utype{}).(Utype)
			spew.Dump(inst)
			goto NextInst
		case 0b0010111:
			// Utype
			// AUIPC
			inst := DecodeInstruction(inst, Utype{}).(Utype)
			spew.Dump(inst)
			goto NextInst
		case 0b1101111:
			// Jtype
			// JAL
			inst := DecodeInstruction(inst, Jtype{}).(Jtype)
			spew.Dump(inst)
			goto NextInst
		case 0b1100111:
			// Itype
			// JALR
			inst := DecodeInstruction(inst, Itype{}).(Itype)
			spew.Dump(inst)
			goto NextInst
		case 0b1100011:
			// Btype
			// conditional branches
			inst := DecodeInstruction(inst, Btype{}).(Btype)
			spew.Dump(inst)
			goto NextInst
		default:
			panic(fmt.Errorf("unhandled opcode: %#b", opcode))
		}
	NextInst:
		fmt.Println("executing next inst")
		e.IncRegPc()
		continue
	}
	return nil
}
