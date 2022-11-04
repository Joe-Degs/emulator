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
	VERBOSE = true

	// verbose output with pc and opcode
	VERBOSE_PC_OPCODE = true

	VERBOSE_INST_DECODE = true

	// log emulator state on error
	LOG_STATE = true

	// dump the info of the loadable elf file
	DUMP_ELF_INFO = true

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
		fmt.Printf("PATH: %s\nFILENAME: %s\n", emu.program.path, emu.program.name)
		fmt.Printf("MEM SIZE: %#x\n", MEM_SIZE-1)
		fmt.Printf("STACK [%#x -> %#x]\n", emu.Stack(), emu.Stack()-STACK_SIZE)
		fmt.Printf("HEAP [%#x -> %#x]\n", emu.Heap(), emu.Heap()+HEAP_SIZE)
		fmt.Printf("CURRENT ALLOCATION: %#x\n", emu.Mmu.curAlloc)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(emu.String())
			log.Fatal("panic")
		}
	}()

	if err := emu.Run(); err != nil {
		handleErrors(emu, err)
	}
}

// handle emulator execution errors
func handleErrors(emu *Emulator, err error) {
	if e, ok := err.(EmuExit); ok {
		switch t := e.cause.(type) {
		case MMUError:
			if LOG_STATE {
				fmt.Println(e.Error())
			}
			emu.Inspect(t.addr, t.size)
			emu.InspectPerms(t.addr, t.size)
			return
		case Done:
			os.Exit(t.status)
		}
		return
	}
	panic(err)
}
