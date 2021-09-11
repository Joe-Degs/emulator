// This file contains implementation of the risc-v instruction set
package main

//go:generate stringer -type=Register

// Register represents a single riscv register file
type Register uint8

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

func GetReg(reg uint32) Register {
	if reg > 31 {
		return Zero
	}
	return Register(uint8(reg))
}

type Instruction interface {
	Decode(inst uint32) Instruction
}

// Rtype instructions represent register to register computations
type Rtype struct {
	rd     Register
	funct3 uint32
	rs1    Register
	rs2    Register
	funct7 uint32
}

func (Rtype) Decode(inst uint32) Instruction {
	return Rtype{
		rd:     GetReg((inst >> 7) & 0b11111),
		funct3: (inst >> 12) & 0b111,
		rs1:    GetReg((inst >> 15) & 0b11111),
		rs2:    GetReg((inst >> 20) & 0b11111),
		funct7: (inst >> 25) & 0b1111111,
	}
}

// Itype for loads and short immediate operations
type Itype struct {
	rd     Register
	funct3 uint32
	rs1    Register
	imm    int32
}

func (Itype) Decode(inst uint32) Instruction {
	return Itype{
		rd:     GetReg((inst >> 7) & 0b11111),
		funct3: (inst >> 12) & 0b111,
		rs1:    GetReg((inst >> 15) & 0b11111),
		imm:    int32(inst) >> 20,
	}
}

// Stype for stores
type Stype struct {
	funct3 uint32
	rs1    Register
	rs2    Register
	imm    int32
}

func (Stype) Decode(inst uint32) Instruction {
	imm115 := (inst >> 25) & 0b1111111
	imm40 := (inst >> 7) & 0b11111
	imm := (imm115 << 5) | imm40
	// sign extend imm
	simm := (int32(imm) << 20) >> 20
	return Stype{
		funct3: (inst >> 12) & 0b111,
		rs1:    GetReg((inst >> 15) & 0b11111),
		rs2:    GetReg((inst >> 20) & 0b11111),
		imm:    simm,
	}
}

// Btype for conditional branch operation
type Btype struct {
	imm    int32
	funct3 uint32
	rs1    Register
	rs2    Register
}

func (Btype) Decode(inst uint32) Instruction {
	imm12 := (inst >> 31) & 0b1
	imm105 := (inst >> 25) & 0b111111
	imm41 := (inst >> 8) & 0b1111
	imm11 := (inst >> 7) & 0b1
	// pieceing them all together
	imm := (imm12 << 12) | (imm11 << 11) | (imm105 << 5) | (imm41 << 1)
	simm := (int32(imm) << 19) >> 19
	return Btype{
		rs1:    GetReg((inst >> 15) & 0b1111111),
		rs2:    GetReg((inst >> 20) & 0b1111111),
		funct3: (inst >> 12) & 0b111,
		imm:    simm,
	}
}

// Utype for long immediate operations
type Utype struct {
	rd  Register
	imm int32
}

func (Utype) Decode(inst uint32) Instruction {
	return Utype{
		rd:  GetReg((inst >> 7) & 0b11111),
		imm: int32(uint32((inst >> 12) & 0b11111111111111111111)),
	}
}

// Jtype for unconditional jump operations
type Jtype struct {
	rd  Register
	imm int32
}

func (Jtype) Decode(inst uint32) Instruction {
	imm20 := (inst >> 31) & 0b1
	imm101 := (inst >> 21) & 0b1111111111
	imm11 := (inst >> 20) & 0b1
	imm1912 := (inst >> 12) & 0b11111111

	// shift bits to their position
	imm := (imm20 << 20) | (imm1912 << 12) | (imm11 << 11) | (imm101 << 1)

	// sign extend immediate
	simm := (int32(imm) << 11) >> 11
	return Jtype{
		rd:  GetReg((inst >> 7) & 0b11111),
		imm: simm,
	}
}

// this will switch between the type here and return
func Decode(inst uint32, instruction Instruction) Instruction {
	return instruction.Decode(inst)
}
