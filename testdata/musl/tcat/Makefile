CFLAGS = -g -target riscv64 -march=rv64i -I$(MUSL_INSTALL_PATH)/include

all: tcat

tcat: tcat.o
	ld.lld-$(CLANG_VERSION) -o tcat \
		$(MUSL_INSTALL_PATH)/lib/crt1.o \
		$(MUSL_INSTALL_PATH)/lib/libc.a \
		$(COMPILER_RT_PATH) \
		*.o

tcat.o: tcat.c
	clang-$(CLANG_VERSION) $(CFLAGS) -c tcat.c

clean:
	rm -f tcat *.o
