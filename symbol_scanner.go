package gmars

import (
	"fmt"
	"strings"
)

// symbol scanner accepts a tokenReader and scans for any
// equ symbols contained. Symbols defined inside for loops
// are ignored, allowing us to run the same code both before
// and after for loops have been expanded.
type symbolScanner struct {
	lex tokenReader

	nextToken token
	atEOF     bool
	valBuf    []token
	labelBuf  []string
	forLevel  int
	err       error

	symbols map[string][]token
}

type scanStateFn func(p *symbolScanner) scanStateFn

func newSymbolScanner(lex tokenReader) *symbolScanner {
	pre := &symbolScanner{
		lex:     lex,
		symbols: make(map[string][]token),
	}

	pre.next()

	return pre
}

func (p *symbolScanner) next() token {
	if p.atEOF {
		return token{typ: tokEOF}
	}
	tok, err := p.lex.NextToken()
	if err != nil {
		p.atEOF = true
		return token{tokError, fmt.Sprintf("%s\n", err)}
	}
	if tok.typ == tokEOF || tok.typ == tokError {
		p.atEOF = true
	}
	retTok := p.nextToken
	p.nextToken = tok
	return retTok
}

// run the preprocessor
func (p *symbolScanner) ScanInput() (map[string][]token, error) {
	for state := scanLine; state != nil; {
		state = state(p)
	}
	if p.err != nil {
		return nil, p.err
	}
	return p.symbols, nil
}

func (p *symbolScanner) consume(nextState scanStateFn) scanStateFn {
	p.next()
	if p.nextToken.typ == tokEOF {
		return nil
	}
	return nextState
}

// run at start of each line
// on text: preLabels
// on other: preConsumeLine
func scanLine(p *symbolScanner) scanStateFn {
	switch p.nextToken.typ {
	case tokText:
		p.labelBuf = make([]string, 0)
		return scanLabels
	default:
		return scanConsumeLine
	}
}

// text equ: consumeValue
// text op: consumLine
// text default: scanLabels
// anything else: consumeLine
func scanLabels(p *symbolScanner) scanStateFn {
	switch p.nextToken.typ {
	case tokText:
		if p.nextToken.IsPseudoOp() {
			opLower := strings.ToLower(p.nextToken.val)
			switch opLower {
			case "equ":
				if p.forLevel == 0 {
					p.valBuf = make([]token, 0)
					return p.consume(scanEquValue)
				}
			case "for":
				p.forLevel++
				return scanConsumeLine
			case "rof":
				if p.forLevel > 0 {
					p.forLevel--
				}
				return scanConsumeLine
			case "end":
				if p.forLevel > 1 {
					return scanConsumeLine
				} else {
					return nil
				}
			default:
				return scanConsumeLine
			}
		} else if p.nextToken.IsOp() {
			return scanConsumeLine
		} else if p.nextToken.typ == tokInvalid {
			return nil
		}
		p.labelBuf = append(p.labelBuf, p.nextToken.val)
		return p.consume(scanLabels)
	case tokComment:
		fallthrough
	case tokNewline:
		return p.consume(scanLabels)
	case tokEOF:
		return nil
	default:
		return scanConsumeLine
	}
}

func scanConsumeLine(p *symbolScanner) scanStateFn {
	switch p.nextToken.typ {
	case tokNewline:
		return p.consume(scanLine)
	case tokError:
		return nil
	case tokEOF:
		return nil
	default:
		return p.consume(scanConsumeLine)
	}
}

func scanEquValue(p *symbolScanner) scanStateFn {
	for p.nextToken.typ != tokNewline && p.nextToken.typ != tokEOF && p.nextToken.typ != tokError {
		p.valBuf = append(p.valBuf, p.nextToken)
		p.next()
	}
	for _, label := range p.labelBuf {
		_, ok := p.symbols[label]
		if ok {
			p.err = fmt.Errorf("symbol '%s' redefined", label)
			return nil
		}
		p.symbols[label] = p.valBuf
	}
	p.valBuf = make([]token, 0)
	p.labelBuf = make([]string, 0)
	return p.consume(scanLine)
}
