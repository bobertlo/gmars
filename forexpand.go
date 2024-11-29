package gmars

import (
	"fmt"
	"strings"
)

type forExpander struct {
	lex tokenReader

	// lexing state fields
	nextToken token
	labelBuf  []string
	exprBuf   []token
	atEOF     bool

	// for state fields
	forCount   int
	forIndex   int
	forContent []token

	symbols map[string][]token

	// output fields
	tokens chan token
	closed bool
	err    error
}

type forStateFn func(f *forExpander) forStateFn

func newForExpander(lex tokenReader, symbols map[string][]token) *forExpander {
	f := &forExpander{lex: lex, symbols: symbols}
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
	// add an extra EOF in case we end without one
	// we don't want to block on reading from the channel
	f.tokens <- token{tokEOF, ""}
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
		return f.emitConsume(forConsumeEmitLine)
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
				f.next()
				f.exprBuf = make([]token, 0)
				return forConsumeExpression
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

// forWriteLabelsConsumeLine writes all the stored labels to the token channel,
// emits the current nextToken and returns forConsumeLine
func forWriteLabelsConsumeLine(f *forExpander) forStateFn {
	for _, label := range f.labelBuf {
		f.tokens <- token{tokText, label}
	}
	f.labelBuf = make([]string, 0)
	return f.emitConsume(forConsumeEmitLine)
}

// forConsumeEmitLine consumes and emits tokens until a newline is reached
// the newline is consumed and emitted before calling forLine
func forConsumeEmitLine(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokNewline:
		return f.emitConsume(forLine)
	case tokError:
		return f.emitConsume(nil)
	case tokEOF:
		return f.emitConsume(nil)
	default:
		return f.emitConsume(forConsumeEmitLine)
	}
}

// forConsumeExpressions consumes tokens into the exprBuf until
// a newline is reached then returns forInnerLine after consuming
// the newline to
// newline: forFor
// error: emit, nil
// eof: nil
// otherwise: forConsumeExpression
func forConsumeExpression(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokNewline:
		f.next()
		return forFor
	case tokError:
		return f.emitConsume(nil)
	case tokEOF:
		return nil
	default:
		// f.tokens <- f.nextToken
		f.exprBuf = append(f.exprBuf, f.nextToken)
		f.next()
		return forConsumeExpression
	}
}

// input: exprBuf from forConsumeExpression
// evaluates count expression and
func forFor(f *forExpander) forStateFn {
	expr := make([]token, 0, len(f.exprBuf))
	for _, token := range f.exprBuf {
		if token.typ == tokEOF || token.typ == tokError {
			f.err = fmt.Errorf("unexpected expression term: %s", token)
		}
		expr = append(expr, token)
	}
	f.exprBuf = expr

	val, err := ExpandAndEvaluate(f.exprBuf, f.symbols)
	if err != nil {
		f.tokens <- token{tokError, fmt.Sprintf("%s", err)}
		return nil
	}

	f.forCount = val
	f.forIndex = 0 // should not be necessary
	f.forContent = make([]token, 0)

	return forInnerLine
}

func forInnerLine(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokText:
		return forInnerLabels
	default:
		return nil
	}
}

// this is really just to drop labels before 'rof'
func forInnerLabels(f *forExpander) forStateFn {
	return nil
}
