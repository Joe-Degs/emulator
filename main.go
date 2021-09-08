package main

import (
	"fmt"
	"log"
	"sync"
)

var wg sync.WaitGroup

func main() {
	// Load the test elf binary at these sections and use it to test the emulator
	// "./hello" is the test binary file and it outputs hello world to screen
	// Elf file type is EXEC (Executable file)
	// Entry point 0x11190
	// There are 5 program headers, starting at offset 64
	//
	// Program Headers:
	//   Type           Offset             VirtAddr           PhysAddr
	//                  FileSiz            MemSiz              Flags  Align
	//   PHDR           0x0000000000000040 0x0000000000010040 0x0000000000010040
	//                  0x0000000000000118 0x0000000000000118  R      0x8
	//   LOAD           0x0000000000000000 0x0000000000010000 0x0000000000010000
	//                  0x0000000000000190 0x0000000000000190  R      0x1000
	//   LOAD           0x0000000000000190 0x0000000000011190 0x0000000000011190
	//                  0x000000000000255c 0x000000000000255c  R E    0x1000
	//   LOAD           0x00000000000026f0 0x00000000000146f0 0x00000000000146f0
	//                  0x00000000000000f8 0x0000000000000750  RW     0x1000
	//   GNU_STACK      0x0000000000000000 0x0000000000000000 0x0000000000000000
	//                  0x0000000000000000 0x0000000000000000  RW     0x0
	//
	//  Section to Segment mapping:
	//   Segment Sections...
	//    00
	//    01     .rodata
	//    02     .text
	//    03     .sdata .data .sbss .bss
	//    04

	emu := NewEmulator(32 * 1024 * 1024)

	// Load test app into memory loading all the necessary sections too
	err := emu.Load("./hello", []Section{
		Section{
			fileOffset:  0x0000000000000000,
			memSize:     0x0000000000000190,
			fileSize:    0x0000000000000190,
			virtualAddr: VirtAddr(0x0000000000010000),
			permissions: PERM_READ,
		},
		Section{
			fileOffset:  0x0000000000000190,
			memSize:     0x000000000000255c,
			fileSize:    0x000000000000255c,
			virtualAddr: VirtAddr(0x0000000000011190),
			permissions: PERM_READ | PERM_EXEC,
		},
		Section{
			fileOffset:  0x00000000000026f0,
			memSize:     0x0000000000000750,
			fileSize:    0x00000000000000f8,
			virtualAddr: VirtAddr(0x00000000000146f0),
			permissions: PERM_READ | PERM_WRITE,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	// read first 4 bytes of program start
	buf := make([]byte, 4)
	emu.ReadInto(VirtAddr(0x11190), buf)
	emu.SetReg(Pc, 0x11190)
	fmt.Printf("%#v, %#v\n", buf, emu.registers)
}

// func mainSpeed() {
// 	emu := NewEmulator(1024 * 1024)
// 	tmp := emu.Allocate(4)
// 	emu.WriteFrom(tmp, []byte("asdf"))
//
// 	wg.Add(1)
// 	go func(forked *Emulator) {
// 		defer wg.Done()
// 		for i := 0; i < 1000; i++ {
// 			forked.Reset(emu.Mmu)
// 		}
// 	}(emu.Fork())
// 	wg.Wait()
// }

// func mainTestWriteAndReads() {
// 	emu := NewEmulator(0x1000)
// 	tmp := emu.Allocate(4)
// 	emu.WriteFrom(tmp, []byte("asdf"))
//
// 	wg := sync.WaitGroup{}
//
// 	wg.Add(1)
// 	go func(forked *Emulator) {
// 		defer wg.Done()
//
// 		buf := make([]byte, 4)
// 		forked.WriteFrom(tmp, []byte("AAAA"))
// 		n := forked.ReadInto(tmp, buf)
// 		if n == 0 {
// 			panic(n)
// 		}
// 		fmt.Printf("dirtied: %x\n", buf)
// 		forked.Reset(emu.Mmu)
// 		buf = make([]byte, 4)
// 		forked.ReadInto(tmp, buf)
// 		fmt.Printf("after reset: %x\n", buf)
// 		spew.Dump(forked)
// 	}(emu.Fork())
//
// 	wg.Wait()
// }
