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

func LexInput(r io.Reader) ([]token, error) {
	lexer := newLexer(r)
	return lexer.Tokens()
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

	// dispatch based on next rune, or error
	switch l.nextRune {
	case '\x00':
		l.tokens <- token{tokEOF, ""}
	case ';':
		return lexComment
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
		return l.emitConsume(token{tokSymbol, string(l.nextRune)}, lexInput)
	case '$':
		fallthrough
	case '#':
		fallthrough
	case '@':
		fallthrough
	case '{':
		fallthrough
	case '}':
		return l.emitConsume(token{tokSymbol, string(l.nextRune)}, lexInput)
	case '<':
		return l.consume(lexLt)
	case '>':
		return l.consume(lexGt)
	case ':':
		return l.emitConsume(token{tokColon, ":"}, lexInput)
	case '=':
		return l.consume(lexEquals)
	case '|':
		return l.consume(lexPipe)
	case '&':
		return l.consume(lexAnd)
	case '\x1a':
		return l.consume(lexInput)
	default:
		// we will put this in the stream. if a file is formatted
		// properly, and invalid input should be after an 'end'
		// pseudo-op which will cause the parser to stop before
		// processing this token, otherwise it is an error
		l.tokens <- token{tokInvalid, string(l.nextRune)}
		l.tokens <- token{typ: tokEOF}
		return nil
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

	// // remove leading zeros
	for l.nextRune == '0' {
		_, eof := l.next()
		if eof {
			l.tokens <- token{tokNumber, "0"}
			l.tokens <- token{typ: tokEOF}
			return nil
		}
	}

	for unicode.IsDigit(l.nextRune) {
		r, eof := l.next()
		numberBuf = append(numberBuf, r)
		if eof {
			l.tokens <- token{tokNumber, string(numberBuf)}
			l.tokens <- token{typ: tokEOF}
			return nil
		}
	}

	if len(numberBuf) == 0 {
		l.tokens <- token{tokNumber, "0"}
		return lexInput
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

func lexEquals(l *lexer) lexStateFn {
	if l.nextRune == '=' {
		return l.emitConsume(token{tokSymbol, "=="}, lexInput)
	} else {
		l.tokens <- token{tokError, fmt.Sprintf("expected '=' after '=', got '%s'", string(l.nextRune))}
		return nil
	}
}

func lexPipe(l *lexer) lexStateFn {
	if l.nextRune == '|' {
		return l.emitConsume(token{tokSymbol, "||"}, lexInput)
	} else {
		l.tokens <- token{tokError, fmt.Sprintf("expected '|' after '|', got '%s'", string(l.nextRune))}
		return nil
	}
}

func lexAnd(l *lexer) lexStateFn {
	if l.nextRune == '&' {
		return l.emitConsume(token{tokSymbol, "&&"}, lexInput)
	} else {
		l.tokens <- token{tokError, fmt.Sprintf("expected '&' after '&', got '%s'", string(l.nextRune))}
		return nil
	}
}

func lexGt(l *lexer) lexStateFn {
	if l.nextRune == '=' {
		return l.emitConsume(token{tokSymbol, ">="}, lexInput)
	} else {
		l.tokens <- token{tokSymbol, ">"}
		return lexInput
	}
}

func lexLt(l *lexer) lexStateFn {
	if l.nextRune == '=' {
		return l.emitConsume(token{tokSymbol, "<="}, lexInput)
	} else {
		l.tokens <- token{tokSymbol, "<"}
		return lexInput
	}
}
