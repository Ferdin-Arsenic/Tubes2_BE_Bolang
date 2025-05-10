package main

import (
	"strings"
)

func bidirectionalSearch(elementMap map[string]Element, target string) []string {
	reverseMap := buildReverseMap(elementMap)

	// forward BFS
	forwardQueue := [][]string{}
	forwardVisited := make(map[string][]string)

	for _, basic := range basicElements {
		forwardQueue = append(forwardQueue, []string{basic})
		forwardVisited[strings.ToLower(basic)] = []string{basic}
	}

	// backward BFS
	backwardQueue := [][]string{{target}}
	backwardVisited := make(map[string][]string)
	backwardVisited[strings.ToLower(target)] = []string{target}

	for len(forwardQueue) > 0 && len(backwardQueue) > 0 {
		// expand forward
		pathF := forwardQueue[0]
		forwardQueue = forwardQueue[1:]
		nodeF := strings.ToLower(pathF[len(pathF)-1])

		if pathB, ok := backwardVisited[nodeF]; ok {
			// intersection found
			return append(pathF, reverseSlice(pathB[:len(pathB)-1])...)
		}

		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				if a == nodeF || b == nodeF {
					if _, seen := forwardVisited[name]; !seen {
						newPath := append([]string{}, pathF...)
						newPath = append(newPath, name)
						forwardQueue = append(forwardQueue, newPath)
						forwardVisited[name] = newPath
					}
				}
			}
		}

		// expand backward
		pathB := backwardQueue[0]
		backwardQueue = backwardQueue[1:]
		nodeB := strings.ToLower(pathB[len(pathB)-1])

		if pathF, ok := forwardVisited[nodeB]; ok {
			// intersection found
			return append(pathF, reverseSlice(pathB[:len(pathB)-1])...)
		}

		for _, parent := range reverseMap[nodeB] {
			if _, seen := backwardVisited[parent]; !seen {
				newPath := append([]string{}, pathB...)
				newPath = append(newPath, parent)
				backwardQueue = append(backwardQueue, newPath)
				backwardVisited[parent] = newPath
			}
		}
	}

	return nil
}

func buildReverseMap(elementMap map[string]Element) map[string][]string {
	reverse := make(map[string][]string)
	for name, elem := range elementMap {
		for _, recipe := range elem.Recipes {
			if len(recipe) == 2 {
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				reverse[a] = append(reverse[a], name)
				reverse[b] = append(reverse[b], name)
			}
		}
	}
	return reverse
}

func reverseSlice(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

func bidirectionalMultiple(elementMap map[string]Element, target string, maxRecipe int) [][]string {
	reverseMap := buildReverseMap(elementMap)

	forwardQueue := [][]string{}
	forwardVisited := make(map[string][]string)

	for _, basic := range basicElements {
		forwardQueue = append(forwardQueue, []string{basic})
		forwardVisited[strings.ToLower(basic)] = []string{basic}
	}

	backwardQueue := [][]string{{target}}
	backwardVisited := make(map[string][]string)
	backwardVisited[strings.ToLower(target)] = []string{target}

	results := [][]string{}

	for len(forwardQueue) > 0 && len(backwardQueue) > 0 && len(results) < maxRecipe {
		// Expand forward
		pathF := forwardQueue[0]
		forwardQueue = forwardQueue[1:]
		nodeF := strings.ToLower(pathF[len(pathF)-1])

		if pathB, ok := backwardVisited[nodeF]; ok {
			combined := append(pathF, reverseSlice(pathB[:len(pathB)-1])...)
			results = append(results, combined)
		}

		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				if a == nodeF || b == nodeF {
					if _, seen := forwardVisited[name]; !seen {
						newPath := append([]string{}, pathF...)
						newPath = append(newPath, name)
						forwardQueue = append(forwardQueue, newPath)
						forwardVisited[name] = newPath
					}
				}
			}
		}

		// Expand backward
		pathB := backwardQueue[0]
		backwardQueue = backwardQueue[1:]
		nodeB := strings.ToLower(pathB[len(pathB)-1])

		if pathF, ok := forwardVisited[nodeB]; ok {
			combined := append(pathF, reverseSlice(pathB[:len(pathB)-1])...)
			results = append(results, combined)
		}

		for _, parent := range reverseMap[nodeB] {
			if _, seen := backwardVisited[parent]; !seen {
				newPath := append([]string{}, pathB...)
				newPath = append(newPath, parent)
				backwardQueue = append(backwardQueue, newPath)
				backwardVisited[parent] = newPath
			}
		}
	}

	return results
}
