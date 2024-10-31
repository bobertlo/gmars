package mars

import "fmt"

type WarriorIndex uint32
type Address uint32
type OpCode uint8
type OpMode uint8
type AddressMode uint8

type SimulatorMode uint8

const (
	ICWS88 SimulatorMode = iota
	NOP94
	ICWS94
)

type Instruction struct {
	Op     OpCode
	OpMode OpMode
	A      Address
	AMode  AddressMode
	B      Address
	BMode  AddressMode
}

func (i Instruction) String() string {
	return fmt.Sprintf("%s.%-2s %s %5d %s %5d", i.Op, i.OpMode, i.AMode, i.A, i.BMode, i.B)
}

func signedAddress(a, coresize Address) int {
	if a > (coresize / 2) {
		return -(int(coresize) - int(a))
	}
	return int(a)
}

func (i Instruction) NormString(coresize Address) string {
	anorm := signedAddress(i.A, coresize)
	bnorm := signedAddress(i.A, coresize)
	return fmt.Sprintf("%s.%-2s %s %5d %s %5d", i.Op, i.OpMode, i.AMode, anorm, i.BMode, bnorm)
}

const (
	DAT OpCode = iota
	MOV
	ADD
	SUB
	MUL
	DIV
	MOD
	JMP
	JMZ
	JMN
	CMP
	SEQ
	SLT
	SNE
	DJN
	SPL
	NOP
)

func (o OpCode) String() string {
	switch o {
	case DAT:
		return "DAT"
	case MOV:
		return "MOV"
	case ADD:
		return "ADD"
	case SUB:
		return "SUB"
	case MUL:
		return "MUL"
	case CMP:
		return "CMP"
	case SEQ:
		return "SEQ"
	case SNE:
		return "SNE"
	case SLT:
		return "SLT"
	case JMP:
		return "JMP"
	case JMN:
		return "JMN"
	case JMZ:
		return "JMZ"
	case DJN:
		return "DJN"
	case SPL:
		return "SPL"
	}
	return "___"
}

const (
	F OpMode = iota
	A
	B
	AB
	BA
	X
	I
)

func (om OpMode) String() string {
	switch om {
	case A:
		return "A"
	case B:
		return "B"
	case AB:
		return "AB"
	case BA:
		return "BA"
	case F:
		return "F"
	case X:
		return "X"
	case I:
		return "I"
	}
	return "_"
}

const (
	DIRECT AddressMode = iota
	IMMEDIATE
	A_INDIRECT
	B_INDIRECT
	A_DECREMENT
	B_DECREMENT
	A_INCREMENT
	B_INCREMENT
)

func (am AddressMode) String() string {
	switch am {
	case IMMEDIATE:
		return "#"
	case DIRECT:
		return "$"
	case A_INDIRECT:
		return "*"
	case B_INDIRECT:
		return "@"
	case A_DECREMENT:
		return "{"
	case B_DECREMENT:
		return "<"
	case A_INCREMENT:
		return "}"
	case B_INCREMENT:
		return ">"
	}
	return "_"
}
