package main

// func previous_main() {
// 	// Load the test elf binary at these sections and use it to test the emulator
// 	// "./hello" is the test binary file and it outputs hello world to screen
// 	// Elf file type is EXEC (Executable file)
// 	// Entry point 0x11190
// 	// There are 5 program headers, starting at offset 64
// 	//
// 	// Program Headers:
// 	//   Type           Offset             VirtAddr           PhysAddr
// 	//                  FileSiz            MemSiz              Flags  Align
// 	//   PHDR           0x0000000000000040 0x0000000000010040 0x0000000000010040
// 	//                  0x0000000000000118 0x0000000000000118  R      0x8
// 	//   LOAD           0x0000000000000000 0x0000000000010000 0x0000000000010000
// 	//                  0x0000000000000190 0x0000000000000190  R      0x1000
// 	//   LOAD           0x0000000000000190 0x0000000000011190 0x0000000000011190
// 	//                  0x000000000000255c 0x000000000000255c  R E    0x1000
// 	//   LOAD           0x00000000000026f0 0x00000000000146f0 0x00000000000146f0
// 	//                  0x00000000000000f8 0x0000000000000750  RW     0x1000
// 	//   GNU_STACK      0x0000000000000000 0x0000000000000000 0x0000000000000000
// 	//                  0x0000000000000000 0x0000000000000000  RW     0x0
// 	//
// 	//  Section to Segment mapping:
// 	//   Segment Sections...
// 	//    00
// 	//    01     .rodata
// 	//    02     .text
// 	//    03     .sdata .data .sbss .bss
// 	//    04
//
// 	emu := NewEmulator(32 * 1024 * 1024)
//
// 	// Load test app into memory loading all the necessary sections
// 	err := emu.LoadFile("./testdata/tcat/tcat")
//
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	//emu.ReadIntoPerms(VirtAddr(0x11190), buf, PERM_EXEC)
// 	//emu.SetProgramStart(0x11190)
//
// 	// set up a stack
// 	stack := emu.Allocate(32 * 1024)
// 	emu.SetReg(Sp, uint64(stack)+32*1024) // set sp to bottom of stack
//
// 	// set up null terminated arg values.
// 	argv := emu.Allocate(8)
// 	err = emu.WriteFrom(argv, []byte("hello\\0"))
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// stack push routine
// 	push := func(i interface{}) {
// 		isize := reflect.ValueOf(i).Type().Size()
// 		sp := emu.Reg(Sp) - uint64(isize)
// 		err = emu.WriteFromVal(VirtAddr(sp), i)
// 		if err != nil {
// 			panic(err)
// 		}
// 		emu.SetReg(Sp, sp)
// 	}
//
// 	push(uint64(0)) // auxp
// 	push(uint64(0)) // envp
// 	push(uint64(0)) // argv end
// 	push(uint64(argv))
// 	push(uint64(1)) // auxp
//
// 	err = emu.Run()
// 	if err != nil {
// 		panic(err)
// 	}
// }

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
