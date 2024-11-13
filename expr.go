package gmars

type nodeType uint8

const (
	nodeLiteral nodeType = iota
	nodeSymbol
	nodeOp // + - * / %
)

type expNode struct {
	typ    nodeType
	symbol string
	value  int
	a      *expNode
	b      *expNode
}

type expression struct {
	tokens []token
	root   *expNode
}

func newExpression(t []token) *expression {
	return &expression{tokens: t}
}

func (e *expression) AppendToken(t token) {
	e.tokens = append(e.tokens, t)
}
