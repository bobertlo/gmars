package mars

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
	case "mul":
		return MUL, nil
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

func getOpModeAndValidate88(Op OpCode, AMode AddressMode, BMode AddressMode) (OpMode, error) {
	switch Op {
	case DAT:
		// DAT:
		// F always, any combination of # and < allowed
		if AMode != IMMEDIATE && AMode != B_DECREMENT {
			return 0, fmt.Errorf("invalid a mode '%s' for op 'dat'", AMode)
		}
		if BMode != IMMEDIATE && BMode != B_DECREMENT {
			return 0, fmt.Errorf("invalid b mode '%s' for op 'dat'", BMode)
		}
		return F, nil

	case CMP:
		fallthrough
	case MOV:
		// CMP and MOV:
		// AB IF #A, I otherwise, no #B allowed
		if BMode == IMMEDIATE {
			return 0, fmt.Errorf("invalid b mode '#' for op '%s'", Op)
		}
		if AMode == IMMEDIATE {
			return AB, nil
		} else {
			return I, nil
		}

	case SLT:
		// SLT;
		// AB if #A, B otherwise, no #B allowed
		if BMode == IMMEDIATE {
			return 0, fmt.Errorf("invalid b mode '#' for op 'slt'")
		}
	case ADD:
		fallthrough
	case SUB:
		// ADD and SUB:
		//
		if BMode == IMMEDIATE {
			return 0, fmt.Errorf("invalid b mode '#' for op '%s'", Op)
		}
		if AMode == IMMEDIATE {
			return AB, nil
		} else {
			return F, nil
		}
	case JMP:
		fallthrough
	case JMN:
		fallthrough
	case JMZ:
		fallthrough
	case DJN:
		fallthrough
	case SPL:
		if AMode == IMMEDIATE {
			return 0, fmt.Errorf("invalid a mode '#' for op '%s'", Op)
		}
		return B, nil
	}
	return B, fmt.Errorf("unknown op code: '%s'", Op)
}

func parse88LoadFile(reader io.Reader, coresize Address) (WarriorData, error) {
	data := WarriorData{
		Name:     "Unknown",
		Author:   "Anonymous",
		Strategy: "",
		Code:     make([]Instruction, 0),
		Start:    0,
	}

	lineNum := 0
	breader := bufio.NewReader(reader)
	for {
		raw_line, err := breader.ReadString('\n')
		if err != nil {
			break
		}
		lineNum++

		if len(raw_line) == 0 {
			continue
		}

		lower := strings.ToLower(raw_line)

		// handle metadata comments
		if raw_line[0] == ';' {
			if strings.HasPrefix(lower, ";name") {
				data.Name = strings.TrimSpace(raw_line[5:])
			} else if strings.HasPrefix(lower, ";author") {
				data.Author = strings.TrimSpace(raw_line[7:])
			} else if strings.HasPrefix(lower, ";strategy") {
				data.Strategy += raw_line[10:]
			}
			continue
		}

		// trim comments
		if strings.Contains(lower, ";") {
			lower = strings.Split(lower, ";")[0]
		}

		// remove commas, since whitespace is required
		lower = strings.ReplaceAll(lower, ",", " ")

		// split into fields based on whitespace
		fields := strings.Fields(lower)

		// valid instructions need exactly 5 fields
		// only other option is "END" pseudo opcode with 0 or 1 arguments
		if len(fields) != 5 {
			// empty line
			if len(fields) == 0 {
				continue
			}

			if fields[0] != "end" {
				return WarriorData{}, fmt.Errorf("line %d: invalid op-code '%s'", lineNum, fields[0])
			} else if len(fields) > 2 {
				return WarriorData{}, fmt.Errorf("line %d: too many arguments to 'end'", lineNum)
			}

			// no arguments
			if len(fields) == 1 {
				break
			}

			val, err := strconv.ParseInt(fields[1], 10, 32)
			if err != nil {
				return WarriorData{}, fmt.Errorf("line %d: error parsing integer: %s", lineNum, err)
			}
			if val < 0 || val > int64(len(data.Code)) {
				return WarriorData{}, fmt.Errorf("line %d: start address outside warrior code", lineNum)
			}

			data.Start = int(val)
			break
		}

		// attempt to parse the 5 fields as an instruction and append to code
		op, err := getOpCode88(fields[0])
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", lineNum, err)
		}

		amode, err := getAddressMode88(fields[1])
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", lineNum, err)
		}
		aval, err := parseAddress(fields[2], coresize)
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: error parsing a field integer: %s", lineNum, err)
		}

		bmode, err := getAddressMode88(fields[3])
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", lineNum, err)
		}
		bval, err := parseAddress(fields[4], coresize)
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: error parsing b field integer: %s", lineNum, err)
		}

		opmode, err := getOpModeAndValidate88(op, amode, bmode)
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", lineNum, err)
		}

		data.Code = append(data.Code, Instruction{
			Op:     op,
			OpMode: opmode,
			AMode:  amode,
			A:      aval,
			BMode:  bmode,
			B:      bval,
		})

	}
	return data, nil
}

func ParseLoadFile(reader io.Reader, simConfig SimulatorConfig) (WarriorData, error) {
	if simConfig.Mode != ICWS88 {
		return WarriorData{}, fmt.Errorf("not implemented")
	}
	return parse88LoadFile(reader, simConfig.CoreSize)
}
