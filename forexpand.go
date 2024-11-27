package gmars

import (
	"fmt"
	"strings"
)

type forExpander struct {
	lex tokenReader

	nextToken token
	labelBuf  []string
	// valueBuf  []token
	atEOF bool

	tokens chan token
	closed bool
	err    error
}

type forStateFn func(f *forExpander) forStateFn

func newForExpander(lex tokenReader) *forExpander {
	f := &forExpander{lex: lex}
	f.next()
	f.tokens = make(chan token)
	go f.run()
	return f
}

func (p *forExpander) next() token {
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

func (f *forExpander) run() {
	if f.closed || f.atEOF {
		return
	}
	for state := forLine; state != nil; {
		state = state(f)
	}
	f.closed = true
}

func (f *forExpander) NextToken() (token, error) {
	if f.closed {
		return token{}, fmt.Errorf("no more tokens")
	}
	return <-f.tokens, nil
}

func (f *forExpander) Tokens() ([]token, error) {
	if f.closed {
		return nil, fmt.Errorf("no more tokens")
	}
	tokens := make([]token, 0)
	for !f.closed {
		tok := <-f.tokens
		tokens = append(tokens, tok)
		if tok.typ == tokEOF || tok.typ == tokError {
			break
		}
	}
	return tokens, nil
}

func (f *forExpander) emitConsume(nextState forStateFn) forStateFn {
	f.tokens <- f.nextToken
	f.next()
	return nextState
}

// forLine is the base state and returned to after every newline outside a for loop
// text: forConsumeLabels
// anything else: forConsumeLine
func forLine(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokText:
		f.labelBuf = make([]string, 0)
		return forConsumeLabels
	default:
		return f.emitConsume(forConsumeLine)
	}
}

// consume labels into labelBuf and go to next state
// text "for": forFor
// text op/pseudo: forWriteLabelsConsumeLine
// text other: append to labelBuf, forConsumeLabels
// newline/comment: forConsumeLabels
// other: nil
func forConsumeLabels(f *forExpander) forStateFn {
	if f.nextToken.typ == tokText {

		if f.nextToken.IsPseudoOp() {
			opLower := strings.ToLower(f.nextToken.val)
			if opLower == "for" {
				return forFor
			} else {
				return forWriteLabelsConsumeLine
			}
		} else if f.nextToken.IsOp() {
			return forWriteLabelsConsumeLine
		} else {
			f.labelBuf = append(f.labelBuf, f.nextToken.val)
			f.next()
			return forConsumeLabels
		}
	} else if f.nextToken.typ == tokNewline || f.nextToken.typ == tokComment {
		f.next()
		return forConsumeLabels
	} else {
		f.err = fmt.Errorf("expected label, op, newlines, or comment, got '%s'", f.nextToken)
		return nil
	}
}

func forWriteLabelsConsumeLine(f *forExpander) forStateFn {
	for _, label := range f.labelBuf {
		f.tokens <- token{tokText, label}
	}
	f.labelBuf = make([]string, 0)
	return f.emitConsume(forConsumeLine)
}

func forConsumeLine(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokNewline:
		return f.emitConsume(forLine)
	case tokError:
		return f.emitConsume(nil)
	case tokEOF:
		return f.emitConsume(nil)
	default:
		return f.emitConsume(forConsumeLine)
	}
}

func forFor(f *forExpander) forStateFn {
	// if len(f.labelBuf) == 0 {

	// }
	return nil
}

func forInnerLine(f *forExpander) forStateFn {
	return nil
}
