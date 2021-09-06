package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

// Set of permission types supported
const (
	PERM_READ  uint8 = 1 << 0 // read permission
	PERM_WRITE       = 1 << 1 // write permission
	PERM_EXEC        = 1 << 2 // executable permission
	PERM_RAW         = 1 << 3 // read-after-write permission
)

// Perm represent permissions of memory addresses
type Perm uint8

type VirtAddr uint

// An isolated memory space
type Mmu struct {
	memory      []uint8
	permissions []Perm
	curAlloc    VirtAddr
}

func NewMmu(size uint) *Mmu {
	return &Mmu{
		memory:      make([]uint8, size),
		permissions: make([]Perm, size),
		curAlloc:    VirtAddr(0x38),
	}
}

// allocate region of memory as RW in the address space
func (m *Mmu) Allocate(size uint) VirtAddr {
	// 16-byte align the allocation
	alignSize := (size + 0xf) &^ 0xf

	// get the base addr
	base := m.curAlloc

	// allocation is bigger than available memory
	if int(base) >= len(m.memory) {
		return 0
	}

	// update current allocation size
	m.curAlloc += VirtAddr(alignSize)

	// could not satisfy allocation without going out of memory
	if int(m.curAlloc) > len(m.memory) {
		// abort allocation and revert back to base
		m.curAlloc = base
		return 0
	}

	// mark memory as uninitialized and writable
	m.SetPermissions(base, size, Perm(PERM_RAW|PERM_WRITE))

	return base
}

func (m *Mmu) SetPermissions(addr VirtAddr, size uint, perm Perm) {
	// set permissions for the allocated memory
	for i := uint(addr); i < uint(addr)+size; i++ {
		m.permissions[i] = perm
	}
}

// WriteFrom writes from buffer into memory
func (m *Mmu) WriteFrom(addr VirtAddr, buf []uint8) {
	// copy the slice `buf` into memory pointed to by `addr`
	copy(m.memory[int(addr):len(buf)+int(addr)], buf)
}

// Read bytes from memory into a buffer
func (m *Mmu) ReadInto(addr VirtAddr, buf []uint8) {
	// copy from the address pointed to by `addr` to len(buf) into `buf`
	copy(buf, m.memory[int(addr):len(buf)+int(addr)])
}

// All the state of the emulated system
type Emulator struct{ *Mmu }

func NewEmulator(size uint) *Emulator {
	return &Emulator{NewMmu(size)}
}

func main() {
	emu := NewEmulator(112)
	emu.Allocate(0x16)
	emu.WriteFrom(emu.curAlloc, []uint8("joe"))
	buf := make([]uint8, 4)
	emu.ReadInto(emu.curAlloc, buf)
	fmt.Println(string(buf))
	spew.Dump(emu)
}
