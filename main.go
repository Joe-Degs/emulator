package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:generate stringer -type=Register,Perm,MemErrType -output=string.go

var (
	VERBOSE             bool
	VERBOSE_PC_OPCODE   bool
	VERBOSE_INST_DECODE bool
	VERBOSE_SYSCALL     bool
	LOG_STATE           bool
	DUMP_ELF_INFO       bool
	MEM_SIZE            uint // = 2 * 1024 * 1024
)

func init() {
	flag.BoolVar(&VERBOSE, "v", false, "verbose output: emulator runtime information")
	flag.BoolVar(&VERBOSE_PC_OPCODE, "verbose-pc", false, "verbose output; PC")
	flag.BoolVar(&VERBOSE_INST_DECODE, "verbose-inst", false, "verbose output: instruction decode info")
	flag.BoolVar(&VERBOSE_SYSCALL, "verbose-syscall", false, "verbose output: system call information")
	flag.BoolVar(&LOG_STATE, "dump-state", false, "dump state of emulator when the inferior program encounters error")
	flag.BoolVar(&DUMP_ELF_INFO, "elf-info", false, "dump loaded elf binary info")
	flag.UintVar(&MEM_SIZE, "memsize", 1024*1024, "specify the memory size")
}

func exitf(pattern string, args ...any) {
	if !strings.HasSuffix(pattern, "\n") {
		pattern = pattern + "\n"
	}
	fmt.Fprintf(os.Stderr, pattern, args...)
	flag.Usage()
	os.Exit(1)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		exitf("%s [OPTIONS] <path/to/binary> [PROG ARGS]", os.Args[0])
	}
	var (
		path string
		err  error
	)
	if path, err = filepath.Abs(args[0]); err != nil {
		exitf("%v", err)
	}

	emu := NewEmulator(MEM_SIZE)
	if err := emu.MapProgram(path, args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	if VERBOSE {
		fmt.Println("")
		fmt.Printf("PATH: %s\nFILENAME: %s\n", emu.program.path, emu.program.name)
		fmt.Printf("MEM SIZE: %#x\n", MEM_SIZE-1)
		fmt.Printf("STACK [%#x -> %#x]\n", emu.Stack(), emu.Stack()-STACK_SIZE)
		fmt.Printf("HEAP [%#x -> %#x]\n", emu.Heap(), emu.Heap()+HEAP_SIZE)
		fmt.Printf("CURRENT ALLOCATION: %#x\n", emu.Mmu.curAlloc)
		fmt.Println("")
	}

	defer func() {
		if r := recover(); r != nil {
			exitf(emu.String())
		}
	}()

	if err := emu.Run(); err != nil {
		handleErrors(emu, err)
	}
}

// handle emulator execution errors
func handleErrors(emu *Emulator, err error) {
	if e, ok := err.(EmuExit); ok {
		switch t := e.cause.(type) {
		case MMUError:
			if LOG_STATE {
				emu.Inspect(t.addr, t.size)
				emu.InspectPerms(t.addr, t.size)
				exitf("%s", e.Error())
			}
			return
		case Done:
			os.Exit(t.status)
		}
		return
	}
	exitf("%v", err)
}
