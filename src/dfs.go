package main

import "strings"

func dfsMultiple(elementMap map[string]Element, basicElements []string, target string, maxPaths int) [][]string {
	foundPaths := [][]string{}
	targetNormalized := strings.ToLower(target)

	for _, startElement := range basicElements {
		visitedOnCurrentPath := make(map[string]bool)
		initialPath := []string{startElement}

		dfsHelper(elementMap, startElement, targetNormalized, visitedOnCurrentPath, initialPath, &foundPaths, maxPaths)
		if len(foundPaths) >= maxPaths {break}
	}
	return foundPaths
}

// Recursive DFS helper function
func dfsHelper(elementMap map[string]Element, currentElement string, targetNode string, visitedOnCurrentPath map[string]bool, currentPathInProgress []string, foundPaths *[][]string, maxPaths int) {
	if len(*foundPaths) >= maxPaths {return}

	currentElementLower := strings.ToLower(currentElement)

	if visitedOnCurrentPath[currentElementLower] {return}

	visitedOnCurrentPath[currentElementLower] = true
	
	defer func() { delete(visitedOnCurrentPath, currentElementLower) }()

	if currentElementLower == targetNode {
		pathCopy := make([]string, len(currentPathInProgress))
		copy(pathCopy, currentPathInProgress)
		*foundPaths = append(*foundPaths, pathCopy)
		return
	}

	for potentialNextProduct, details := range elementMap {
		if len(*foundPaths) >= maxPaths {
			return
		}

		for _, recipe := range details.Recipes {
			if len(recipe) != 2 {
				continue
			}
			ing1 := strings.ToLower(recipe[0])
			ing2 := strings.ToLower(recipe[1])

			if ing1 == currentElementLower || ing2 == currentElementLower {
				nextPath := append([]string{}, currentPathInProgress...)
				nextPath = append(nextPath, potentialNextProduct)

				dfsHelper( elementMap, potentialNextProduct, targetNode, visitedOnCurrentPath, nextPath, foundPaths, maxPaths)

				if len(*foundPaths) >= maxPaths {return}
				break
			}
		}
	}
}