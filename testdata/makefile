exe=

debug: $(exe)
	@qemu-riscv64 -g 1234 $^ &
	@gdb-multiarch $< -iex "target remote :1234"

clean:
	rm *.o exe
