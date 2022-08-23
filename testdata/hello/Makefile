CFLAGS = -g -target riscv64 -march=rv64i -I$(MUSL_INSTALL_PATH)/include

all: hello

hello: hello.o
	ld.lld-$(CLANG_VERSION) -o hello \
		$(MUSL_INSTALL_PATH)/lib/crt1.o \
		$(MUSL_INSTALL_PATH)/lib/libc.a \
		$(COMPILER_RT_PATH) \
		*.o


hello.o: hello.S
	clang-$(CLANG_VERSION) $(CFLAGS) -c hello.S

llvm:
	llvm-objdump-11 -D -j .text hello | vim -

gcc:
	riscv64-linux-gnu-objdump -D -j .text hello | vim -

clean:
	rm -f hello *.o
