## simpmulator

A simple [RV64I](https://book.rvemu.app/instruction-set/01-rv64i.html) RISC-V emulator.
It emulates a simple linux environment running on the 64bit base instruction set of the
RISC-V architecture.

### Building and Testing
The toolchain for building test applications for this project can be installed
using instructions from [here](https://github.com/Joe-Degs/riscv). After installing
and building the toolchain use [this](https://github.com/Joe-Degs/riscv/tree/master/projects),
it contains instructions on how to generate makefiles for compiling test apps to run
against the emulator.

Any RV64I (base instruction set) toolchain should work too (i have not tried any
myself). Any elf binary compiled with the above mentioned toolchain should work
with little work.

The `testdata/` directory contains sample source files, makefiles, and compiled
binaries used to test the emulator during its development.

The emulator is a go program so running is as trivial as executing
```
go run . <path/to/test-binary>
```
You could also just execute the `run.sh` to run the hello-world test program.

### Next Steps
For now we are able to run simple programs in the emulator, but you will hit a panic when
you encounter a syscall that is not yet implemented. The panic contains info on the syscall
that is causing the error and the contents of its argument registers.
```
PATH: /home/joe/go/src/github.com/Joe-Degs/emulator/testdata/hello/hello
FILENAME: hello
CURRENT ALLOCATION: 0x140b0
STACK [0x240b0 -> 0x140b0]
panic: Syscall{num: 64, a0: 1, a1: 77416, a2: 12, a3: 0, a4: 7, a5: 70816}


goroutine 1 [running]:
main.main()
        /home/joe/go/src/github.com/Joe-Degs/emulator/main.go:62 +0x5c5
exit status 2
```
The panic up above is caused by the absense of syscall number 64, there is also
the values of argument register `a0-a5`. Syscalls are implemented in the file
`syscall.go` and all syscalls are functions of type;
```go
func(*Emulator, SysCall) error

```
Implementing the syscalls should be trivial, but I have
quite a lot on my plate right now. So I'll be doing it incrementally when I have
new test programs that request unimplemented syscalls.
Syscall names and their respective numbers can be found in `syscall-numbers.txt`
So implement a syscall, stick in the syscall table `syscalls` in the `syscall.go`
and you are good to go.

### resources
This project is inspired by [gamozolabs'](https://github.com/gamozolabs) [Fuzz
Week 2020](https://gamozolabs.github.io/2020/07/12/fuzz_week_2020.html).
Check out his [youtube](https://youtube.com/user/gamozolabs).

- [risc-v learning resources](https://github.com/Joe-Degs/riscv/tree/master/projects#resources)
- [elf file format](http://www.skyfree.org/linux/references/ELF_Format.pdf)
- [writing a riscv emulator in rust](https://book.rvemu.app/index.html)
