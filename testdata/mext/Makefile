CC = riscv64-unknown-elf-gcc
CFLAGS = -Wall -static -g -march=rv64im -mabi=lp64 -nostdlib

all: test

test: test.c
	$(CC) $(CFLAGS) -o $@ $<

disas:
	riscv64-unknown-elf-objdump -d test

disas-full:
	riscv64-unknown-elf-objdump -D test

disas-vim:
	riscv64-unknown-elf-objdump -d test 2>&1 | vim -

clean:
	rm -f test
