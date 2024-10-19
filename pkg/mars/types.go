package mars

type WarriorIndex uint32
type Address uint32
type OpCode uint8
type OpMode uint8
type AddressMode uint8

type Instruction struct {
	Op     OpCode
	OpMode OpMode
	A      Address
	AMode  AddressMode
	B      Address
	BMode  AddressMode
}

const (
	DAT OpCode = iota
	MOV
	ADD
	SUB
	JMP
	JMZ
	JMN
	CMP
	SLT
	DJN
	SPL
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
	case JMP:
		return "JMP"
	}
	return "___"
}

const (
	A OpMode = iota
	B
	AB
	BA
	F
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
	IMMEDIATE AddressMode = iota
	DIRECT
	B_INDIRECT
	B_DECREMENT
)

func (am AddressMode) String() string {
	switch am {
	case IMMEDIATE:
		return "#"
	case DIRECT:
		return "$"
	case B_INDIRECT:
		return "@"
	case B_DECREMENT:
		return "<"
	}
	return "_"
}
