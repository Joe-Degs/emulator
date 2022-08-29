// memory mapping unit - contains logic to coordinate memory access and allocation
package main

import (
	"fmt"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
)

// Perm represent permissions of memory addresses
type Perm uint8

// Enum of variants of permission types on memory locations, the values
// correspond to the values of permissions of the elf file format.
const (
	PERM_EXEC  Perm = 0x1 // executable permission
	PERM_WRITE Perm = 0x2 // write permission
	PERM_READ  Perm = 0x4 // read permission
	PERM_RAW   Perm = 0x3 // read-after-write permission

	DIRTY_BLOCK_SIZE = 0x7f

	STACK_SIZE = 0x1000
	HEAP_SIZE  = 0x1000
)

// MemErrType represents the types of errors encountered during memory access
type MemErrType uint8

const (
	ErrCopy  MemErrType = iota // mem copy error
	ErrPerms                   // mem permission error
)

// MMUError contains values that make it easier to trace memory access errors
type MMUError struct {
	typ  MemErrType
	addr VirtAddr
	size uint
	perm Perm
}

func (m MMUError) Error() string {
	return fmt.Sprintf("MMUError{typ: %s, addr: %#v, size: %d, perm: %s}",
		m.typ, m.addr, m.size, m.perm)
}

// VirtAddr is any point in the program's address space
type VirtAddr uint

// Block is a block of memory, it maps the start of an allocated block to its
// end
type Block = map[VirtAddr]VirtAddr

// Mmu is an isolated memory space
type Mmu struct {
	// memory is blob of memory space available to the emulator
	memory []uint8

	// access restrictions on individual locations in memory
	permissions []Perm

	// map of modified blocks of memory
	dirty Block

	// tracks the current allocation
	curAlloc VirtAddr

	// the start of the stack
	stack VirtAddr

	// the start of the heap memory
	heap VirtAddr

	// keep track of the program start in memory
	programStart VirtAddr
}

// get the size of the memory
func (m Mmu) Len() int { return len(m.memory) }

func (m *Mmu) setHeap(addr VirtAddr) { m.heap = addr }

func (m *Mmu) setStack(addr VirtAddr) { m.stack = addr }

func (m Mmu) Heap() VirtAddr { return m.heap }

func (m Mmu) Stack() VirtAddr { return m.stack }

func NewMmu(size uint) *Mmu {
	return &Mmu{
		memory:       make([]uint8, size),
		permissions:  make([]Perm, size),
		dirty:        make(Block),
		curAlloc:     VirtAddr(0x100),
		programStart: 0,
	}
}

// Reset restores all memory back to the original state.
func (m *Mmu) Reset(other *Mmu) {
	for addr, endAddr := range m.dirty {
		start := int(addr)
		end := int(endAddr)

		// restore memory state
		copy(m.memory[start:end], other.memory[start:end])
		// restore permissions
		copy(m.permissions[start:end], other.permissions[start:end])
	}
	// clear dirty list
	m.dirty = make(Block)
}

// Fork an existing Mmu
func (m *Mmu) Fork() *Mmu {
	mmu := &Mmu{
		memory:       append(make([]uint8, 0, len(m.memory)), m.memory...),
		permissions:  append(make([]Perm, 0, len(m.permissions)), m.permissions...),
		dirty:        make(Block),
		curAlloc:     m.curAlloc,
		programStart: m.programStart,
	}
	return mmu
}

// Allocate region of memory as RW
func (m *Mmu) Allocate(size uint) VirtAddr {
	return m.AllocatePerms(size, PERM_READ|PERM_WRITE)
}

// Allocate memory with specified permissions
func (m *Mmu) AllocatePerms(size uint, perm Perm) VirtAddr {
	// 16-byte align the allocation
	alignSize := (size + 0xf) &^ 0xf

	base := m.curAlloc
	if int(base) >= len(m.memory) {
		return 0
	}
	m.curAlloc += VirtAddr(alignSize)

	// could not satisfy allocation without going out of memory
	if int(m.curAlloc) > len(m.memory) {
		m.curAlloc = base
		return 0
	}
	m.SetPermissions(base, size, perm)
	return base
}

// SetPermission sets the required permissions on memory locations starting
// from the	`addr` to `addr+size`
func (m *Mmu) SetPermissions(addr VirtAddr, size uint, perm Perm) {
	for i := uint(addr); i < uint(addr)+size; i++ {
		m.permissions[i] = perm
	}
}

// WriteFrom copies the buffer `buf` into memory checking the necessary
// permission before doing so
func (m *Mmu) WriteFrom(addr VirtAddr, buf []uint8) error {
	perms := m.permissions[int(addr) : len(buf)+int(addr)]

	hasRAW := false
	for _, p := range perms {
		// check if any part of the memory has read-after-write
		hasRAW = hasRAW || ((p & PERM_RAW) != 0)

		// check if all perms are set to write
		if (p & PERM_WRITE) == 0 {
			return MMUError{typ: ErrPerms, addr: addr, size: uint(len(buf))}
		}
	}

	// copy the slice `buf` into memory pointed to by `addr`
	n := copy(m.memory[int(addr):len(buf)+int(addr)], buf)

	// update permissions and allow reading after writing
	if hasRAW {
		for i, p := range perms {
			if (p & PERM_RAW) != 0 {
				perms[i] |= PERM_READ
			}
		}
	}

	// update the dirty block map
	// aligned block to keep track of modified memory
	blockStart := (int(addr) + DIRTY_BLOCK_SIZE) &^ DIRTY_BLOCK_SIZE
	round := DIRTY_BLOCK_SIZE + 1
	if blockStart > int(addr) {
		blockStart -= round
	}
	// align block to the dirty block size
	numBlocks := int((n+DIRTY_BLOCK_SIZE)&^DIRTY_BLOCK_SIZE) / round
	// end index of the aligned block
	blockEnd := VirtAddr(blockStart + (numBlocks * round) - 1)
	// add block `start - end` to dirty blocks map
	m.dirty[VirtAddr(blockStart)] = blockEnd
	return nil
}

func (m Mmu) copyBytes(addr VirtAddr, buf []uint8) error {
	// copy from the address pointed to by `addr` to len(buf) into `buf`
	if copy(buf, m.memory[int(addr):len(buf)+int(addr)]) != len(buf) {
		return MMUError{typ: ErrCopy, addr: addr, size: uint(len(buf))}
	}
	return nil
}

// ReadIntoPerms reads data of `len(buf)` from memory into buf only if the region
// of memory been read has `perm` set on it
func (m Mmu) ReadIntoPerms(addr VirtAddr, buf []uint8, perm Perm) error {
	//get the permission on the region of memory to read from
	perms := m.permissions[int(addr) : len(buf)+int(addr)]

	for _, p := range perms {
		// check if all perms on region of memory is expected perm
		if (p & perm) != perm {
			return MMUError{typ: ErrPerms, addr: addr, size: uint(len(buf)), perm: perm}
		}
	}
	return m.copyBytes(addr, buf)
}

func (m Mmu) Inspect(addr VirtAddr, size uint) {
	alignSize := (size + 0xf) &^ 0xf
	spew.Dump(m.memory[addr : uint(addr)+alignSize])
}

func (m Mmu) InspectPerms(addr VirtAddr, size uint) {
	alignSize := (size + 0xf) &^ 0xf
	spew.Dump(m.permissions[addr : uint(addr)+alignSize])
}

// ReadInto reads data of `len(buf)` from readable memory starting at addr into buf
func (m Mmu) ReadInto(addr VirtAddr, buf []uint8) error {
	return m.ReadIntoPerms(addr, buf, PERM_READ)
}

// Go generics are stable now, i've been waiting for it to become
// stable for sometime, its here so we change functions to use it
//
// Primitive is a generic type consisting of all integer types in go
type Primitive interface {
	uint8 | int8 | uint16 | int16 | uint32 | int32 | uint64 | int64
}

// ValToBytes converts a primitive interger type to sizeof(val) byte slice
func ValToBytes[T Primitive](val T) []byte {
	size := unsafe.Sizeof(val)
	if size == 1 {
		return (*[1]byte)(unsafe.Pointer(&val))[:]
	}

	buf := make([]byte, size)
	for i := uintptr(0); i < size; i++ {
		buf[i] = *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + i))
	}
	return buf
}

// WriteFromVal writes a T to virtual memory at address `addr`
func WriteFromVal[T Primitive](m *Mmu, addr VirtAddr, val T) error {
	return m.WriteFrom(addr, ValToBytes(val))
}

// ReadIntoVal reads a T from address `addr` from virtual memory, checking the
// necessary permission before doing so.
func ReadIntoValPerms[T Primitive](m *Mmu, addr VirtAddr, val T, perm Perm) (T, error) {
	buf := make([]byte, unsafe.Sizeof(val))
	if err := m.ReadIntoPerms(addr, buf, perm); err != nil {
		return 0, err
	}
	return *(*T)(unsafe.Pointer(&buf[0])), nil
}

// ReadIntoVal reads and returns a T from address `addr` in memory
func ReadIntoVal[T Primitive](m *Mmu, addr VirtAddr, val T) (T, error) {
	return ReadIntoValPerms(m, addr, val, PERM_READ)
}
