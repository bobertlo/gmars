package gmars

import (
	"fmt"
	"io"
)

// compiler holds the input and state required to compile a program.
type compiler struct {
	m        Address // coresize
	lines    []sourceLine
	config   SimulatorConfig
	values   map[string][]token // symbols that represent expressions
	labels   map[string]int     // symbols that represent addresses
	metadata WarriorData
}

func newCompiler(src []sourceLine, metadata WarriorData, config SimulatorConfig) (*compiler, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid condif: %s", err)
	}
	return &compiler{
		lines:    src,
		config:   config,
		metadata: metadata,
		m:        config.CoreSize,
	}, nil
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

func (c *compiler) assembleLine(in sourceLine) (Instruction, error) {
	var aMode, bMode AddressMode
	if in.amode == "" {
		aMode = DIRECT
	} else {
		mode, err := getAddressMode(in.amode)
		if err != nil {
			return Instruction{}, fmt.Errorf("invalid amode: '%s'", in.amode)
		}
		aMode = mode
	}
	if in.bmode == "" {
		bMode = DIRECT
	} else {
		mode, err := getAddressMode(in.bmode)
		if err != nil {
			return Instruction{}, fmt.Errorf("invalid amode: '%s'", in.bmode)
		}
		bMode = mode
	}

	var op OpCode
	var opMode OpMode
	if c.config.Mode == ICWS88 {
		op88, err := getOpCode88(in.op)
		if err != nil {
			return Instruction{}, err
		}
		opMode88, err := getOpModeAndValidate88(op88, aMode, bMode)
		if err != nil {
			return Instruction{}, err
		}
		op, opMode = op88, opMode88
	} else {
		op94, opMode94, err := getOp94(in.op)
		if err == nil {
			op, opMode = op94, opMode94
		} else {
			op94, err := getOpCode(in.op)
			if err != nil {
				return Instruction{}, err
			}
			opMode94, err = getOpMode94(op94, aMode, bMode)
			if err != nil {
				return Instruction{}, err
			}
			op, opMode = op94, opMode94
		}
	}

	aExpr, err := c.expandExpression(in.a, in.codeLine)
	if err != nil {
		return Instruction{}, err
	}
	aVal, err := evaluateExpression(aExpr)
	if err != nil {
		return Instruction{}, err
	}

	var bVal int
	if len(in.b) == 0 {
		if op == DAT {
			// move aVal/aMode to B
			bMode = aMode
			bVal = aVal
			// set A to #0
			aMode = IMMEDIATE
			aVal = 0
		} else {
			// set B to #0
			bMode = IMMEDIATE
		}
	} else {
		bExpr, err := c.expandExpression(in.b, in.codeLine)
		if err != nil {
			return Instruction{}, err
		}
		b, err := evaluateExpression(bExpr)
		if err != nil {
			return Instruction{}, err
		}
		bVal = b
	}

	aVal = aVal % int(c.m)
	if aVal < 0 {
		aVal = (int(c.m) + aVal) % int(c.m)
	}
	bVal = bVal % int(c.m)
	if aVal < 0 {
		bVal = (int(c.m) + bVal) % int(c.m)
	}

	return Instruction{
		Op:     op,
		OpMode: opMode,
		AMode:  aMode,
		A:      Address(aVal),
		BMode:  bMode,
		B:      Address(bVal),
	}, nil
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

	code := make([]Instruction, 0)
	for _, line := range c.lines {
		if line.typ != lineInstruction {
			continue
		}

		instruction, err := c.assembleLine(line)
		if err != nil {
			return WarriorData{}, fmt.Errorf("line %d: %s", line.line, err)
		}
		code = append(code, instruction)
	}

	c.metadata.Code = code

	return c.metadata, nil
}

func CompileWarrior(r io.Reader, config SimulatorConfig) (WarriorData, error) {
	lexer := newLexer(r)
	parser := newParser(lexer)
	sourceLines, metadata, err := parser.parse()
	if err != nil {
		return WarriorData{}, err
	}

	compiler, err := newCompiler(sourceLines, metadata, config)
	if err != nil {
		return WarriorData{}, err
	}
	return compiler.compile()
}
