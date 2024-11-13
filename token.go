package gmars

import "strings"

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
	_, err := getOpCode(t.val)
	if err == nil {
		return true
	}
	return t.IsPseudoOp()
}

func (t token) NoOperandsOk() bool {
	return strings.ToLower(t.val) == "end"
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

func (t token) IsExpressionTerm() bool {
	if t.typ == tokExprOp || t.typ == tokNumber {
		return true
	}
	return false
}
