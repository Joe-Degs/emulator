CFLAGS = -g -target riscv64 -march=rv64i -I$(MUSL_INSTALL_PATH)/include

all: printy

printy: printy.o
	ld.lld-$(CLANG_VERSION) -o printy \
		$(MUSL_INSTALL_PATH)/lib/crt1.o \
		$(MUSL_INSTALL_PATH)/lib/libc.a \
		$(COMPILER_RT_PATH) \
		*.o

printy.o: printy.c
	clang-$(CLANG_VERSION) $(CFLAGS) -c printy.c

clean:
	rm -f printy *.o
