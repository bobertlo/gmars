package gmars

import (
	"fmt"
	gtoken "go/token"
	"go/types"
	"strconv"
)

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

func evaluateExpression(expr []token) (int, error) {
	exprStr := ""
	for _, tok := range expr {
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
