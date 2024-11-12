package gmars

import "fmt"

type lineType uint8

const (
	lineEmpty lineType = iota
	lineInstruction
	linePseudoOp
	lineComment
)

type sourceLine struct {
	line int
	typ  lineType

	// string values of input to parse tokens from lexer into
	labels   []string
	op       string
	amode    string
	a        expression
	bmode    string
	b        expression
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
	return &sourceFile{lines: p.lines, symbols: p.symbols}, nil
}

func (p *parser) next() (token, bool) {
	if p.atEOF {
		return token{}, true
	}

	nextToken, err := p.lex.NextToken()
	if err != nil {
		p.atEOF = true
		return p.nextToken, true
	}

	lastToken := p.nextToken
	p.nextToken = nextToken

	if lastToken.typ == tokNewline {
		p.line += 1
	}
	return lastToken, false
}

// helper function to emit the current working line and consume
// the current token. return nextState or nil on EOF
func (p *parser) consumeEmitLine(nextState parseStateFn) parseStateFn {
	// consume current character
	_, eof := p.next()
	if eof {
		return nil
	}

	if p.nextToken.typ != tokNewline {
		p.err = fmt.Errorf("expected newline, got: '%s'", p.nextToken)
		return nil
	}

	p.currentLine.newlines += 1
	p.lines = append(p.lines, p.currentLine)

	_, eof = p.next()
	if eof {
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
	}

	return nil
}

// parseNewlines consumes newlines and then returns:
// eof: nil
// anything else: parseLine
func parseEmptyLines(p *parser) parseStateFn {
	for p.nextToken.typ == tokNewline {
		p.currentLine.newlines += 1
		_, eof := p.next()
		if eof {
			p.lines = append(p.lines, p.currentLine)
			return nil
		}
	}
	p.lines = append(p.lines, p.currentLine)
	return parseLine
}

// parseComment emits a comment and deals with newlines
// newline: parseLine
// eof: nil
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

	p.currentLine.labels = append(p.currentLine.labels, p.nextToken.val)
	nextToken, eof := p.next()
	if eof {
		p.err = fmt.Errorf("line %d: label or op expected, got eof", p.line)
		return nil
	}

	if nextToken.typ != tokText {
		p.err = fmt.Errorf("line %d: label or op expected, got '%s'", p.line, nextToken)
		return nil
	}
	return parseLabels
}

func parseOp(p *parser) parseStateFn {
	return nil
}

func parsePseudoOp(p *parser) parseStateFn {
	return nil
}

func parseModeA(p *parser) parseStateFn {
	return nil
}

func parseExprA(p *parser) parseStateFn {
	return nil
}

func parseModeB(p *parser) parseStateFn {
	return nil
}

func parseExprB(p *parser) parseStateFn {
	return nil
}

func parsePseudoExpr(p *parser) parseStateFn {
	return nil
}
