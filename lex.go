package gmars

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type tokenType uint8

const (
	tokError       tokenType = iota // returned when an error is encountered
	tokText                         // used for labels, symbols, and opcodes
	tokAddressMode                  // $ # { } < >
	tokNumber                       // (optionally) signed integer
	tokExprOp                       // + - * / % ==
	tokComma
	tokParenL
	tokParenR
	tokComment // includes semi-colon, no newline char
	tokNewline
	tokEOF
)

type token struct {
	typ tokenType
	val string
}

func (t token) String() string {
	switch t.typ {
	case tokEOF:
		return "EOF"
	case tokNewline:
		return "newline"
	default:
		return t.val
	}
}

func (t token) IsOp() bool {
	if t.typ != tokText {
		return false
	}
	if strings.Contains(t.val, ".") {
		return true
	}
	return t.IsPseudoOp()
}

func (t token) IsPseudoOp() bool {
	switch strings.ToLower(t.val) {
	case "end":
		return true
	case "equ":
		return true
	case "org":
		return true
	case "for":
		return true
	case "rof":
		return true
	default:
		return false
	}
}

type lexer struct {
	reader   *bufio.Reader
	nextRune rune
	atEOF    bool
	closed   bool
	tokens   chan token
}

type lexStateFn func(l *lexer) lexStateFn

func newLexer(r io.Reader) *lexer {
	lex := &lexer{
		reader: bufio.NewReader(r),
		tokens: make(chan token),
	}
	lex.next()
	go lex.run()
	return lex
}

func (l *lexer) next() (rune, bool) {
	if l.atEOF {
		return '\x00', true
	}

	r, _, err := l.reader.ReadRune()
	if err != nil {
		l.atEOF = true
		return l.nextRune, true
	}

	lastRune := l.nextRune
	l.nextRune = r
	return lastRune, false
}

func (l *lexer) run() {
	for state := lexInput; state != nil; {
		state = state(l)
	}
	close(l.tokens)
	l.closed = true
}

func (l *lexer) NextToken() (token, error) {
	if l.closed {
		return token{}, fmt.Errorf("no more tokens")
	}
	return <-l.tokens, nil
}

func (l *lexer) Tokens() ([]token, error) {
	tokens := make([]token, 0)
	for {
		token, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
		if token.typ == tokEOF || token.typ == tokError {
			break
		}
	}
	return tokens, nil
}

func (l *lexer) emitConsume(tok token, nextState lexStateFn) lexStateFn {
	l.tokens <- tok
	_, eof := l.next()
	if eof {
		l.tokens <- token{tokEOF, ""}
		return nil
	}
	return nextState
}

func lexInput(l *lexer) lexStateFn {
	// consume any space until non-space characters, emitting tokNewlines
	if unicode.IsSpace(l.nextRune) {
		for unicode.IsSpace(l.nextRune) {
			if l.nextRune == '\n' {
				l.tokens <- token{typ: tokNewline}
			}
			_, eof := l.next()
			if eof {
				l.tokens <- token{typ: tokEOF}
				return nil
			}
		}
		return lexInput
	}

	// handle alphanumeric input
	if unicode.IsLetter(l.nextRune) || l.nextRune == '_' {
		return lexText
	}

	if unicode.IsDigit(l.nextRune) {
		return lexNumber
	}

	// handle comments
	if l.nextRune == ';' {
		return lexComment
	}

	// dispatch based on next rune, or error
	switch l.nextRune {
	case '\x00':
		l.tokens <- token{tokEOF, ""}
	case ',':
		return l.emitConsume(token{tokComma, ","}, lexInput)
	case '(':
		return l.emitConsume(token{tokParenL, "("}, lexInput)
	case ')':
		return l.emitConsume(token{tokParenR, ")"}, lexInput)
	case '+':
		fallthrough
	case '-':
		fallthrough
	case '*':
		fallthrough
	case '/':
		fallthrough
	case '%':
		return l.emitConsume(token{tokExprOp, string(l.nextRune)}, lexInput)
	case '$':
		fallthrough
	case '#':
		fallthrough
	case '{':
		fallthrough
	case '}':
		fallthrough
	case '<':
		fallthrough
	case '>':
		return l.emitConsume(token{tokAddressMode, string(l.nextRune)}, lexInput)
	default:
		l.tokens <- token{tokError, fmt.Sprintf("unexpected character: '%s'", string(l.nextRune))}
	}

	return nil
}

func lexText(l *lexer) lexStateFn {
	runeBuf := make([]rune, 0, 10)

	for unicode.IsLetter(l.nextRune) || unicode.IsDigit(l.nextRune) || l.nextRune == '.' || l.nextRune == '_' {
		r, eof := l.next()
		runeBuf = append(runeBuf, r)
		if eof {
			l.tokens <- token{typ: tokText, val: string(runeBuf)}
			l.tokens <- token{typ: tokEOF}
			return nil
		}
	}

	if len(runeBuf) > 0 {
		l.tokens <- token{typ: tokText, val: string(runeBuf)}
	}

	return lexInput
}

func lexNumber(l *lexer) lexStateFn {
	numberBuf := make([]rune, 0, 10)
	for unicode.IsDigit(l.nextRune) {
		r, eof := l.next()
		numberBuf = append(numberBuf, r)
		if eof {
			l.tokens <- token{tokNumber, string(numberBuf)}
			l.tokens <- token{typ: tokEOF}
			return nil
		}
	}

	if len(numberBuf) > 0 {
		l.tokens <- token{tokNumber, string(numberBuf)}
	}

	return lexInput
}

func lexComment(l *lexer) lexStateFn {
	commentBuf := make([]rune, 0, 32)

	for l.nextRune != '\n' {
		commentBuf = append(commentBuf, l.nextRune)
		_, eof := l.next()
		if eof {
			l.tokens <- token{tokComment, string(commentBuf)}
			l.tokens <- token{tokEOF, ""}
			return nil
		}
	}
	l.tokens <- token{typ: tokComment, val: string(commentBuf)}
	return lexInput
}
