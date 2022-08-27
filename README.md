## simpmulator

A simple [RV64I](https://book.rvemu.app/instruction-set/01-rv64i.html) RISC-V emulator.

### Building and Testing
The toolchain for building test applications for this project can be installed
using instructions from [here](https://github.com/Joe-Degs/riscv). After installing
and building the toolchain use [this](https://github.com/Joe-Degs/riscv/tree/master/projects),
it contains instructions on how to generate makefiles for compiling test apps to run
against the emulator.

Any RV64I (base instruction set) toolchain should work too (i have not tried any
myself). But the compiled programs to run against the emulator must be
statically linked and contain only the 64 bit base instruction set.

The `testdata/` directory contains sample compiled binaries that can be used to
test the emulator.

The emulator is a go program so you could just build it or run it using the script
`run.sh`
```
go run . <path/to/test-binary>
```

### Next Steps
For now we are able to run simple programs in the emulator, but you will hit a panic when
you encounter a syscall that is not yet implemented. The panic contains info on the syscall
that is causing the error and the contents of its argument registers.
```
PATH: /home/joe/go/src/github.com/Joe-Degs/emulator/testdata/hello/hello
FILENAME: hello
CURRENT ALLOCATION: 0x140b0
STACK [0x240b0 -> 0x140b0]
panic: Syscall{num: 64, a0: 1, a1: 77416, a2: 12, a3: 0, a4: 7, a5: 70816, a6: 0}


goroutine 1 [running]:
main.main()
        /home/joe/go/src/github.com/Joe-Degs/emulator/main.go:62 +0x5c5
exit status 2
```
Implementing the syscalls should be trivial and a learning experience, but I have
quite a lot on my plate right now. So I'll do that another time.
Syscall names and their respective numbers can be found in `syscall-numbers.txt`

### resources
This project is inspired by [gamozolabs'](https://github.com/gamozolabs) [Fuzz
Week 2020](https://gamozolabs.github.io/2020/07/12/fuzz_week_2020.html).
Check out his [youtube](https://youtube.com/user/gamozolabs).

- [risc-v learning resources](https://github.com/Joe-Degs/riscv/tree/master/projects#resources)
- [elf file format](http://www.skyfree.org/linux/references/ELF_Format.pdf)
- [writing a riscv emulator in rust](https://book.rvemu.app/index.html)
