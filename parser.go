package gmars

import (
	"fmt"
	"strings"
)

type lineType uint8

const (
	lineEmpty lineType = iota
	lineInstruction
	linePseudoOp
	lineComment
)

type sourceLine struct {
	line     int
	codeLine int
	typ      lineType

	// string values of input to parse tokens from lexer into
	labels   []string
	op       string
	amode    string
	a        []token
	bmode    string
	b        []token
	comment  string
	newlines int
}

func subExprSymbol(expr []token, label string, value []token) []token {
	output := make([]token, 0, len(expr))
	for _, tok := range expr {
		if tok.typ == tokText && tok.val == label {
			output = append(output, value...)
		} else {
			output = append(output, tok)
		}
	}
	return output
}

func (line sourceLine) subSymbol(label string, value []token) sourceLine {
	// output := make()
	line.a = subExprSymbol(line.a, label, value)
	line.b = subExprSymbol(line.b, label, value)
	return line
}

type parser struct {
	lex tokenReader

	// state for the running parser
	nextToken   token
	line        int
	codeLine    int
	atEOF       bool
	err         error
	currentLine sourceLine
	metadata    WarriorData
	endSeen     bool

	// collected lines
	lines []sourceLine

	// maps of symbol definitions and references used to verify that each
	// symbol is defined exactly once and each reference is defined.
	symbols    map[string]int
	references map[string]int
}

func newParser(lex tokenReader) *parser {
	p := &parser{
		lex:        lex,
		symbols:    make(map[string]int),
		references: make(map[string]int),
		line:       1,
	}
	p.next()

	p.loadPredefinedSymbols()
	return p
}

type parseStateFn func(p *parser) parseStateFn

func (p *parser) loadPredefinedSymbols() {
	p.symbols["CORESIZE"] = -1
	p.symbols["MAXLENGTH"] = -1
	p.symbols["MAXPROCESSES"] = -1
	p.symbols["MINDISTANCE"] = -1
	// p.symbols["CURLINE"] = -1
}

// parse runs the state machine. the main flows are:
//
// code lines:
//
//	line -> labels -> op -> aMode -> aExpr -> bMode -> bExpr -> line
//	line -> labels -> op -> aMode -> aExpr -> line
//	line -> labels -> psuedoOp -> expr -> line
//
// empty line:
//
//	line -> emptyLines -> line
//
// comment line:
//
//	line -> line
func (p *parser) parse() ([]sourceLine, WarriorData, error) {
	for state := parseLine; state != nil; {
		state = state(p)
	}
	if p.err != nil {
		return nil, WarriorData{}, p.err
	}

	err := p.validateSymbols()
	if err != nil {
		return nil, WarriorData{}, err
	}

	return p.lines, p.metadata, nil
}

func (p *parser) validateSymbols() error {
	for symbol, i := range p.references {
		_, ok := p.symbols[symbol]
		if !ok {
			return fmt.Errorf("line %d: symbol '%s' undefined", i, symbol)
		}
	}
	return nil
}

func (p *parser) next() token {
	if p.atEOF {
		return token{typ: tokEOF}
	}

	nextToken, err := p.lex.NextToken()
	if err != nil {
		p.atEOF = true
		return p.nextToken
	}

	lastToken := p.nextToken
	p.nextToken = nextToken

	if lastToken.typ == tokNewline {
		p.line += 1
	}
	return lastToken
}

// helper function to emit the current working line and consume
// the current token. return nextState or nil on EOF
func (p *parser) consumeEmitLine(nextState parseStateFn) parseStateFn {
	// consume current character
	nextToken := p.next()

	if p.nextToken.typ != tokNewline {
		p.err = fmt.Errorf("expected newline, got: '%s'", p.nextToken)
		return nil
	}

	p.currentLine.newlines += 1
	p.lines = append(p.lines, p.currentLine)

	nextToken = p.next()
	if nextToken.typ == tokEOF {
		return nil
	}
	return nextState
}

// initial state, dispatches to new states based on the first token:
// newline: parseEmptyLines
// comment: parseComment
// text: parseLabels
// eof: nil
// anything else: error
func parseLine(p *parser) parseStateFn {
	if p.endSeen {
		return nil
	}

	p.currentLine = sourceLine{line: p.line}

	switch p.nextToken.typ {
	case tokNewline:
		p.currentLine.typ = lineEmpty
		return parseEmptyLines
	case tokComment:
		if strings.HasPrefix(p.nextToken.val, ";name") {
			p.metadata.Name = strings.TrimSpace(p.nextToken.val[5:])
		} else if strings.HasPrefix(p.nextToken.val, ";author") {
			p.metadata.Author = strings.TrimSpace(p.nextToken.val[7:])
		} else if strings.HasPrefix(p.nextToken.val, ";strategy") {
			if len(p.nextToken.val) > 10 {
				p.metadata.Strategy += p.nextToken.val[10:] + "\n"
			}
		}
		p.currentLine.typ = lineComment
		return parseComment
	case tokText:
		return parseLabels
	case tokEOF:
		return nil
	default:
		p.err = fmt.Errorf("line %d: unexpected token: '%s' type %d", p.line, p.nextToken, p.nextToken.typ)
		return nil
	}
}

// parseNewlines consumes newlines and then returns:
// eof: nil
// anything else: parseLine
func parseEmptyLines(p *parser) parseStateFn {
	for p.nextToken.typ == tokNewline {
		p.currentLine.newlines += 1
		p.next()
	}
	p.lines = append(p.lines, p.currentLine)
	return parseLine
}

// parseComment emits a comment and deals with newlines
// newline: parseLine
func parseComment(p *parser) parseStateFn {
	p.currentLine.comment = p.nextToken.val
	return p.consumeEmitLine(parseLine)
}

// parseLabels consumes text tokens until an op is read
// label text token: parseLabels
// op text token: parseOp
// colon: parseColon
// newline / comments: consume
// anyting else: nil
func parseLabels(p *parser) parseStateFn {
	// just consume newlines and comments for now
	if p.nextToken.typ == tokNewline || p.nextToken.typ == tokComment {
		p.next()
		return parseLabels
	}

	if p.nextToken.IsOp() {
		if p.nextToken.IsPseudoOp() {
			return parsePseudoOp
		}
		return parseOp
	}

	if p.nextToken.typ == tokColon {
		return parseColon
	}

	_, ok := p.symbols[p.nextToken.val]
	if ok {
		p.err = fmt.Errorf("line %d: symbol '%s' redefined", p.line, p.nextToken.val)
	}

	p.symbols[p.nextToken.val] = p.line
	p.currentLine.labels = append(p.currentLine.labels, p.nextToken.val)
	nextToken := p.next()

	if nextToken.typ != tokText {
		p.err = fmt.Errorf("line %d: label or op expected, got '%s'", p.line, nextToken)
		return nil
	}
	return parseLabels
}

// from: parseLabels
// newline or comments: parseColon
// op text: parseOp
// psudo op text: parsePseudoOp
// text: parseLabels
// anything else: nil
func parseColon(p *parser) parseStateFn {
	for p.nextToken.typ == tokColon {
		p.next()
	}

	// just consume newlines and comments for now
	if p.nextToken.typ == tokNewline || p.nextToken.typ == tokComment {
		p.next()
		return parseColon
	}

	if p.nextToken.IsOp() {
		if p.nextToken.IsPseudoOp() {
			return parsePseudoOp
		}
		return parseOp
	}

	if p.nextToken.typ == tokText {
		return parseLabels
	}

	p.err = fmt.Errorf("line %d: op expected after colon, got '%s'", p.line, p.nextToken)
	return nil
}

// from: parseLabels
// comment: parseComment
// newline: parseLine
// exprssionterm: parsePseudoExpr
// anything else: error
func parsePseudoOp(p *parser) parseStateFn {
	p.currentLine.op = p.nextToken.val
	p.currentLine.typ = linePseudoOp

	if strings.ToLower(p.nextToken.val) == "end" {
		p.endSeen = true
	}

	lastToken := p.nextToken
	p.next()

	if p.nextToken.IsExpressionTerm() {
		return parsePseudoExpr
	} else if p.nextToken.typ == tokComment {
		return parseComment
	} else if p.nextToken.typ == tokEOF {
		if lastToken.NoOperandsOk() {
			p.next()
			p.currentLine.newlines += 1
			p.lines = append(p.lines, p.currentLine)
			return nil
		}
	} else if p.nextToken.typ == tokNewline {
		if lastToken.NoOperandsOk() {
			p.next()
			p.currentLine.newlines += 1
			p.lines = append(p.lines, p.currentLine)
			return parseLine
		}
		p.err = fmt.Errorf("line %d: expected operand expression after psuedo-op '%s', got newline", p.line, lastToken.val)
		return nil
	}

	p.err = fmt.Errorf("line %d: expected operand expression, comment, or newline after pseudo-op, got: '%s'", p.line, p.nextToken)
	return nil
}

// from: parsePseudoOp
// on comment: parseComment
// expressionterm: parsePseudoExpr
// anything else: error
func parsePseudoExpr(p *parser) parseStateFn {
	if p.currentLine.a == nil {
		p.currentLine.a = make([]token, 0)
	}

	for p.nextToken.IsExpressionTerm() {
		if p.nextToken.typ == tokText {
			_, ok := p.references[p.nextToken.val]
			if !ok {
				p.references[p.nextToken.val] = p.line
			}
		}
		p.currentLine.a = append(p.currentLine.a, p.nextToken)
		p.next()
	}
	switch p.nextToken.typ {
	case tokComment:
		return parseComment
	case tokNewline:
		p.next()
		p.currentLine.newlines += 1
		p.lines = append(p.lines, p.currentLine)
		return parseLine
	case tokEOF:
		p.lines = append(p.lines, p.currentLine)
		return parseLine
	default:
		p.err = fmt.Errorf("line %d: expected comment or newline after expression, got '%s'", p.line, p.nextToken)
		return nil
	}
}

// from: parseLabels
// addressmode: parseModeA
// expressionterm: parseExprA
// anything else: error
func parseOp(p *parser) parseStateFn {
	p.currentLine.op = p.nextToken.val
	p.currentLine.typ = lineInstruction
	p.currentLine.codeLine = p.codeLine
	p.codeLine += 1

	p.next()

	if p.nextToken.IsExpressionTerm() && p.nextToken.val != "*" {
		return parseExprA
	}

	switch p.nextToken.typ {
	case tokAddressMode:
		return parseModeA
	case tokExprOp:
		if p.nextToken.val == "*" {
			return parseModeA
		}
		return parseExprA
	default:
		p.err = fmt.Errorf("line %d: expected operand expression after op, got '%s'", p.line, p.nextToken)
		return nil
	}
}

// from: parseOp
// experssionterm: parseExprA
// anything else: error
func parseModeA(p *parser) parseStateFn {
	p.currentLine.amode = p.nextToken.val
	p.next()
	if p.nextToken.IsExpressionTerm() {
		return parseExprA
	}
	p.err = fmt.Errorf("line %d: expected address mode or operand expression, got '%s'", p.line, p.nextToken)
	return nil
}

// from: parseOp, parseModeA
// expression term: recursively consume tokens to exprA
// comma: parseComma
// comment: parseComment
// newline: emit and parseLine
// anything else: error
func parseExprA(p *parser) parseStateFn {
	if p.currentLine.a == nil {
		p.currentLine.a = make([]token, 0)
	}

	for p.nextToken.IsExpressionTerm() {
		if p.nextToken.typ == tokText {
			_, ok := p.references[p.nextToken.val]
			if !ok {
				p.references[p.nextToken.val] = p.line
			}
		}
		p.currentLine.a = append(p.currentLine.a, p.nextToken)
		p.next()
	}
	switch p.nextToken.typ {
	case tokComment:
		return parseComment
	case tokComma:
		return parseComma
	case tokNewline:
		fallthrough
	case tokEOF:
		p.lines = append(p.lines, p.currentLine)
		return parseLine
	default:
		p.err = fmt.Errorf("line %d: expected comma or newline after op, got '%s'", p.line, p.nextToken)
		return nil
	}
}

// from: parseExprA
// addressmode: parseModeB
// expression term: parseExprB
// anything else: error
func parseComma(p *parser) parseStateFn {
	p.next()

	if p.nextToken.typ == tokAddressMode || (p.nextToken.typ == tokExprOp && p.nextToken.val == "*") {
		return parseModeB
	} else if p.nextToken.IsExpressionTerm() {
		return parseExprB
	} else {
		p.err = fmt.Errorf("expected address mode or expression after comma, got '%s'", p.nextToken)
		return nil
	}
}

// from: parseComma
// expressionterm: parseExprB
// anything else: error
func parseModeB(p *parser) parseStateFn {
	p.currentLine.bmode = p.nextToken.val
	p.next()
	if p.nextToken.IsExpressionTerm() {
		return parseExprB
	}
	p.err = fmt.Errorf("line %d: expected address mode or operand expression, got '%s'", p.line, p.nextToken)
	return nil
}

// from parseComma, parseModeB
// expressionTerm: recursively consume tokens to exprB
// comment: parseComment
// newline: parseLine
// anything else: error
func parseExprB(p *parser) parseStateFn {
	if p.currentLine.b == nil {
		p.currentLine.b = make([]token, 0)
	}

	for p.nextToken.IsExpressionTerm() {
		if p.nextToken.typ == tokText {
			_, ok := p.references[p.nextToken.val]
			if !ok {
				p.references[p.nextToken.val] = p.line
			}
		}
		p.currentLine.b = append(p.currentLine.b, p.nextToken)
		p.next()
	}

	switch p.nextToken.typ {
	case tokComment:
		return parseComment
	case tokNewline:
		p.currentLine.newlines += 1
		p.lines = append(p.lines, p.currentLine)
		p.next()
		return parseLine
	case tokEOF:
		p.lines = append(p.lines, p.currentLine)
		return parseLine
	default:
		p.err = fmt.Errorf("line %d: expected comma or newline after op, got '%s'", p.line, p.nextToken)
		return nil
	}
}
