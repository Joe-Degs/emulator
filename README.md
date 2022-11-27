# simpmulator

Hello lads and lasses! 
This project features a [RV64I](https://book.rvemu.app/instruction-set/01-rv64i.html) emulator.
It is a learning project that seeks to emulate a simple linux runtime environment for 
the RISC-V 64bit base instruction set. The project is built to run code compiled
for the risc-v integer architecture using any compiler toolchain of your choosing assuming you
have implemented the syscalls and other runtime dependencies that the toolchain.

It should be easy to extend the project to other use cases that require an emulator, and this
is something I hope to explore in the near future. Right now I'm still learning
and trying to get the simple use case of executing programs right.

If you want to hack on this project or on something similar to this, you can hit me up. I would
like any criticism and help I can get. Even better if you wanna join me on the journey of
learning how computers work.

### Building and running emulator
To build the project, clone this repository and run `make build`
After that you can take the built emulator for a spin by running `make` or `make run`, which
runs a test hello world program.

The emulator supports commandline options for tweaking the runtime environment it provides
to the programs that it executes. To see those, run the emulator without any arguments.
You can find other sample test programs in the `testdata` directory.

Executing a helloworld program is as simple as doing the following:
```
joe@debian:~/dev/emulator$ ./simpmulator testdata/musl/hello/hello
Hello World
```

The rest of the sections below contain guides on how to build your own `rv64i` program to run
against the emulator. And how to extend the emulator if you want to.

### Testing with the musl_libc toolchain 
The [musl](https://www.musl-libc.org/intro.html) toolchain for building test 
applications for this project can be installed using instructions from 
[here](https://github.com/Joe-Degs/riscv). 
After installing and building the toolchain checkout this 
[project](https://github.com/Joe-Degs/riscv/tree/master/projects), it contains 
instructions on how to generate makefiles for generating binaries that can be 
executed in the emulator.

The `testdata/musl/` directory contains sample source files, makefiles and binaries
you can try out before going out to build your own.

### Testing with newlib toolchain
The [newlib](https://en.wikipedia.org/wiki/Newlib) toolchain used in this project
can be installed by doing as follows:
A precompiled `riscv64-gcc-unknown-elf` with `newlib` toolchain for embedded systems 
can be found [here](https://random-oracles.org/risc-v-gcc-toolchain/). Download the
the tar archive, unarchive it into a `/opt/risc-newlib-toolchain` directory and
have fun hacking.

Took me some doing to get it to work because I'm new to the stuff.
If you don't know what you are doing just like me and you want to use the toolchain to build your
own binaries, copy any of  the makefiles from the 
[newlibc](https://github.com/Joe-Degs/emulator/tree/master/testdata/newlibc) `testdata/newlibc`
directory, tweak it if need be: *if you can't, fuck around a lil bit, you'll figure it out!*. 
Or hit me up, with our collective energies we might be able to *fuck around faster and figure it out*

If you succesfully compile a program with the toolchain, execute the program
with the emulator to see if it works. If you encounter any hiccups try to figure
it out or reach out to me let's figure it out together.

Everything else should work fine (LOL!), the only difference is if you are using
`musl_libc` you have more syscalls to implement and the many problems that come
with doing it. It's _LINUX_ baby!

It doesn't really matter the C library you use, so long as you are willing, and
you have time to make it work, have a go at it.

### Next Steps
I am able to run simple programs in the emulator (atleast I was at some point). 
You will hit a panic when you encounter a syscall that is not yet implemented,
which contains info on the syscall
that is causing the error and the contents of its argument registers.
```
PATH: /home/joe/go/src/github.com/Joe-Degs/emulator/testdata/hello/hello
FILENAME: hello
CURRENT ALLOCATION: 0x140b0
STACK [0x240b0 -> 0x140b0]
panic: Syscall {num: 64, a0: 1, a1: 77416, a2: 12, a3: 0, a4: 7, a5: 70816}


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
Implementing the syscalls should be trivial (LOL!), but I'll be doing it
incrementally, i.e implementing syscalls only when I have need of them.
Syscall names and their respective numbers can be found in `*syscalls*.txt`
So implement a syscall, stick in the syscall table `syscalls` in the `syscall.go`
and you might be good to go.

### features
- Memory mapping unit for mapping programs into memory
- Memory permissions to ensure secured access.
- Ability to reset/clone/fork the execution context provided by the emulator.
- Ability to dump execution context for easy debugging of issues.
```
joe@debian:~/dev/emulator$ ./simpmulator -v -elf-info testdata/newlibc/newtool/test
PATH: /home/joe/go/src/github.com/Joe-Degs/emulator/testdata/newtool/test
FILENAME: test
MEM SIZE: 0x1fffff
STACK [0x1fff00 -> 0x1fef00]
HEAP [0x17990 -> 0x18990]
CURRENT ALLOCATION: 0x18990
EmuExit {
    zero:  0000000000000000  ra: 0000000000012304  sp:  00000000001ffdcc   gp:  0000000000015188
    tp:    0000000000000000  t0: 00000000000102d4  t1:  000000000000000f   t2:  0000000000000000
    s0/fp: 0000000000014968  s1: 0000000000000020  a0:  0000000000000020   a1:  0000000000013920
    a2:    0000000000000020  a3: 0000000000000000  a4:  0000000000000000   a5:  0000000000000000
    a6:    0000000000018dc0  a7: 0000000000000040  s2:  00000000001ffe9c   s3:  0000000000013920
    s4:    00000000001ffe74  s5: 0000000000014968  s6:  000000007ffffc00   s7:  0000000000000000
    s8:    0000000000000000  s9: 0000000000000000  s10: 0000000000000000   s11: 0000000000000000
    t3:    0000000000000000  t4: 0000000000000000  t5:  0000000000000000   t6:  0000000000000000
    pc:    0000000000013750
    Syscall {num: 64, a0: 1, a1: 80160, a2: 32, a3: 0, a4: 0, a5: 0},
    opcode: 0b01110011
}
```

### resources
This project is inspired by [gamozolabs'](https://github.com/gamozolabs) [Fuzz
Week 2020](https://gamozolabs.github.io/2020/07/12/fuzz_week_2020.html).
Check out his [youtube](https://youtube.com/user/gamozolabs).

- [risc-v learning resources](https://github.com/Joe-Degs/riscv/tree/master/projects#resources)
- [elf file format](http://www.skyfree.org/linux/references/ELF_Format.pdf)
- [writing a riscv emulator in rust](https://book.rvemu.app/index.html)
