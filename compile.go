package gmars

import (
	"fmt"
	"io"
	"strings"
)

// compiler holds the input and state required to compile a program.
type compiler struct {
	m         Address // coresize
	lines     []sourceLine
	config    SimulatorConfig
	values    map[string][]token // symbols that represent expressions
	labels    map[string]int     // symbols that represent addresses
	startExpr []token
	metadata  WarriorData
}

func newCompiler(src []sourceLine, metadata WarriorData, config SimulatorConfig) (*compiler, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %s", err)
	}
	return &compiler{
		lines:     src,
		config:    config,
		metadata:  metadata,
		m:         config.CoreSize,
		startExpr: []token{{tokNumber, "0"}},
	}, nil
}

// load symbol []token values into value map and code line numbers of
// instruction labels into the label map
func (c *compiler) loadSymbols() {
	c.values = make(map[string][]token)
	c.labels = make(map[string]int)

	curPseudoLine := 0
	for _, line := range c.lines {
		if line.typ == linePseudoOp {
			if strings.ToLower(line.op) == "equ" {
				for _, label := range line.labels {
					c.values[label] = line.a
				}
			} else if strings.ToLower(line.op) == "org" {
				c.startExpr = line.a
			} else if strings.ToLower(line.op) == "end" {
				if len(line.a) > 0 {
					c.startExpr = line.a
				}
				for _, label := range line.labels {
					c.labels[label] = curPseudoLine
				}
			}

		}
		if line.typ == lineInstruction {
			for _, label := range line.labels {
				c.labels[label] = line.codeLine
			}
			curPseudoLine++
		}
	}
}

func (c *compiler) reloadReferences() error {
	c.labels = make(map[string]int)

	curPseudoLine := 1
	for _, line := range c.lines {
		if line.typ == lineInstruction {
			for _, label := range line.labels {
				_, ok := c.labels[label]
				if ok {
					return fmt.Errorf("line %d: label '%s' redefined", line.line, label)
				}
				c.labels[label] = line.codeLine
				curPseudoLine++
			}
		} else if line.typ == linePseudoOp {
			for _, label := range line.labels {
				_, ok := c.labels[label]
				if ok {
					return fmt.Errorf("line %d: label '%s' redefined", line.line, label)
				}
				c.labels[label] = curPseudoLine
			}
		}
	}

	return nil
}

func (c *compiler) expandExpression(expr []token, line int) ([]token, error) {
	input := expr
	var output []token

	for !exprEqual(input, output) {
		if len(output) > 0 {
			input = output
		}

		output = make([]token, 0)
		for _, tok := range input {
			if tok.typ == tokText {
				val, valOk := c.values[tok.val]
				if valOk {
					output = append(output, val...)
					continue
				}

				label, labelOk := c.labels[tok.val]
				if labelOk {
					val := (label - line) % int(c.m)
					if val < 0 {
						output = append(output, token{tokExprOp, "-"}, token{tokNumber, fmt.Sprintf("%d", -val)})
					} else {
						output = append(output, token{tokNumber, fmt.Sprintf("%d", val)})
					}
				} else {
					return nil, fmt.Errorf("unresolved symbol '%s'", tok.val)
				}
			} else {
				output = append(output, tok)
			}
		}
	}
	return output, nil
}

func (c *compiler) assembleLine(in sourceLine) (Instruction, error) {
	opLower := strings.ToLower(in.op)
	var aMode, bMode AddressMode
	if in.amode == "" {
		if c.config.Mode == ICWS88 && opLower == "dat" {
			aMode = IMMEDIATE
		} else {
			aMode = DIRECT
		}
	} else {
		mode, err := getAddressMode(in.amode)
		if err != nil {
			return Instruction{}, fmt.Errorf("invalid amode: '%s'", in.amode)
		}
		aMode = mode
	}
	if in.bmode == "" {
		if c.config.Mode == ICWS88 && opLower == "dat" {
			bMode = IMMEDIATE
		} else {
			bMode = DIRECT
		}
	} else {
		mode, err := getAddressMode(in.bmode)
		if err != nil {
			return Instruction{}, fmt.Errorf("invalid bmode: '%s'", in.bmode)
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
	if bVal < 0 {
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

func (c *compiler) expandFor(start, end int) error {
	output := make([]sourceLine, 0)
	codeLineIndex := 0

	// concatenate lines preceding start
	for i := 0; i < start; i++ {
		// curLine := c.lines[i]
		if c.lines[i].typ == lineInstruction {
			// curLine.line = codeLineIndex
			codeLineIndex++
		}
		output = append(output, c.lines[i])
	}

	// get labels and count from for line
	labels := c.lines[start].labels

	countExpr, err := c.expandExpression(c.lines[start].a, start)
	if err != nil {
		return err
	}
	count, err := evaluateExpression(countExpr)
	if err != nil {
		return fmt.Errorf("line %d: invalid for count '%s", c.lines[start].line, c.lines[start].a)
	}

	for j := 1; j <= count; j++ {
		for i := start + 1; i < end; i++ {
			if c.lines[i].typ == lineInstruction {
				thisLine := c.lines[i]

				// subtitute symbols in line
				for iLabel, label := range labels {
					var newValue []token
					if iLabel == len(labels)-1 {
						newValue = []token{{tokNumber, fmt.Sprintf("%d", j)}}
					} else {
						if j == 1 {
							newValue = []token{{tokNumber, "0"}}
						} else {
							newValue = []token{{tokExprOp, "-"}, {tokNumber, fmt.Sprintf("%d", -(1 - j))}}
						}
					}
					thisLine = thisLine.subSymbol(label, newValue)
				}

				// update codeLine
				thisLine.codeLine = codeLineIndex
				codeLineIndex++

				output = append(output, thisLine)
			} else {
				output = append(output, c.lines[i])
			}
		}

	}

	// continue appending lines until the end of the file
	for i := end + 1; i < len(c.lines); i++ {
		if c.lines[i].typ == lineInstruction {
			thisLine := c.lines[i]
			thisLine.codeLine = codeLineIndex
			codeLineIndex++
			output = append(output, thisLine)
		} else {
			output = append(output, c.lines[i])
		}
	}

	c.lines = output
	return c.reloadReferences()
}

// look for for statements from the bottom up. if one is found it is expanded
// and the function calls itself again.
func (c *compiler) expandForLoops() error {
	rofSourceIndex := -1
	for i := len(c.lines) - 1; i >= 0; i-- {
		if c.lines[i].typ == linePseudoOp {
			lop := strings.ToLower(c.lines[i].op)
			if lop == "rof" {
				rofSourceIndex = i
			} else if lop == "for" {
				if rofSourceIndex == -1 {
					return fmt.Errorf("line %d: unmatched for", c.lines[i].codeLine)
				}
				err := c.expandFor(i, rofSourceIndex)
				if err != nil {
					return err
				}
				return c.expandForLoops()
			}
		}
	}
	if rofSourceIndex != -1 {
		return fmt.Errorf("line %d: unmatched rof", c.lines[rofSourceIndex].line)
	}
	return nil
}

func (c *compiler) compile() (WarriorData, error) {

	c.loadSymbols()

	graph := buildReferenceGraph(c.values)
	cyclic, cyclicKey := graphContainsCycle(graph)
	if cyclic {
		return WarriorData{}, fmt.Errorf("expression '%s' is cyclic", cyclicKey)
	}

	resolved, err := expandExpressions(c.values, graph)
	if err != nil {
		return WarriorData{}, err
	}
	c.values = resolved

	err = c.expandForLoops()
	if err != nil {
		return WarriorData{}, err
	}

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

	startExpr, err := c.expandExpression(c.startExpr, 0)
	if err != nil {
		return WarriorData{}, fmt.Errorf("invalid start expression")
	}
	startVal, err := evaluateExpression(startExpr)
	if err != nil {
		return WarriorData{}, fmt.Errorf("invalid start expression: %s", err)
	}
	if startVal < 0 || startVal > len(code) {
		return WarriorData{}, fmt.Errorf("invalid start value: %d", startVal)
	}

	c.metadata.Code = code
	c.metadata.Start = startVal

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
