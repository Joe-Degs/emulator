package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

//go:generate stringer -type=Register,Perm,MemErrType -output=string.go

var (
	// verbose output
	VERBOSE = false

	// verbose output with pc and opcode
	VERBOSE_PC_OPCODE = false

	// size of the programs address space
	MEM_SIZE uint = 2 * 1024 * 1024
)

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

	emu := NewEmulator(MEM_SIZE)
	if err := emu.MapProgram(path, os.Args[1:]); err != nil {
		panic(err)
	}

	if VERBOSE {
		fmt.Printf("PATH: %s\nFILENAME: %s\n", emu.file.path, emu.file.name)
		fmt.Printf("MEM SIZE: %#x\n", MEM_SIZE-1)
		fmt.Printf("STACK [%#x -> %#x]\n", emu.Stack(), emu.Stack()-STACK_SIZE)
		fmt.Printf("HEAP [%#x -> %#x]\n", emu.Heap(), emu.Heap()+HEAP_SIZE)
		fmt.Printf("CURRENT ALLOCATION: %#x\n", emu.Mmu.curAlloc)
	}

	if err := emu.Run(); err != nil {
		handleErrors(emu, err)
	}
}

// handle emulator execution errors
func handleErrors(emu *Emulator, err error) {
	if e, ok := err.(EmuExit); ok {
		switch t := e.cause.(type) {
		case MMUError:
			emu.Inspect(t.addr, t.size)
			emu.InspectPerms(t.addr, t.size)
			return
		case SysCall:
			fmt.Fprintln(os.Stderr, t)
			return
		case Done:
			os.Exit(t.status)
		default:
			return
		}
	}
	log.Fatal(err)
}
