package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"unsafe"
)

// var wg sync.WaitGroup

//go:generate stringer -type=Register,Perm,MemErrType -output=string.go

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: simpmulator <path/to/binary>")
	}
	var (
		path string
		err  error
	)
	if path, err = filepath.Abs(os.Args[1]); err != nil {
		log.Fatal(err)
	}
	_, binary := filepath.Split(path)
	fmt.Printf("PATH: %s\nFILENAME: %s\n", path, binary)

	// create a new emulator
	emu := NewEmulator(1024 * 1024 * 32)

	// map the elf binary into memory and set `pc` to the programs entry point
	if err := emu.LoadFile(path); err != nil {
		panic(err)
	}
	fmt.Printf("CURRENT ALLOCATION: %#x\n", emu.Mmu.curAlloc)

	// allocate stack and set the sp to the end of the memory. this is because
	// by convention the stack grows from from the higher memory to lower memory
	stack := emu.Allocate(64 * 1024)
	stackEnd := stack + 64*1024
	emu.SetHeap(stack) // set heap to end of stack
	emu.SetReg(Sp, uint64(stackEnd))

	fmt.Printf("STACK [%#x -> %#x]\n", stackEnd, stack)

	// set up null terminated string arg values.
	argv := emu.Allocate(uint(len(path) + 1))
	if err := emu.WriteFrom(argv, nullTerminate(binary)); err != nil {
		panic(err)
	}

	push(emu, uint64(0))    // auxp
	push(emu, uint64(0))    // envp
	push(emu, uint64(argv)) // argv
	push(emu, int32(1))     // argc

	// emu.Inspect(VirtAddr(emu.Reg(Sp)-0x100), 500)
	// os.Exit(1)

	if err := emu.Run(); err != nil {
		if eerr, ok := err.(EmuExit); ok {
			if merr, ok := eerr.cause.(MMUError); ok {
				emu.Inspect(merr.addr, merr.size+20)
				emu.InspectPerms(merr.addr, merr.size+20)
			}
		}
		panic(err)
	}
}

// push is a routine for pushing values onto the stack
func push[T Primitive](emu *Emulator, val T) {
	size := unsafe.Sizeof(val)
	sp := emu.Reg(Sp) - uint64(size)
	if err := WriteFromVal(emu.Mmu, VirtAddr(sp), val); err != nil {
		panic(err)
	}
	emu.SetReg(Sp, sp)
}

// null terminate strings before sticking them in memory
func nullTerminate(str string) []byte {
	return append([]byte(str), 0)
}
