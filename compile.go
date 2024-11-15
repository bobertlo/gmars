package gmars

import (
	"fmt"
)

// compiler holds the input and state required to compile a program.
type compiler struct {
	lines  []sourceLine
	config SimulatorConfig
	values map[string][]token // symbols that represent expressions
	labels map[string]int     // symbols that represent addresses
	code   []Instruction
}

func newCompiler(src []sourceLine, config SimulatorConfig) (*compiler, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid condif: %s", err)
	}
	return &compiler{lines: src, config: config}, nil
}

// load symbol []token values into value map and code line numbers of
// instruction labels into the label map
func (c *compiler) loadSymbols() {
	c.values = make(map[string][]token)
	c.labels = make(map[string]int)

	for _, line := range c.lines {
		if line.typ == linePseudoOp && line.op == "equ" {
			for _, label := range line.labels {
				c.values[label] = line.a
			}
		}
		if line.typ == lineInstruction {
			for _, label := range line.labels {
				c.labels[label] = line.codeLine
			}
		}
	}
}

func (c *compiler) compile() (WarriorData, error) {
	c.loadSymbols()

	graph := buildReferenceGraph(c.values)
	cyclic, cyclicKey := graphContainsCycle(graph)
	if cyclic {
		return WarriorData{}, fmt.Errorf("expressiong '%s' is cyclic", cyclicKey)
	}

	resolved, err := expandExpressions(c.values, graph)
	if err != nil {
		return WarriorData{}, err
	}
	c.values = resolved

	return WarriorData{}, fmt.Errorf("not implemented")
}
