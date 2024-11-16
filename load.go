package gmars

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func parseLoadFile94(reader io.Reader, coresize Address) (WarriorData, error) {
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
		// empty lines and last lines without newlines seem to be missed
		// should something else be used? or are these not worth handling?
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

		// remove comma before counting fields
		nocomma := strings.ReplaceAll(lower, ",", " ")

		// split into fields based on whitespace
		fields := strings.Fields(nocomma)

		// valid instructions need exactly 5 fields
		// only other option is "ORG" pseudo opcode with exactly 1 arguments
		if len(fields) != 5 {
			// empty line
			if len(fields) == 0 {
				continue
			}

			// accept end and break
			if len(fields) == 1 && fields[0] == "end" {
				break
			}

			if fields[0] != "org" {
				return WarriorData{}, fmt.Errorf("line %d: invalid op-code '%s'", lineNum, fields[0])
			} else if len(fields) != 2 {
				return WarriorData{}, fmt.Errorf("line %d: 'org' requires 1 argument", lineNum)
			}

			val, err := strconv.ParseInt(fields[1], 10, 32)
			if err != nil {
				return WarriorData{}, fmt.Errorf("line %d: error parsing integer: %s", lineNum, err)
			}
			if val < 0 {
				return WarriorData{}, fmt.Errorf("line %d: start address outside warrior code", lineNum)
			}

			data.Start = int(val)
			continue
		}

		// comma is ignored, but required
		if !strings.Contains(lower, ",") {
			return WarriorData{}, fmt.Errorf("line %d: missing comma", lineNum)
		}

		op, opmode, err := getOp94(fields[0])
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", lineNum, err)
		}

		amode, err := getAddressMode(fields[1])
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", lineNum, err)
		}
		aval, err := parseAddress(fields[2], coresize)
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: error parsing a field integer: %s", lineNum, err)
		}

		bmode, err := getAddressMode(fields[3])
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", lineNum, err)
		}
		bval, err := parseAddress(fields[4], coresize)
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: error parsing b field integer: %s", lineNum, err)
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

	if data.Start >= len(data.Code) {
		return WarriorData{}, fmt.Errorf("invalid start position")
	}

	return data, nil
}

func getOpMode94(Op OpCode, AMode AddressMode, BMode AddressMode) (OpMode, error) {
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
		if AMode == IMMEDIATE {
			return AB, nil
		} else {
			return B, nil
		}
	case ADD:
		fallthrough
	case SUB:
		// ADD and SUB:
		// AB if #A, F otherwise, no #B allowed
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
		if AMode == IMMEDIATE {
			return AB, nil
		} else {
			return B, nil
		}
	case ADD:
		fallthrough
	case SUB:
		// ADD and SUB:
		// AB if #A, F otherwise, no #B allowed
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

func parseLoadFile88(reader io.Reader, coresize Address) (WarriorData, error) {
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
		// empty lines and last lines without newlines seem to be missed
		// should something else be used? or are these not worth handling?
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

		// remove comma before counting fields
		nocomma := strings.ReplaceAll(lower, ",", " ")

		// split into fields based on whitespace
		fields := strings.Fields(nocomma)

		// valid instructions need exactly 5 fields
		// only other option is "END" pseudo opcode with 0 or 1 arguments
		if len(fields) != 5 {
			// empty line
			if len(fields) == 0 {
				continue
			}

			if fields[0] != "end" && fields[0] != "org" {
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
			if fields[0] != "org" && (val < 0 || val > int64(len(data.Code))) {
				return WarriorData{}, fmt.Errorf("line %d: start address outside warrior code", lineNum)
			}

			data.Start = int(val)

			if fields[0] == "end" {
				break
			}
		}

		// comma is ignored, but required
		if !strings.Contains(lower, ",") {
			return WarriorData{}, fmt.Errorf("line %d: missing comma", lineNum)
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

	if data.Start != 0 && data.Start >= len(data.Code) {
		return WarriorData{}, fmt.Errorf("invalid start position")
	}

	return data, nil
}

func ParseLoadFile(reader io.Reader, simConfig SimulatorConfig) (WarriorData, error) {
	if simConfig.Mode == ICWS88 {
		return parseLoadFile88(reader, simConfig.CoreSize)
	}
	return parseLoadFile94(reader, simConfig.CoreSize)
}
