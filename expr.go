package gmars

import (
	"fmt"
	gtoken "go/token"
	"go/types"
	"strconv"
)

func ExpandAndEvaluate(expr []token, symbols map[string][]token) (int, error) {
	graph := buildReferenceGraph(symbols)

	cyclic, key := graphContainsCycle(graph)
	if cyclic {
		return 0, fmt.Errorf("symbol graph contains cycles: %s", key)
	}

	resolved, err := expandExpressions(symbols, graph)
	if err != nil {
		return 0, err
	}

	expanded := make([]token, 0)
	for _, tok := range expr {
		if tok.typ == tokText {
			symVal, ok := resolved[tok.val]
			if ok {
				expanded = append(expanded, symVal...)
				continue
			}
		}
		expanded = append(expanded, tok)
	}

	return evaluateExpression(expanded)
}

func expandValue(key string, values, resolved map[string][]token, graph map[string][]string) ([]token, error) {
	// load key value or error
	value, valOk := values[key]
	if !valOk {
		return nil, fmt.Errorf("symbol '%s' key not found", key)
	}

	// return resolved value if exists. on principle
	if res, ok := resolved[key]; ok {
		return res, nil
	}

	// recursively expand dependent values if not already resolved
	deps, ok := graph[key]
	if ok {
		for _, dep := range deps {
			_, resOk := resolved[dep]
			if !resOk {
				_, err := expandValue(dep, values, resolved, graph)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// create new token slice and append tokens from the symbol value
	// while replacing reference tokens with their resolved values
	output := make([]token, 0)
	for _, token := range value {
		if token.typ == tokText {
			depVal, depOk := resolved[token.val]
			if depOk {
				// variable names will be resolved
				output = append(output, depVal...)
			} else {
				// otherwise it is a label
				output = append(output, token)
			}
		} else {
			output = append(output, token)
		}
	}

	resolved[key] = output

	return output, nil
}

func expandExpressions(values map[string][]token, graph map[string][]string) (map[string][]token, error) {
	resolved := make(map[string][]token)

	for key := range values {
		_, ok := resolved[key]
		if ok {
			continue
		}
		expanded, err := expandValue(key, values, resolved, graph)
		if err != nil {
			return nil, err
		}
		resolved[key] = expanded
	}
	return resolved, nil
}

func combineSigns(expr []token) []token {
	out := make([]token, 0, len(expr))
	lastOut := token{typ: tokEOF}

	// please forgive me for this lol
	for i := 0; i < len(expr); i++ {
		if lastOut.typ == tokSymbol {
			negativeFound := false
			for ; i < len(expr); i++ {
				if !(expr[i].val == "-" || expr[i].val == "+") {
					break
				}
				if expr[i].val == "-" {
					negativeFound = true
				}
			}
			if negativeFound {
				out = append(out, token{tokSymbol, "-"})
			}
			if i < len(expr) {
				out = append(out, expr[i])
				lastOut = expr[i]
			}
		} else {
			if i < len(expr) {
				out = append(out, expr[i])
				lastOut = expr[i]
			}
		}
	}
	return out
}

func flipDoubleNegatives(expr []token) []token {
	out := make([]token, 0, len(expr))
	for i := 0; i < len(expr); i++ {
		if expr[i].val == "-" {
			if i+1 < len(expr) && expr[i+1].val == "-" {
				out = append(out, token{tokSymbol, "+"})
				i += 1
				continue
			}
		}
		out = append(out, expr[i])
	}
	return out
}

func evaluateExpression(expr []token) (int, error) {
	for _, tok := range expr {
		if tok.typ == tokText || !tok.IsExpressionTerm() {
			return 0, fmt.Errorf("unexpected token in expression: '%s'", tok)
		}
	}

	combinedExpr := combineSigns(expr)
	flippedExpr := flipDoubleNegatives(combinedExpr)

	exprStr := ""
	for _, tok := range flippedExpr {
		exprStr += tok.val
	}

	fs := gtoken.NewFileSet()
	tv, err := types.Eval(fs, nil, gtoken.NoPos, exprStr)
	if err != nil {
		return 0, err
	}

	valStr := tv.Value.String()
	if valStr == "true" {
		return 1, nil
	} else if valStr == "false" {
		return 0, nil
	}

	val, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func exprEqual(a, b []token) bool {
	if len(a) != len(b) {
		return false
	}
	for i, aval := range a {
		if b[i] != aval {
			return false
		}
	}
	return true
}
