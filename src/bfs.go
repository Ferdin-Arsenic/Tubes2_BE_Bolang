package main

import (
	"strings"
)

// bfsShortest finds a single shortest path from basic elements to target
func bfsShortest(elementMap map[string]Element, target string) []string {
	queue := [][]string{}
	visited := make(map[string]bool)

	for _, basic := range basicElements {
		queue = append(queue, []string{basic})
	}

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		node := strings.ToLower(path[len(path)-1])

		if visited[node] {
			continue
		}
		visited[node] = true

		if node == target {
			return path
		}

		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				if a == node || b == node {
					newPath := append([]string{}, path...)
					newPath = append(newPath, name)
					queue = append(queue, newPath)
				}
			}
		}
	}
	return nil
}

// bfsMultiple returns multiple recipe paths in the format required by the tree builder
func bfsMultiple(elementMap map[string]Element, target string, maxRecipe int) []map[string][]string {
	if isBasicElement(target) {
		emptyRecipe := make(map[string][]string)
		return []map[string][]string{emptyRecipe}
	}

	queue := [][]string{}
	visited := make(map[string]bool)
	var results [][]string

	// Start BFS from basic elements
	for _, basic := range basicElements {
		queue = append(queue, []string{basic})
	}

	for len(queue) > 0 && len(results) < maxRecipe {
		path := queue[0]
		queue = queue[1:]
		node := strings.ToLower(path[len(path)-1])

		if visited[node] {
			continue
		}
		visited[node] = true

		if node == target {
			results = append(results, path)
			continue
		}

		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				if a == node || b == node {
					newPath := append([]string{}, path...)
					newPath = append(newPath, name)
					queue = append(queue, newPath)
				}
			}
		}
	}

	// Convert the path results to the recipe map format
	return convertPathsToRecipeMaps(results, target, elementMap)
}

// Convert a list of paths to the recipe map format required by the tree builder
func convertPathsToRecipeMaps(paths [][]string, target string, elementMap map[string]Element) []map[string][]string {
	var recipeMaps []map[string][]string

	for _, path := range paths {
		// Skip if the path doesn't lead to the target
		if len(path) == 0 || strings.ToLower(path[len(path)-1]) != target {
			continue
		}

		recipeMap := make(map[string][]string)

		// Build the recipe map from the path
		for i := 1; i < len(path); i++ {
			currentElem := strings.ToLower(path[i])

			// Find which elements combined to create this element
			for _, recipe := range elementMap[currentElem].Recipes {
				if len(recipe) != 2 {
					continue
				}

				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])

				// Check if either ingredient is in our path
				aInPath := false
				bInPath := false

				for j := 0; j < i; j++ {
					if strings.ToLower(path[j]) == a {
						aInPath = true
					}
					if strings.ToLower(path[j]) == b {
						bInPath = true
					}
				}

				// If both ingredients are in our path or at least one is in our path
				// (this is a simplification, may need more sophisticated logic)
				if aInPath || bInPath {
					recipeMap[currentElem] = []string{a, b}
					break
				}
			}
		}

		// Only add the recipe map if it's not empty
		if len(recipeMap) > 0 {
			recipeMaps = append(recipeMaps, recipeMap)
		}
	}

	return recipeMaps
}
