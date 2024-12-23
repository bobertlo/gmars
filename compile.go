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

func (c *compiler) loadConstants() {
	c.values["CORESIZE"] = []token{{tokNumber, fmt.Sprintf("%d", c.config.CoreSize)}}
	c.values["MAXLENGTH"] = []token{{tokNumber, fmt.Sprintf("%d", c.config.Length)}}
	c.values["MAXPROCESSES"] = []token{{tokNumber, fmt.Sprintf("%d", c.config.Processes)}}
	c.values["MINDISTANCE"] = []token{{tokNumber, fmt.Sprintf("%d", c.config.Distance)}}
	// c.values["CURLINE"] = []token{{tokNumber, "0"}}
}

// load symbol []token values into value map and code line numbers of
// instruction labels into the label map
func (c *compiler) loadSymbols() {
	c.values = make(map[string][]token)
	c.labels = make(map[string]int)

	c.loadConstants()

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
						output = append(output, token{tokSymbol, "-"}, token{tokNumber, fmt.Sprintf("%d", -val)})
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

func (c *compiler) evaluateAssertion(assertText string) error {

	assertTokens, err := LexInput(strings.NewReader(assertText))
	if err != nil {
		return err
	}
	assertTokens = assertTokens[:len(assertTokens)-1]
	exprTokens, err := c.expandExpression(assertTokens, 0)
	if err != nil {
		return err
	}
	exprVal, err := evaluateExpression(exprTokens)
	if err != nil {
		return err
	}
	if exprVal == 0 {
		return fmt.Errorf("assertion '%s' failed", assertText)
	}
	return nil
}

func (c *compiler) evaluateAssertions() error {
	for _, line := range c.lines {
		if line.typ != lineComment {
			continue
		}
		if strings.HasPrefix(line.comment, ";assert") {
			assertText := line.comment[7:]
			err := c.evaluateAssertion(assertText)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

func (c *compiler) compile() (WarriorData, error) {
	c.loadSymbols()

	err := c.evaluateAssertions()
	if err != nil {
		return WarriorData{}, err
	}

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
		// c.values["CURLINE"] = []token{{tokNumber, fmt.Sprintf("%d", len(code))}}
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
	tokens, err := lexer.Tokens()
	if err != nil {
		return WarriorData{}, err
	}

	depth := 0
	for {
		symbols, forSeen, err := ScanInput(newBufTokenReader(tokens))
		if err != nil {
			return WarriorData{}, fmt.Errorf("symbol scanner: %s", err)
		}
		if forSeen {
			expandedTokens, err := ForExpand(newBufTokenReader(tokens), symbols)
			if err != nil {
				return WarriorData{}, fmt.Errorf("for: %s", err)
			}
			tokens = expandedTokens
			// oops the embedded for loops are not implemented
			// break
		} else {
			break
		}
		depth++
		if depth > 12 {
			return WarriorData{}, fmt.Errorf("for loop depth exceeded")
		}
	}

	parser := newParser(newBufTokenReader(tokens))
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
