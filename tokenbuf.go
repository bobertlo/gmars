package gmars

import "fmt"

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
