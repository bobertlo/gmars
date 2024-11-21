package gmars

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

// tokenReader defines an interface shared between the stream based lexer
// and a bufTokenReader to cache tokens in memory.
type tokenReader interface {
	NextToken() (token, error)
	Tokens() ([]token, error)
}

type lexer struct {
	reader   *bufio.Reader
	nextRune rune
	atEOF    bool
	closed   bool
	tokens   chan token
}

// butTokenReader implements the same interface as a streaming parser to let
// us cache and reuse the token stream instead of making multiple passes with
// the lexer
type bufTokenReader struct {
	tokens []token
	i      int
}

func newBufTokenReader(tokens []token) *bufTokenReader {
	return &bufTokenReader{tokens: tokens}
}

func (r *bufTokenReader) NextToken() (token, error) {
	if r.i >= len(r.tokens) {
		return token{}, fmt.Errorf("no more tokens")
	}
	next := r.tokens[r.i]
	r.i++
	return next, nil
}

func (r *bufTokenReader) Tokens() ([]token, error) {
	if r.i >= len(r.tokens) {
		return nil, fmt.Errorf("no more tokens")
	}
	subslice := r.tokens[r.i:]
	ret := make([]token, len(subslice))
	copy(subslice, ret)
	return ret, nil
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

func (l *lexer) consume(nextState lexStateFn) lexStateFn {
	_, eof := l.next()
	if eof {
		l.tokens <- token{tokEOF, ""}
		return nil
	}
	return nextState
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
	case '@':
		fallthrough
	case '{':
		fallthrough
	case '}':
		fallthrough
	case '<':
		fallthrough
	case '>':
		return l.emitConsume(token{tokAddressMode, string(l.nextRune)}, lexInput)
	case ':':
		return l.emitConsume(token{tokColon, ":"}, lexInput)
	case '\x1a':
		return l.consume(lexInput)
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

func LexInput(r io.Reader) ([]token, error) {
	lexer := newLexer(r)
	return lexer.Tokens()
}
