package mars

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

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
	case "spl":
		return SPL, nil
	default:
		return 0, fmt.Errorf("invalid opcode '%s'", op)
	}
}

func getOpMode(opMode string) (OpMode, error) {
	switch strings.ToLower(opMode) {
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
		return 0, fmt.Errorf("invalid op mode: '%s'", opMode)
	}
}

func (m *Simulator) LoadWarrior(reader io.Reader) (*Warrior, error) {
	data := &WarriorData{
		Name:     "Unknown",
		Author:   "Anonymous",
		Strategy: "",
		Code:     make([]Instruction, 0),
		Start:    0,
	}

	breader := bufio.NewReader(reader)
	for {
		raw_line, err := breader.ReadString('\n')
		if err != nil {
			break
		}

		// check for these codes on the raw line
		if strings.HasPrefix(raw_line, ";name") {
			data.Name = strings.TrimSpace(raw_line[5:])
			continue
		}
		if strings.HasPrefix(raw_line, ";author") {
			data.Author = strings.TrimSpace(raw_line[7:])
			continue
		}

		// clean up the line
		line := strings.TrimSpace(raw_line)
		if len(line) == 0 || line[0] == ';' {
			continue
		}
		line = strings.ToLower(line)
		line = strings.Split(line, ";")[0]

		fields := strings.Fields(line)
		if len(fields) < 2 {
			fmt.Println("line to short")
		} else if len(fields) == 2 {
			if fields[0] == "org" && fields[1] == "start" {
				continue
			} else if fields[0] == "end" && fields[1] == "start" {
				continue
			}
		}

		if len(fields) > 3 {
		}
	}

	w := &Warrior{
		data: data,
		sim:  m,
	}

	return w, nil
}
