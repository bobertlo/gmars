package gmars

import (
	"fmt"
)

// compiler holds the input and state required to compile a program.
type compiler struct {
	m      Address // coresize
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
	return &compiler{lines: src, config: config, m: config.CoreSize}, nil
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

func (c *compiler) expandExpression(expr []token, line int) ([]token, error) {
	output := make([]token, 0)
	for _, tok := range expr {
		if tok.typ == tokText {
			val, valOk := c.values[tok.val]
			if valOk {
				output = append(output, val...)
				continue
			}

			label, labelOk := c.labels[tok.val]
			if labelOk {
				val := (label - line) % int(c.m)
				output = append(output, token{tokNumber, fmt.Sprintf("%d", val)})
			} else {
				return nil, fmt.Errorf("unresolved symbol '%s'", tok.val)
			}
		} else {
			output = append(output, tok)
		}
	}
	return output, nil
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
