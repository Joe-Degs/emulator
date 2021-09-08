package main

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
