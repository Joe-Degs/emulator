package main

import (
	"log"
	"sync"
	"unsafe"
)

var wg sync.WaitGroup

//go:generate stringer -type=Register,Perm -output=string.go

func main() {
	emu := NewEmulator(32 * 1024 * 1024)

	// map the elf binary into memory and set `pc` to the programs
	// entry point
	err := emu.LoadFile("./testdata/tcat/tcat")

	if err != nil {
		log.Fatal(err)
	}

	stack := emu.Allocate(32 * 1024)
	emu.SetReg(Sp, uint64(stack)+32*1024)

	// set up null terminated string arg values.
	argv := emu.Allocate(8)
	err = emu.WriteFrom(argv, nullTerminate("tcat"))
	if err != nil {
		panic(err)
	}

	push(emu, uint64(0)) // auxp
	push(emu, uint64(0)) // envp
	push(emu, uint64(0)) // argv end
	push(emu, uint64(argv))
	push(emu, uint64(1)) // auxp

	err = emu.Run()
	if err != nil {
		panic(err)
	}
}

// stack push routine
func push[T Primitive](emu *Emulator, val T) {
	size := unsafe.Sizeof(val)
	sp := emu.Reg(Sp) - uint64(size)
	err := WriteFromVal(emu.Mmu, VirtAddr(sp), val)
	if err != nil {
		panic(err)
	}
	emu.SetReg(Sp, sp)
}

func nullTerminate(str string) []byte {
	return append([]byte(str), 0)
}
