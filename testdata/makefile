# use this makefile to start the debug if you have the debuggers
# run "make file=<path/to/binary>"
# It depends on qemu-riscv64 and gdb-multiarch to run
debug:
	@qemu-riscv64 -g 1234 $(target) &
	@gdb-multiarch $(target) -iex "target remote :1234"
