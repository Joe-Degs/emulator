// RISC-V instruction operation logic - functions that perform the operation
// of the instruction
package main

import (
	"fmt"
)

// Rtype register-register arithmetic operations
func (e *Emulator) decodeRtypeArith(ins uint32) {
	inst := Decode(ins, Rtype{}).(Rtype)
	rs1 := e.Reg(inst.rs1)
	rs2 := e.Reg(inst.rs2)

	switch inst.funct3 | inst.funct7 {
	case 0x0:
		// ADD
		e.SetReg(inst.rd, rs1+rs2)
	case 0x20:
		// SUB
		e.SetReg(inst.rd, rs1-rs2)
	case 0x4:
		// XOR
		e.SetReg(inst.rd, rs1^rs2)
	case 0x6:
		// OR
		e.SetReg(inst.rd, rs1|rs2)
	case 0x7:
		// AND
		e.SetReg(inst.rd, rs1&rs2)
	case 0x1:
		// SLL
		shamt := rs2 & 0b111111
		e.SetReg(inst.rd, rs1<<shamt)
	case 0x5:
		// SRL
		shamt := rs2 & 0b111111
		e.SetReg(inst.rd, rs1>>shamt)
	case 0x5 | 0x20:
		// SRA
		shamt := rs2 & 0b111111
		e.SetReg(inst.rd, uint64(int64(rs1)>>shamt))
	case 0x2:
		// SLT
		if int64(rs1) < int64(rs2) {
			e.SetReg(inst.rd, 1)
		} else {
			e.SetReg(inst.rd, 0)
		}
	case 0x3:
		// SLTU
		if rs1 < rs2 {
			e.SetReg(inst.rd, 1)
		} else {
			e.SetReg(inst.rd, 0)
		}
	}
}

// Rtype 32-bit register-register arithmetic
func (e *Emulator) decodeRtype32RegArith(ins uint32) {
	inst := Decode(ins, Rtype{}).(Rtype)
	rs1 := uint32(e.Reg(inst.rs1))
	rs2 := uint32(e.Reg(inst.rs2))

	switch inst.funct3 | inst.funct7 {
	case 0x0:
		// ADDW
		e.SetReg(inst.rd, uint64(int64(int32(rs1+rs2))))
	case 0x20:
		// SUBW
		e.SetReg(inst.rd, uint64(int64(int32(rs1-rs2))))
	case 0x1:
		// SLLW
		shamt := rs2 & 0b11111
		e.SetReg(inst.rd, uint64(int64(int32(rs1<<shamt))))
	case 0x5:
		// SRLW
		shamt := rs2 & 0b11111
		e.SetReg(inst.rd, uint64(int64(int32(rs1>>shamt))))
	case 0x5 | 0x20:
		// SRAW
		shamt := rs2 & 0b11111
		e.SetReg(inst.rd, uint64(int64(int32(rs1)>>shamt)))
	}
}

// Itype register-immediate arithmetic operations
func (e *Emulator) decodeItypeImmArith(ins uint32) {
	inst := Decode(ins, Itype{}).(Itype)
	rs1 := int64(e.Reg(inst.rs1))
	imm := int64(inst.imm)

	switch inst.funct3 {
	case 0x0:
		// ADDI
		e.SetReg(inst.rd, uint64(rs1+imm))
	case 0x4:
		// XORI
		e.SetReg(inst.rd, uint64(rs1^imm))
	case 0x6:
		// ORI
		e.SetReg(inst.rd, uint64(rs1|imm))
	case 0x7:
		// ANDI
		e.SetReg(inst.rd, uint64(rs1&imm))
	case 0x1:
		// SLLI
		funct7 := (inst.imm >> 6) & 0b111111
		if funct7 == 0x0 {
			shamt := inst.imm & 0b111111
			e.SetReg(inst.rd, uint64(rs1<<shamt))
		} else {
			panic("unreachable slli")
		}
	case 0x5:
		funct7 := (inst.imm >> 6) & 0b111111
		shamt := inst.imm & 0b111111
		if funct7 == 0x0 {
			// SRLI
			e.SetReg(inst.rd, uint64(rs1>>shamt))
		} else if funct7 == 0x10 {
			// SRAI
			e.SetReg(inst.rd, uint64(rs1>>shamt))
		} else {
			panic("unreachable srai")
		}
	case 0x2:
		// SLTI
		if rs1 < imm {
			e.SetReg(inst.rd, 1)
		} else {
			e.SetReg(inst.rd, 0)
		}
	case 0x3:
		// SLTIU
		if uint64(rs1) < uint64(imm) {
			e.SetReg(inst.rd, 1)
		} else {
			e.SetReg(inst.rd, 0)
		}
	default:
		panic(fmt.Errorf("uimplemented Itype with opcode"))
	}
}

// Itype 32-bit arithmetic operations
func (e *Emulator) decodeItype32bitArith(ins uint32) {
	inst := Decode(ins, Itype{}).(Itype)
	rs1 := uint32(e.Reg(inst.rs1))
	imm := uint32(inst.imm)

	switch inst.funct3 {
	case 0x0:
		// ADDIW
		e.SetReg(inst.rd, uint64(int64(int32(rs1+imm))))
	case 0x1:
		// SLLIW
		funct7 := (inst.imm >> 5) & 0b1111111
		if funct7 == 0x0 {
			shamt := inst.imm & 0b11111
			e.SetReg(inst.rd, uint64(int64(int32(rs1<<shamt))))
		} else {
			panic("unreachable slli")
		}
	case 0x5:
		funct7 := (inst.imm >> 5) & 0b1111111
		shamt := inst.imm & 0b11111
		if funct7 == 0x0 {
			// SRLIW
			e.SetReg(inst.rd, uint64(int64(int32(rs1>>shamt))))
		} else if funct7 == 0x20 {
			// SRAIW
			e.SetReg(inst.rd, uint64(int64(int32(rs1)>>shamt)))
		} else {
			panic(fmt.Errorf("itype32load: funct7: %d, shamt: %d\n",
				funct7, shamt))
		}
	}
}

// Itype perform load operations
func (e *Emulator) decodeItypeLoads(ins uint32) error {
	inst := Decode(ins, Itype{}).(Itype)
	addr := VirtAddr(e.Reg(inst.rs1) + uint64(int64(inst.imm)))

	switch inst.funct3 {
	case 0x0:
		// LB
		val, err := ReadIntoVal(e.Mmu, addr, int8(0))
		if err != nil {
			return err
		}
		e.SetReg(inst.rd, uint64(int64(val)))
	case 0x1:
		// LH
		val, err := ReadIntoVal(e.Mmu, addr, int16(0))
		if err != nil {
			return err
		}
		e.SetReg(inst.rd, uint64(int64(val)))
	case 0x2:
		// LW
		val, err := ReadIntoVal(e.Mmu, addr, int32(0))
		if err != nil {
			return err
		}
		e.SetReg(inst.rd, uint64(int64(val)))
	case 0x3:
		// LD
		val, err := ReadIntoVal(e.Mmu, addr, uint64(0))
		if err != nil {
			return err
		}
		e.SetReg(inst.rd, val)
	case 0x4:
		// LBU
		val, err := ReadIntoVal(e.Mmu, addr, uint8(0))
		if err != nil {
			return err
		}
		e.SetReg(inst.rd, uint64(val))
	case 0x5:
		// LHU
		val, err := ReadIntoVal(e.Mmu, addr, uint16(0))
		if err != nil {
			return err
		}
		e.SetReg(inst.rd, uint64(val))
	case 0x6:
		// LWU
		var val uint32
		val, err := ReadIntoVal(e.Mmu, addr, uint32(0))
		if err != nil {
			return err
		}
		e.SetReg(inst.rd, uint64(val))
	}
	return nil
}

// Stype perform store operations
func (e *Emulator) decodeStypeStore(ins uint32) (err error) {
	inst := Decode(ins, Stype{}).(Stype)
	addr := VirtAddr(e.Reg(inst.rs1) + uint64(int64(inst.imm)))
	val := e.Reg(inst.rs2)

	switch inst.funct3 {
	case 0x0:
		// SB
		err = WriteFromVal(e.Mmu, addr, uint8(val&0xff))
		if err != nil {
			return err
		}
	case 0x1:
		// SH
		err = WriteFromVal(e.Mmu, addr, uint16(val&0xffff))
		if err != nil {
			return err
		}
	case 0x2:
		// SW
		err = WriteFromVal(e.Mmu, addr, uint32(val&0xffffffff))
		if err != nil {
			return err
		}
	case 0x3:
		// SD
		err = WriteFromVal(e.Mmu, addr, val)
		if err != nil {
			return err
		}
	}
	return nil
}
