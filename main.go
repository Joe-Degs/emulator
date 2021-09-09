package main

import (
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

	// Load test app into memory loading all the necessary sections
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

	//emu.ReadIntoPerms(VirtAddr(0x11190), buf, PERM_EXEC)
	emu.SetReg(Pc, 0x11190)
	emu.WriteFrom32(0, 0x4197)
	emu.Run()
}
