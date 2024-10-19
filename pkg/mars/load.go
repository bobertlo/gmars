package mars

import (
	"bufio"
	"io"
)

func (m *MARS) LoadWarrior(reader io.Reader) (*Warrior, error) {
	breader := bufio.NewReader(reader)
	for {
		line, err := breader.ReadString('\n')
		if err != nil {
			break
		}

		if len(line) < 1 {
			continue
		}

		line = line[:len(line)-1]

		// fmt.Println(line)
	}

	data := &WarriorData{
		Name:     "Unknown",
		Author:   "Anonymous",
		Strategy: "",
		Code:     make([]Instruction, 0),
		Start:    0,
	}

	w := &Warrior{
		data: data,
		sim:  m,
	}

	return w, nil
}
