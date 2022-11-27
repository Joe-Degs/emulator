toolchain = /opt/riscv-newlib-toolchain
tool = riscv64-unknown-elf
cc = $(toolchain)/bin/$(tool)-gcc
incpath = $(toolchain)/$(tool)/include
rv64i_newlib = $(toolchain)/$(tool)/lib/rv64i/lp64
cflags = -Wall -Werror -static -g -march=rv64i -mabi=lp64
cflags += -I$(incpath)
# cflags += -fuse-ld=lld
cflags += --specs=$(rv64i_newlib)/sim.specs

all:
	$(cc) $(cflags) -o test \
		-L$(rv64i_newlib)/crt0.o \
		-L$(rv64i_newlib)/libc.a \
		-L$(rv64i_newlib)/libgloss.a \
		-L$(rv64i_newlib)/libg.a \
		-L$(rv64i_newlib)/libm.a \
		-L$(rv64i_newlib)/libsim.a \
		*.c

disas-gcc:
	riscv64-unknown-elf-objdump -D -j .text test 2>&1 | vim -

disas-llvm:
	llvm-objdump-11 -D -j .text test 2>&1 | vim -

clean:
	rm test
