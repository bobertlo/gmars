package gmars

import (
	"fmt"
	"strconv"
	"strings"
)

type Address uint64
type OpCode uint8
type OpMode uint8
type AddressMode uint8

const (
	DAT OpCode = iota
	MOV
	ADD
	SUB
	MUL
	DIV
	MOD
	CMP
	SEQ
	SNE
	SLT
	JMP
	JMZ
	JMN
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
	case DIV:
		return "DIV"
	case MOD:
		return "MOD"
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
	case NOP:
		return "NOP"
	default:
		return "???"
	}
}

func getOpCode(op string) (OpCode, error) {
	switch strings.ToLower(op) {
	case "dat":
		return DAT, nil
	case "mov":
		return MOV, nil
	case "add":
		return ADD, nil
	case "sub":
		return SUB, nil
	case "mul":
		return MUL, nil
	case "div":
		return DIV, nil
	case "mod":
		return MOD, nil
	case "jmp":
		return JMP, nil
	case "jmz":
		return JMZ, nil
	case "jmn":
		return JMN, nil
	case "djn":
		return DJN, nil
	case "cmp":
		return CMP, nil
	case "seq":
		return SEQ, nil
	case "slt":
		return SLT, nil
	case "sne":
		return SNE, nil
	case "spl":
		return SPL, nil
	case "nop":
		return NOP, nil
	default:
		return 0, fmt.Errorf("invalid opcode '%s'", op)
	}
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

// String returns the string representation of an OpMode, or "?"
func (m OpMode) String() string {
	switch m {
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
	default:
		return "?"
	}
}

func getOpMode(opModeStr string) (OpMode, error) {
	switch strings.ToLower(opModeStr) {
	case "a":
		return A, nil
	case "b":
		return B, nil
	case "ab":
		return AB, nil
	case "ba":
		return BA, nil
	case "i":
		return I, nil
	case "f":
		return F, nil
	case "x":
		return X, nil
	default:
		return 0, fmt.Errorf("invalid op mode: '%s'", opModeStr)
	}
}

func getOp94(op string) (OpCode, OpMode, error) {
	fields := strings.Split(op, ".")
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("invalid op: '%s'", op)
	}

	code, err := getOpCode(fields[0])
	if err != nil {
		return 0, 0, err
	}

	opmode, err := getOpMode(fields[1])
	if err != nil {
		return 0, 0, err
	}

	return code, opmode, nil
}

func getOpCode88(op string) (OpCode, error) {
	switch strings.ToLower(op) {
	case "dat":
		return DAT, nil
	case "mov":
		return MOV, nil
	case "add":
		return ADD, nil
	case "sub":
		return SUB, nil
	case "jmp":
		return JMP, nil
	case "jmz":
		return JMZ, nil
	case "jmn":
		return JMN, nil
	case "djn":
		return DJN, nil
	case "cmp":
		return CMP, nil
	case "slt":
		return SLT, nil
	case "spl":
		return SPL, nil
	default:
		return 0, fmt.Errorf("invalid opcode '%s'", op)
	}
}

const (
	DIRECT      AddressMode = iota // "$" direct reference to another address
	IMMEDIATE                      // "#" use the immediate value of this instruction
	A_INDIRECT                     // "*" use the A-Field of the address referenced by a pointer
	B_INDIRECT                     // "@" use the B-Field of the address referenced by a pointer
	A_DECREMENT                    // "{" use the A-field of the address referenced by a pointer, after decrementing
	B_DECREMENT                    // "<" use the B-field of the address referenced by a pointer, after decrementing
	A_INCREMENT                    // "}" use the A-field of the address referenced by a pointer, before incrementing
	B_INCREMENT                    // ">" use the B-field of the address referenced by a pointer, before incrementing
)

// String returns the single character string representation of an AddressMode, or "?"
func (m AddressMode) String() string {
	switch m {
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
	default:
		return "?"
	}
}

func getAddressMode(modeStr string) (AddressMode, error) {
	switch modeStr {
	case "#":
		return IMMEDIATE, nil
	case "$":
		return DIRECT, nil
	case "*":
		return A_INDIRECT, nil
	case "@":
		return B_INDIRECT, nil
	case "{":
		return A_DECREMENT, nil
	case "<":
		return B_DECREMENT, nil
	case "}":
		return A_INCREMENT, nil
	case ">":
		return B_INCREMENT, nil
	default:
		return 0, fmt.Errorf("invalid address mode: '%s'", modeStr)
	}
}

func getAddressMode88(modeStr string) (AddressMode, error) {
	switch modeStr {
	case "#":
		return IMMEDIATE, nil
	case "$":
		return DIRECT, nil
	case "@":
		return B_INDIRECT, nil
	case "<":
		return B_DECREMENT, nil
	default:
		return 0, fmt.Errorf("invalid 88 address mode: '%s'", modeStr)
	}
}

func parseAddress(input string, coresize Address) (Address, error) {
	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, err
	}

	m := int64(coresize)
	val = val % m
	if val < 0 {
		val = (m + val) % m
	}

	return Address(val), nil
}

func signedAddress(a, coresize Address) int {
	if a > (coresize / 2) {
		return -(int(coresize) - int(a))
	}
	return int(a)
}

// Instruction represents the raw values of a memory address
type Instruction struct {
	Op     OpCode
	OpMode OpMode
	A      Address
	AMode  AddressMode
	B      Address
	BMode  AddressMode
}

// String returns the decompiled instruction with unsigned field values
func (i Instruction) String() string {
	return fmt.Sprintf("%s.%-2s %s %5d %s %5d", i.Op, i.OpMode, i.AMode, i.A, i.BMode, i.B)
}

// NormString returns the decompiled instruction with signed field values normalized to a core size.
func (i Instruction) NormString(coresize Address) string {
	anorm := signedAddress(i.A, coresize)
	bnorm := signedAddress(i.B, coresize)
	return fmt.Sprintf("%s.%-2s %s %5d %s %5d", i.Op, i.OpMode, i.AMode, anorm, i.BMode, bnorm)
}
