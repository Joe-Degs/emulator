package main

import (
	"fmt"
	"unsafe"
)

// syscalls is the syscall table, it maps the syscall number to the syscall
// function.
var syscalls = map[uint64]func(e *Emulator, s SysCall) error{
	64:  sys_write,
	66:  sys_writev,
	94:  sys_exit, // exit_group
	93:  sys_exit,
	214: sys_brk,
	222: sys_mmap,
}

// read len(buf) from virtual memory and write it to file descriptor
func (e *Emulator) write(fd int, addr VirtAddr, count int) MemErrType {
	buf := make([]byte, count)
	if err := e.ReadInto(addr, buf); err != nil {
		return err.(MMUError).typ
	}
	n, _ := fmt.Fprintf(e.files[fd], "%s", buf)
	return MemErrType(n)
}

// SysCall contains the syscall number and arguments. It also double as an
// error for when the syscall is not implemented.
type SysCall struct {
	num, a0, a1, a2, a3, a4, a5 uint64
}

func (s SysCall) Error() string {
	return fmt.Sprintf(
		"Syscall{num: %d, a0: %d, a1: %d, a2: %d, a3: %d, a4: %d, a5: %d}",
		s.num, s.a0, s.a1, s.a2, s.a3, s.a4, s.a5,
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
		e.Reg(A3), e.Reg(A4), e.Reg(A5),
	}
	return syscall.execute(e)
}

func (e *Emulator) RetVal(ret uint64) { e.SetReg(A0, ret) }

// mmap syscall number 222
func sys_mmap(e *Emulator, s SysCall) error {
	e.RetVal(uint64(e.Heap()))
	return nil
}

// ssize_t write(int fd, const void *buf, size_t count)
func sys_write(e *Emulator, s SysCall) error {
	n := e.write(int(s.a0), VirtAddr(s.a1), int(s.a2))
	e.RetVal(uint64(n))
	return nil
}

// void _exit(int status);
func sys_exit(e *Emulator, s SysCall) error {
	fmt.Printf("exiting with %d\n", s.a0)
	return Done{int(s.a0)}
}

// A C scather/gather vector type
// the good thing is I think struct packing works the same way in C as in go
// if that is not the case, then I'm fucked royally (i wouldn't know what to do)
type iovec struct {
	base VirtAddr
	vlen int
}

// ssize_t writev(int fd, const struct iovec *iov, int iovcnt);
// write from a scatter vector
func sys_writev(e *Emulator, s SysCall) error {
	size := unsafe.Sizeof(iovec{})
	buf := make([]byte, size)
	addr := VirtAddr(s.a1)
	var n int
	for i := 0; i < int(s.a2); i++ {
		if err := e.ReadIntoPerms(addr, buf, PERM_READ); err != nil {
			return err
		}
		iov := (*iovec)(unsafe.Pointer(&buf[0]))
		if a := e.write(int(s.a0), iov.base, iov.vlen); a < 0 {
			return MMUError{typ: a, addr: iov.base}
		} else {
			n += int(a)
		}
		addr += VirtAddr(size)
	}
	e.RetVal(uint64(n))
	return nil
}

// the `brk` syscall is used to extend the program break essentially allocating
// more space in the data segment for use by the program
func sys_brk(e *Emulator, s SysCall) error {
	base := e.Allocate(0)
	var incr int64
	if s.a0 != 0 {
		incr = int64(s.a0) - int64(base)
	}

	if incr >= 0 {
		base := e.Allocate(uint(incr))
		// fmt.Printf("Incr: %#x, Base: %#x, Arg: %#x\n", incr, base, s.a0)
		e.SetReg(A0, uint64(base+VirtAddr(incr)))
	} else {
		e.SetReg(A0, ^uint64(0))
	}

	// e.SetReg(A0, 0)
	return nil
}
