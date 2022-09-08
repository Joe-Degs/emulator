path = /opt/riscv-newlib-toolchain
tool = riscv64-unknown-elf
cc = $(path)/bin/$(tool)-gcc
incpath = $(path)/$(tool)/include
rv64i_newlib = $(path)/$(tool)/lib/rv64i/lp64
cflags = -Wall -Werror -static -g -march=rv64i -mabi=lp64
cflags += -I$(incpath)
# cflags += -fuse-ld=lld
cflags += --specs=$(rv64i_newlib)/nosys.specs

all:
	$(cc) $(cflags) -o test \
		-L$(rv64i_newlib)/crt0.o \
		-L$(rv64i_newlib)/libc.a \
		-L$(rv64i_newlib)/libgloss.a \
		-L$(rv64i_newlib)/libg.a \
		-L$(rv64i_newlib)/libm.a \
		-L$(rv64i_newlib)/libnosys.a \
		main.c

clean:
	rm test
