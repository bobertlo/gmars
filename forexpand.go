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
	forCountLabel        string
	forLineLabels        []string
	forLineLabelsToWrite []string
	forCount             int
	forIndex             int
	forContent           []token
	forDepth             int

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

func ForExpand(lex tokenReader, symbols map[string][]token) ([]token, error) {
	expander := newForExpander(lex, symbols)
	tokens, err := expander.Tokens()
	if err != nil {
		return nil, err
	}
	return tokens, nil
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
		return forConsumeEmitLine
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
				return forWriteLabelsEmitConsumeLine
			}
		} else if f.nextToken.IsOp() {
			return forWriteLabelsEmitConsumeLine
		} else {
			f.labelBuf = append(f.labelBuf, f.nextToken.val)
			f.next()
			return forConsumeLabels
		}
	} else if f.nextToken.typ == tokNewline || f.nextToken.typ == tokComment || f.nextToken.typ == tokColon {
		f.next()
		return forConsumeLabels
	} else {
		f.err = fmt.Errorf("expected label, op, newlines, or comment, got '%s'", f.nextToken)
		return nil
	}
}

// forWriteLabelsEmitConsumeLine writes all the stored labels to the token channel,
// emits the current nextToken and returns forConsumeLine
func forWriteLabelsEmitConsumeLine(f *forExpander) forStateFn {
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
	case tokComment:
		f.next()
		return forConsumeExpression
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
// evaluates count expression and sets up for state
// always returns forInnerLine or Error
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

	if len(f.labelBuf) > 0 {
		f.forCountLabel = f.labelBuf[len(f.labelBuf)-1]

		if len(f.labelBuf) > 1 {
			f.forLineLabels = f.labelBuf[:len(f.labelBuf)-1]
		} else {
			f.forLineLabels = []string{}
		}
	} else {
		f.forCountLabel = ""
		f.forLineLabels = []string{}
	}

	f.forLineLabelsToWrite = make([]string, len(f.forLineLabels))
	for i, label := range f.forLineLabels {
		f.forLineLabelsToWrite[i] = fmt.Sprintf("__for_%s_%s", f.forCountLabel, label)
	}

	f.forCount = val
	f.forIndex = 0 // should not be necessary
	f.forContent = make([]token, 0)
	f.labelBuf = make([]string, 0)

	return forInnerLine
}

// text: forInnerConsumeLabels
// other: forInnerConsumeLine
func forInnerLine(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokText:
		f.labelBuf = make([]string, 0)
		return forInnerLabels
	default:
		// emitconsume line into for buffer
		return forInnerEmitConsumeLine
	}
}

// this is really just to drop labels before 'rof'
func forInnerLabels(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokText:
		if f.nextToken.IsPseudoOp() {
			opLower := strings.ToLower(f.nextToken.val)
			if opLower == "for" {
				f.forDepth += 1
				return forInnerEmitLabels
			} else if opLower == "rof" {
				if f.forDepth > 0 {
					f.forDepth -= 1
					return forInnerEmitConsumeLine
				} else {
					return forRof
				}
			} else {
				return forInnerEmitLabels
			}
		} else if f.nextToken.IsOp() {
			if f.forLineLabelsToWrite != nil {
				for _, label := range f.forLineLabelsToWrite {
					f.tokens <- token{tokText, label}
				}
				f.forLineLabelsToWrite = nil
			}
			return forInnerEmitLabels
		} else {
			f.labelBuf = append(f.labelBuf, f.nextToken.val)
			f.next()
			return forInnerLabels
		}
	default:
		// not expecting legal input here, but we will let the parser deal with it
		return forInnerEmitLabels
	}
}

func forInnerEmitLabels(f *forExpander) forStateFn {
	for _, label := range f.labelBuf {
		f.forContent = append(f.forContent, token{tokText, label})
	}
	return forInnerEmitConsumeLine
}

func forInnerEmitConsumeLine(f *forExpander) forStateFn {
	switch f.nextToken.typ {
	case tokError:
		f.tokens <- f.nextToken
		return nil
	case tokEOF:
		return nil
	case tokNewline:
		f.forContent = append(f.forContent, f.nextToken)
		f.next()
		return forInnerLine
	default:
		f.forContent = append(f.forContent, f.nextToken)
		f.next()
		return forInnerEmitConsumeLine
	}
}

func forRof(f *forExpander) forStateFn {
	for f.nextToken.typ != tokNewline {
		if f.nextToken.typ == tokEOF || f.nextToken.typ == tokError {
			f.tokens <- f.nextToken
			return nil
		}
		f.next()
	}
	f.next()

	for i := 1; i <= f.forCount; i++ {
		for _, tok := range f.forContent {
			if tok.typ == tokText {
				if tok.val == f.forCountLabel {
					f.tokens <- token{tokNumber, fmt.Sprintf("%d", i)}
				} else {
					found := false
					for _, label := range f.forLineLabels {
						forLabel := fmt.Sprintf("__for_%s_%s", f.forCountLabel, label)
						if tok.val == label {
							f.tokens <- token{tokText, forLabel}
							found = true
							break
						}
					}
					if !found {
						f.tokens <- tok
					}
				}
			} else {
				f.tokens <- tok
			}
		}
	}

	return forEmitConsumeStream
}

func forEmitConsumeStream(f *forExpander) forStateFn {
	for f.nextToken.typ != tokEOF {
		f.tokens <- f.nextToken
		f.next()
	}
	return nil
}
