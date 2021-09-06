package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

// Perm represent permissions of memory addresses
type Perm uint8

// Set of permission types supported
const (
	PERM_READ  Perm = 1 << 0 // read permission
	PERM_WRITE      = 1 << 1 // write permission
	PERM_EXEC       = 1 << 2 // executable permission
	PERM_RAW        = 1 << 3 // read-after-write permission
)

type VirtAddr uint

// Mmu is an isolated memory space
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

// Allocate allocates region of memory as RW in the address space
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
	m.SetPermissions(base, size, PERM_RAW|PERM_WRITE)

	return base
}

// SetPermission sets the required permissions on memory locations starting
// from the	`addr` to `addr+size`
func (m *Mmu) SetPermissions(addr VirtAddr, size uint, perm Perm) {
	// set permissions for the allocated memory
	for i := uint(addr); i < uint(addr)+size; i++ {
		m.permissions[i] = perm
	}
}

// WriteFrom copies the buffer `buf` into memory checking the necessary
// permission before doing so
func (m *Mmu) WriteFrom(addr VirtAddr, buf []uint8) {
	//get the permission on the region of memory to write to
	perms := m.permissions[int(addr) : len(buf)+int(addr)]
	//fmt.Printf("%v", perms)

	hasRAW := false
	for _, p := range perms {
		// check if any part of the memory has is read-after-write
		hasRAW = hasRAW || ((p & PERM_RAW) != 0)
		// check if all perms are set to write
		if (p & PERM_WRITE) == 0 {
			return
		}
	}

	// copy the slice `buf` into memory pointed to by `addr`
	copy(m.memory[int(addr):len(buf)+int(addr)], buf)

	// update RAW permissions
	if hasRAW {
		for i, p := range perms {
			if (p & PERM_RAW) != 0 {
				perms[i] |= PERM_READ
			}
		}
		//copy(m.permissions[int(addr):len(perms)+int(addr)], perms)
	}
}

// ReadInto copies bytes of `len(buf)` from memory into a buffer, checking
// the necessary permissions before doing so
func (m *Mmu) ReadInto(addr VirtAddr, buf []uint8) {
	//get the permission on the region of memory to read from
	perms := m.permissions[int(addr) : len(buf)+int(addr)]
	//fmt.Printf("%v", perms)

	for _, p := range perms {
		// check if all perms on region of memory is read
		if (p & PERM_READ) == 0 {
			return
		}
	}
	// copy from the address pointed to by `addr` to len(buf) into `buf`
	copy(buf, m.memory[int(addr):len(buf)+int(addr)])
}

// Emulator keeps the state of the emulated system
type Emulator struct{ *Mmu }

func NewEmulator(size uint) *Emulator {
	return &Emulator{NewMmu(size)}
}

func main() {
	emu := NewEmulator(112)
	tmp := emu.Allocate(0x16)
	emu.WriteFrom(tmp, []byte("joe"))
	buf := make([]byte, 4)
	emu.ReadInto(tmp, buf)
	fmt.Println(string(buf))
	spew.Dump(emu)
}
