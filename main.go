package main

import (
	"log"
	"reflect"
	"sync"
)

var wg sync.WaitGroup

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

	// set up null terminated arg values.
	argv := emu.Allocate(8)
	err = emu.WriteFrom(argv, []byte("tcat"))
	if err != nil {
		panic(err)
	}

	// stack push routine
	push := func(i interface{}) {
		isize := reflect.ValueOf(i).Type().Size()
		sp := emu.Reg(Sp) - uint64(isize)
		err = emu.WriteFromVal(VirtAddr(sp), i)
		if err != nil {
			panic(err)
		}
		emu.SetReg(Sp, sp)
	}

	push(uint64(0)) // auxp
	push(uint64(0)) // envp
	push(uint64(0)) // argv end
	push(uint64(argv))
	push(uint64(1)) // auxp

	err = emu.Run()
	if err != nil {
		panic(err)
	}
}
