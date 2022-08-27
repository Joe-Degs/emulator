package main

import "fmt"

// syscalls is the syscall table, it maps the syscall number to the syscall
// function.
var syscalls = map[uint64]func(e *Emulator, s SysCall) error{
	222: mmap,
	64:  write,
	94:  exit, // exit_group
	93:  exit,
}

// SysCall contains the syscall number and arguments. It also double as an
// error for when the syscall is not implemented.
type SysCall struct {
	num, a0, a1, a2, a3, a4, a5, a6 uint64
}

func (s SysCall) Error() string {
	return fmt.Sprintf(
		"Syscall{num: %d, a0: %d, a1: %d, a2: %d, a3: %d, a4: %d, a5: %d, a6: %d}",
		s.num, s.a0, s.a1, s.a2, s.a3, s.a4, s.a5, s.a6,
	)
}

// find the function that executes the syscall
func (s SysCall) execute(e *Emulator) error {
	if syscall, ok := syscalls[s.num]; ok {
		return syscall(e, s)
	}
	return s
}

// TrapIntoSystem prepares the system for syscall execution.
func (e *Emulator) TrapIntoSystem() error {
	syscall := SysCall{
		e.Reg(A7), e.Reg(A0), e.Reg(A1), e.Reg(A2),
		e.Reg(A3), e.Reg(A4), e.Reg(A5), e.Reg(A6),
	}
	return syscall.execute(e)
}

// mmap syscall number 222
func mmap(e *Emulator, s SysCall) error {
	e.SetReg(A0, uint64(e.Heap()))
	return nil
}

// ssize_t write(int fd, const void *buf, size_t count)
func write(e *Emulator, s SysCall) error {
	addr := VirtAddr(s.a1)
	buf := make([]byte, s.a2)
	if err := e.ReadIntoPerms(addr, buf, PERM_READ); err != nil {
		return err
	}
	fmt.Printf("%s", buf)
	return nil
}

// void _exit(int status);
func exit(e *Emulator, s SysCall) error {
	return Done{int(s.a0)}
}
