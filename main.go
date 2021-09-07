package main

import (
	"fmt"
	"sync"

	"github.com/davecgh/go-spew/spew"
)

// Perm represent permissions of memory addresses
type Perm uint8

// Enum of permission variants supported an another variant for keeping
// track of modified memory locations.
const (
	PERM_READ  Perm = 1 << 0 // read permission
	PERM_WRITE      = 1 << 1 // write permission
	PERM_EXEC       = 1 << 2 // executable permission
	PERM_RAW        = 1 << 3 // read-after-write permission

	DIRTY_BIT Perm = 1 << 4 // for tracking modified memory
)

// VirtAddr is a guest virtual address
type VirtAddr uint

// type VirtMemory struct {
// 	addr VirtAddr
// 	size uint
// }

// A block of memory mainly for tracking modified memory,
// this is highly inefficiant but it doesnt
// confuse me as the other mechanisms of keeping track of blocks
// so its cool
type Block = map[VirtAddr]uint

// Mmu is an isolated memory space
type Mmu struct {
	// memory is blob of memory space available to the system
	memory []uint8
	// access restrictions on individual locations in memory
	permissions []Perm
	// map of 16-byte aligned blocks of memory that have been modified
	dirty Block
	// tracks the current allocation
	curAlloc VirtAddr
}

func NewMmu(size uint) *Mmu {
	return &Mmu{
		memory:      make([]uint8, size),
		permissions: make([]Perm, size),
		dirty:       make(Block),
		curAlloc:    VirtAddr(0x38),
	}
}

// Reset restores all memory back to the original state. This allows us to
// create one emulator and fork it to run multiple things.
func (m *Mmu) Reset(other *Mmu) {
	for addr, size := range m.dirty {
		start := int(addr)
		end := start + int(size)

		// restore memory state
		copy(m.memory[start:end], other.memory[start:end])

		// clear dirty byte for restored portion of memory
		for i := start; i < end; i++ {
			m.permissions[i] &= ^DIRTY_BIT
		}

		// clear dirty list
		m.dirty = make(Block)
	}
}

// Fork an existing Mmu
func (m *Mmu) Fork() *Mmu {
	mmu := &Mmu{
		memory:      append(make([]uint8, 0, len(m.memory)), m.memory...),
		permissions: append(make([]Perm, 0, len(m.permissions)), m.permissions...),
		dirty:       make(Block),
		curAlloc:    m.curAlloc,
	}

	// clear dirty bits for new mmu
	for addr, size := range m.dirty {
		start := int(addr)
		end := int(size) + start

		for i := start; i < end; i++ {
			mmu.permissions[i] &= ^DIRTY_BIT
		}
	}

	return mmu
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
func (m *Mmu) WriteFrom(addr VirtAddr, buf []uint8) int {
	//get the permission on the region of memory to write to
	perms := m.permissions[int(addr) : len(buf)+int(addr)]
	//fmt.Printf("%v", perms)

	hasRAW, hasDirtyBitSet := false, false
	for _, p := range perms {
		// check if any part of the memory has is read-after-write
		hasRAW = hasRAW || ((p & PERM_RAW) != 0)

		// check if dirty bit is already set for memory location
		hasDirtyBitSet = hasDirtyBitSet || ((p & DIRTY_BIT) != 0)

		// check if all perms are set to write
		if (p & PERM_WRITE) == 0 {
			return 0
		}
	}

	// copy the slice `buf` into memory pointed to by `addr`
	n := copy(m.memory[int(addr):len(buf)+int(addr)], buf)

	// update RAW permissions and mark the written portion of memory as dirty
	if hasRAW {
		for i, p := range perms {
			if (p & PERM_RAW) != 0 {
				perms[i] |= PERM_READ
			}
			if !hasDirtyBitSet {
				perms[i] |= DIRTY_BIT
			}
		}
	}

	// 16-byte block that has been dirtied and put in in the dirty map
	blockStart := (int(addr) + 0xf) &^ 0xf
	if blockStart > int(addr) {
		blockStart -= 16
	}
	blockEnd := uint((n + 0xf) &^ 0xf)
	m.dirty[VirtAddr(blockStart)] = blockEnd
	return n
}

// ReadInto copies bytes of `len(buf)` from memory into a buffer, checking
// the necessary permissions before doing so
func (m *Mmu) ReadInto(addr VirtAddr, buf []uint8) (n int) {
	//get the permission on the region of memory to read from
	perms := m.permissions[int(addr) : len(buf)+int(addr)]
	//fmt.Printf("%v", perms)

	hasRAW := false
	for _, p := range perms {
		// check if any part of the memory has is read-after-write
		hasRAW = hasRAW || ((p & PERM_RAW) != 0)
		// check if all perms on region of memory is read
		if (p & PERM_READ) == 0 {
			return 0
		}
	}

	if hasRAW {
		for _, p := range perms {
			if (p & DIRTY_BIT) == 0 {
				return 0
			}
		}
	}

	// copy from the address pointed to by `addr` to len(buf) into `buf`
	n = copy(buf, m.memory[int(addr):len(buf)+int(addr)])
	return
}

// Emulator keeps the state of the emulated system
type Emulator struct{ *Mmu }

func NewEmulator(size uint) *Emulator { return &Emulator{NewMmu(size)} }

func main() {
	emu := NewEmulator(1024)
	tmp := emu.Allocate(4)
	emu.WriteFrom(tmp, []byte("asdf"))
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(emu *Mmu) {
		defer wg.Done()

		buf := make([]byte, 4)
		n := emu.ReadInto(tmp, buf)
		if n == 0 {
			panic(n)
		}
		fmt.Println(string(buf))
		spew.Dump(emu)
	}(emu.Fork())

	wg.Wait()
	spew.Dump(emu)
}
