package gmars

import (
	"fmt"
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

type sourceFile struct {
	lines   []sourceLine
	symbols map[string]int
}

type parser struct {
	lex *lexer

	// state for the running parser
	nextToken   token
	line        int
	codeLine    int
	atEOF       bool
	err         error
	currentLine sourceLine

	// collected lines
	lines []sourceLine

	// line number of symbol definitions
	symbols map[string]int

	// line number of first refernces to symbols to check for
	// undeclared references at the end of the token stream
	references map[string]int
}

func newParser(lex *lexer) *parser {
	p := &parser{
		lex:        lex,
		symbols:    make(map[string]int),
		references: make(map[string]int),
		line:       1,
	}
	p.next()
	return p
}

type parseStateFn func(p *parser) parseStateFn

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
func (p *parser) parse() (*sourceFile, error) {
	for state := parseLine; state != nil; {
		state = state(p)
	}
	if p.err != nil {
		return nil, p.err
	}

	err := p.validateSymbols()
	if err != nil {
		return nil, err
	}

	return &sourceFile{lines: p.lines, symbols: p.symbols}, nil
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
	p.currentLine = sourceLine{line: p.line}

	switch p.nextToken.typ {
	case tokNewline:
		p.currentLine.typ = lineEmpty
		return parseEmptyLines
	case tokComment:
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
// anyting else: nil
func parseLabels(p *parser) parseStateFn {
	if p.nextToken.IsOp() {
		if p.nextToken.IsPseudoOp() {
			return parsePseudoOp
		}
		return parseOp
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
// comment: parseComment
// newline: parseLine
// exprssionterm: parsePseudoExpr
// anything else: error
func parsePseudoOp(p *parser) parseStateFn {
	p.currentLine.op = p.nextToken.val
	p.currentLine.typ = linePseudoOp

	lastToken := p.nextToken
	p.next()

	if p.nextToken.IsExpressionTerm() {
		return parsePseudoExpr
	} else if p.nextToken.typ == tokComment {
		return parseComment
	} else if p.nextToken.typ == tokNewline {
		if lastToken.NoOperandsOk() {
			p.currentLine.newlines += 1
			p.lines = append(p.lines, p.currentLine)
			return nil
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
		fallthrough
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

	if p.nextToken.IsExpressionTerm() {
		return parseExprA
	}

	switch p.nextToken.typ {
	case tokAddressMode:
		return parseModeA
	default:
		p.err = fmt.Errorf("line %d: expected operand expression after op, got '%s'", p.line, p.nextToken)
		return nil
	}
}

// from: parseOp
// experssionterm: parseExprA
// anything else: error
func parseModeA(p *parser) parseStateFn {
	if p.nextToken.typ == tokAddressMode {
		p.currentLine.amode = p.nextToken.val
		p.next()
	}
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
	return parseModeB
}

// from: parseComma
// expressionterm: parseExprB
// anything else: error
func parseModeB(p *parser) parseStateFn {
	if p.nextToken.typ == tokAddressMode {
		p.currentLine.bmode = p.nextToken.val
		p.next()
	}
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
