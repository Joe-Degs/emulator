package main

func (e *Emulator) TrapIntoSystem() error {
	switch e.Reg(A7) {
	case 222:
		//mmap
		e.SetReg(A0, uint64(e.Heap()))
	}
	return nil
}
