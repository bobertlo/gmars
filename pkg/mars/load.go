package mars

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func (m *MARS) LoadWarrior(reader io.Reader) (*Warrior, error) {
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
