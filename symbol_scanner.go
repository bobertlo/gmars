package gmars

import (
	"fmt"
	"strings"
)

type symbolScanner struct {
	lex tokenReader

	nextToken token
	atEOF     bool
	valBuf    []token
	labelBuf  []string
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
	for state := preLine; state != nil; {
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
func preLine(p *symbolScanner) scanStateFn {
	switch p.nextToken.typ {
	case tokText:
		p.labelBuf = make([]string, 0)
		return preLabels
	default:
		return preConsumeLine
	}
}

// text equ: consumeValue
// text op: consumLine
// text default: preLabels
// anything else: consumeLine
func preLabels(p *symbolScanner) scanStateFn {
	switch p.nextToken.typ {
	case tokText:
		if p.nextToken.IsPseudoOp() {
			if strings.ToLower(p.nextToken.val) == "equ" {
				p.valBuf = make([]token, 0)
				return p.consume(preScanValue)
			} else {
				return preConsumeLine
			}
		} else if p.nextToken.IsOp() {
			return preConsumeLine
		}
		p.labelBuf = append(p.labelBuf, p.nextToken.val)
		return p.consume(preLabels)
	case tokComment:
		fallthrough
	case tokNewline:
		return p.consume(preLabels)
	case tokEOF:
		return nil
	default:
		return preConsumeLine
	}
}

func preConsumeLine(p *symbolScanner) scanStateFn {
	switch p.nextToken.typ {
	case tokNewline:
		return p.consume(preLine)
	case tokError:
		return nil
	case tokEOF:
		return nil
	default:
		return p.consume(preConsumeLine)
	}
}

func preScanValue(p *symbolScanner) scanStateFn {
	for p.nextToken.typ != tokNewline && p.nextToken.typ != tokEOF {
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
	return p.consume(preLine)
}
