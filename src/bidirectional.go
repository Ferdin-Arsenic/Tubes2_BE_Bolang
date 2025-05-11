package main

import (
	"strings"
)

// Helper function to build a map of which elements can be created from each element
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

// Helper function to reverse a string slice
func reverseSlice(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// bidirectionalSearch finds a single shortest path from basic elements to target
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
			return append(pathF, reverseSlice(pathB[1:])...)
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
		nodeB := strings.ToLower(pathB[0])

		if pathF, ok := forwardVisited[nodeB]; ok {
			// intersection found
			return append(pathF, reverseSlice(pathB[1:])...)
		}

		for ingredient, elements := range reverseMap {
			for _, elem := range elements {
				if nodeB == strings.ToLower(elem) {
					if _, seen := backwardVisited[ingredient]; !seen {
						newPath := []string{ingredient}
						newPath = append(newPath, pathB...)
						backwardQueue = append(backwardQueue, newPath)
						backwardVisited[ingredient] = newPath
					}
				}
			}
		}
	}

	return nil
}

// bidirectionalMultiple returns multiple recipe paths in the format required by the tree builder
func bidirectionalMultiple(elementMap map[string]Element, target string, maxRecipe int) []TreeNode {
	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}
	}
	targetTier := elementMap[strings.ToLower(target)].Tier

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

	var results [][]string

	for len(forwardQueue) > 0 && len(backwardQueue) > 0 && len(results) < maxRecipe {
		// Expand forward
		pathF := forwardQueue[0]
		forwardQueue = forwardQueue[1:]
		nodeF := strings.ToLower(pathF[len(pathF)-1])

		if pathB, ok := backwardVisited[nodeF]; ok {
			combined := append(pathF, reverseSlice(pathB[1:])...)
			if !containsPath(results, combined) {
				results = append(results, combined)
			}
		}

		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				if elem.Tier >= targetTier {
					continue
				}

				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				if a == nodeF || b == nodeF {
					if _, seen := forwardVisited[strings.ToLower(name)]; !seen {
						newPath := append([]string{}, pathF...)
						newPath = append(newPath, name)
						forwardQueue = append(forwardQueue, newPath)
						forwardVisited[strings.ToLower(name)] = newPath
					}
				}
			}
		}

		if len(results) >= maxRecipe {
			break
		}

		// Expand backward
		pathB := backwardQueue[0]
		backwardQueue = backwardQueue[1:]
		nodeB := strings.ToLower(pathB[0])

		if pathF, ok := forwardVisited[nodeB]; ok {
			combined := append(pathF, reverseSlice(pathB[1:])...)
			if !containsPath(results, combined) {
				results = append(results, combined)
			}
		}

		for ingredient, elements := range reverseMap {
			for _, elem := range elements {
				if elementMap[elem].Tier >= targetTier {
					continue
				}

				if nodeB == strings.ToLower(elem) {
					if _, seen := backwardVisited[strings.ToLower(ingredient)]; !seen {
						newPath := []string{ingredient}
						newPath = append(newPath, pathB...)
						backwardQueue = append(backwardQueue, newPath)
						backwardVisited[strings.ToLower(ingredient)] = newPath
					}
				}
			}
		}
	}

	// 🔁 Konversi ke []TreeNode
	rawMaps := convertPathsToRecipeMaps(results, target, elementMap)
	recipeMaps := []map[string][]string{}
	for _, raw := range rawMaps {
		expandRecipePlan(raw, elementMap, targetTier)
		recipeMaps = append(recipeMaps, raw)
	}

	trees := []TreeNode{}
	for _, recipeMap := range recipeMaps {
		expandRecipePlan(recipeMap, elementMap, targetTier)
		tree := buildRecipeTree(
			strings.ToLower(target),
			recipeMap,
			elementMap,
			make(map[string]bool),
			make(map[string]TreeNode),
		)
		trees = append(trees, tree)
	}

	return trees
}

// Helper function to check if two paths are equal
func pathsAreEqual(path1, path2 []string) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i := 0; i < len(path1); i++ {
		if strings.ToLower(path1[i]) != strings.ToLower(path2[i]) {
			return false
		}
	}
	return true
}

func containsPath(all [][]string, path []string) bool {
	for _, p := range all {
		if pathsAreEqual(p, path) {
			return true
		}
	}
	return false
}
