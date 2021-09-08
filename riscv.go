// This file contains implementation of the risc-v instruction set
package main

// Register represents a single riscv register file
type Register uint

// variants of the risc-v register
const (
	Zero Register = iota
	Ra
	Sp
	Gp
	Tp
	T0
	T1
	T2
	S0 // also FP
	S1
	A1
	A2
	A3
	A4
	A5
	A6
	A7
	S2
	S3
	S4
	S5
	S6
	S7
	S8
	S9
	S10
	S11
	T3
	T4
	T5
	T6
	Pc
)

type Instruction interface {
	Decode(inst uint) Instruction
}

// Rtype instructions represent register to register computations
type Rtype struct {
	opcode uint
	rd     Register
	funct3 uint
	rs1    Register
	rs2    Register
	funct7 uint
}

func (Rtype) Decode(inst uint) Instruction {
	return Rtype{}
}

// Itype represents register - immediate instructions
type Itype struct {
	opcode uint
	rd     Register
	funct3 uint
	rs1    Register
	imm    int
}

func (Itype) Decode(inst uint) Instruction {
	return Itype{}
}

// Stype represents store instructions
type Stype struct {
	opcode uint
	funct3 uint
	rs1    Register
	rs2    Register
	imm    int
}

func (Stype) Decode(inst uint) Instruction {
	return Stype{}
}

// Btype represents all conditional branch instructions.
type Btype struct {
	opcode uint
	rd     Register
	imm    int
	funct3 uint
	rs1    Register
	rs2    Register
}

func (Btype) Decode(inst uint) Instruction {
	return Btype{}
}

// Utype represents all upper immediate instructions
type Utype struct {
	opcode uint
	rd     Register
	imm    int
}

func (Utype) Decode(inst uint) Instruction {
	return Utype{}
}

// Jtype represents all unconditional jump instructions
type Jtype struct {
	opcode uint
	rd     Register
	imm    int
}

func (Jtype) Decode(inst uint) Instruction {
	return Jtype{}
}

// this will switch between the type here and return
func Decode(inst uint, instruction Instruction) Instruction {
	switch instruction.(type) {
	case Jtype:
		return instruction.(Jtype).Decode(inst)
	case Btype:
		return instruction.(Btype).Decode(inst)
	case Itype:
		return instruction.(Itype).Decode(inst)
	case Utype:
		return instruction.(Utype).Decode(inst)
	case Stype:
		return instruction.(Stype).Decode(inst)
	case Rtype:
		return instruction.(Rtype).Decode(inst)
	default:
		panic("instruction type not implemented")
	}
}
