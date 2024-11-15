package gmars

import "slices"

// buildReferenceGraph takes a map of expresions and builds a map representing
// a graph of references where each key has a slice of symbols referenced by
// that symbol's tokens
func buildReferenceGraph(values map[string][]token) map[string][]string {
	graph := make(map[string][]string)
	for key, tokens := range values {
		if len(tokens) == 0 {
			continue
		}
		keyRefs := make([]string, 0)
		for _, tok := range tokens {
			if tok.typ != tokText {
				continue
			}
			_, ok := values[tok.val]
			if ok {
				if slices.Contains(keyRefs, tok.val) {
					continue
				}
				keyRefs = append(keyRefs, tok.val)
			}
		}
		graph[key] = keyRefs
	}
	return graph
}

// nodeContainsCycle checks for a cycle in graph by performing a depth first traversal
// recursively, starting from node, and passing the visited nodes to stop if a cycle
// is found
func nodeContainsCycle(node string, graph map[string][]string, visited []string) (bool, string) {
	visited = append(visited, node)

	symRefs, ok := graph[node]
	if !ok {
		return false, ""
	}

	for _, ref := range symRefs {
		if slices.Contains(visited, ref) {
			return true, ref
		}
		subCycle, key := nodeContainsCycle(ref, graph, visited)
		if subCycle {
			return true, key
		}
	}

	return false, ""
}

func graphContainsCycle(graph map[string][]string) (bool, string) {
	for key := range graph {
		nodeCycle, cycleKey := nodeContainsCycle(key, graph, []string{})
		if nodeCycle {
			return true, cycleKey
		}
	}
	return false, ""
}
