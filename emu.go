package main

import "os"

// Emulator keeps the state of the emulated system
type Emulator struct {
	*Mmu
	registers [32]int
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
		NewMmu(size), [32]int{},
	}
}

// Fork an emulator
func (e Emulator) Fork() *Emulator {
	return &Emulator{e.Mmu.Fork(), [32]int{}}
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

func (e *Emulator) SetReg(reg Register, val int) {
	if reg == Zero {
		return
	}
	e.registers[reg] = val
}

func (e Emulator) GetReg(reg Register) int {
	return e.registers[reg]
}
