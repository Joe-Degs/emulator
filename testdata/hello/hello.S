.globl main

.data
	msg: .string "Hello World\n"

.equ: msg_len, 12
.equ STDOUT, 1
.equ WRITE_SYSCALL_NO, 64

.text
main:
	li a0, STDOUT
	la a1, msg
	li a2, msg_len
	li a7, WRITE_SYSCALL_NO
	ecall
	mv a0, zero
