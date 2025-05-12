package main

import (
	"strings"
)

// Function to build a reverse mapping from ingredients to the elements they create
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

// Helper function to check if a path is already in a slice of paths
func containsPath(all [][]string, path []string) bool {
	for _, p := range all {
		if pathsAreEqual(p, path) {
			return true
		}
	}
	return false
}

// PathIntersection represents the meeting point of forward and backward searches
type PathIntersection struct {
	ForwardPath  []string
	BackwardPath []string
}

// bidirectionalSearchMultiple finds multiple paths from basic elements to target
func bidirectionalSearchMultiple(elementMap map[string]Element, target string, maxPaths int) [][]string {
	target = strings.ToLower(target)
	reverseMap := buildReverseMap(elementMap)

	// Use [][][]string to allow multiple paths per node
	forwardVisited := make(map[string][][]string)
	backwardVisited := make(map[string][][]string)

	forwardQueue := [][]string{}
	backwardQueue := [][]string{{target}}

	for _, basic := range basicElements {
		path := []string{basic}
		forwardQueue = append(forwardQueue, path)
		forwardVisited[strings.ToLower(basic)] = append(forwardVisited[strings.ToLower(basic)], path)
	}

	backwardVisited[target] = append(backwardVisited[target], []string{target})

	maxDepth := 17
	var intersections []PathIntersection

	forwardDepth, backwardDepth := 0, 0
	currForwardSize := len(forwardQueue)
	currBackwardSize := len(backwardQueue)

	for len(forwardQueue) > 0 && len(backwardQueue) > 0 &&
		forwardDepth < maxDepth && backwardDepth < maxDepth {

		// Expand forward
		for i := 0; i < currForwardSize; i++ {
			pathF := forwardQueue[0]
			forwardQueue = forwardQueue[1:]
			node := strings.ToLower(pathF[len(pathF)-1])

			// If backward already visited node, record intersection
			if pathsB, ok := backwardVisited[node]; ok {
				for _, pathB := range pathsB {
					intersections = append(intersections, PathIntersection{
						ForwardPath:  pathF,
						BackwardPath: pathB,
					})
					if maxPaths > 0 && len(intersections) >= maxPaths {
						goto build
					}
				}
			}

			expandForwardMulti(node, pathF, elementMap, forwardVisited, &forwardQueue)
		}
		forwardDepth++
		currForwardSize = len(forwardQueue)

		// Expand backward
		for i := 0; i < currBackwardSize; i++ {
			pathB := backwardQueue[0]
			backwardQueue = backwardQueue[1:]
			node := strings.ToLower(pathB[0])

			if pathsF, ok := forwardVisited[node]; ok {
				for _, pathF := range pathsF {
					intersections = append(intersections, PathIntersection{
						ForwardPath:  pathF,
						BackwardPath: pathB,
					})
					if maxPaths > 0 && len(intersections) >= maxPaths {
						goto build
					}
				}
			}
			expandBackwardMulti(node, pathB, reverseMap, elementMap, backwardVisited, &backwardQueue)
		}
		backwardDepth++
		currBackwardSize = len(backwardQueue)
	}

build:
	var completePaths [][]string
	for _, intersection := range intersections {
		merged := append([]string{}, intersection.ForwardPath...)
		merged = append(merged, reverseSlice(intersection.BackwardPath[1:])...)
		if !containsPath(completePaths, merged) {
			completePaths = append(completePaths, merged)
		}
		if maxPaths > 0 && len(completePaths) >= maxPaths {
			break
		}
	}
	return completePaths
}

// Helper function to expand forward search
func expandForward(node string, path []string,
	elementMap map[string]Element,
	visited map[string][]string,
	queue *[][]string) {

	for name, elem := range elementMap {
		for _, recipe := range elem.Recipes {
			if len(recipe) != 2 {
				continue
			}

			a := strings.ToLower(recipe[0])
			b := strings.ToLower(recipe[1])

			// If this node is an ingredient in the recipe
			if a == node || b == node {
				// Skip cycles and paths causing tier violations
				if _, seen := visited[name]; !seen {
					// Add result element to path
					newPath := append([]string{}, path...)
					newPath = append(newPath, name)
					*queue = append(*queue, newPath)
					visited[name] = newPath
				}
			}
		}
	}
}

// Helper function to expand backward search
func expandBackward(node string, path []string,
	reverseMap map[string][]string,
	visited map[string][]string,
	queue *[][]string) {

	// Find all elements that can create this node
	for ingredient, elements := range reverseMap {
		for _, elem := range elements {
			if node == strings.ToLower(elem) {
				// Skip if already visited
				if _, seen := visited[ingredient]; !seen {
					// Add ingredient to path
					newPath := []string{ingredient}
					newPath = append(newPath, path...)
					*queue = append(*queue, newPath)
					visited[ingredient] = newPath
				}
			}
		}
	}
}

func expandForwardMulti(node string, path []string,
	elementMap map[string]Element,
	visited map[string][][]string,
	queue *[][]string) {

	for name, elem := range elementMap {
		for _, recipe := range elem.Recipes {
			if len(recipe) != 2 {
				continue
			}
			a := strings.ToLower(recipe[0])
			b := strings.ToLower(recipe[1])

			if a != node && b != node {
				continue
			}

			// ❗ Validasi: pastikan bahan-bahan tier-nya lebih rendah dari hasil
			ae, aok := elementMap[a]
			be, bok := elementMap[b]
			if !aok || !bok || ae.Tier >= elem.Tier || be.Tier >= elem.Tier {
				continue
			}

			newPath := append([]string{}, path...)
			newPath = append(newPath, name)
			visited[name] = append(visited[name], newPath)
			*queue = append(*queue, newPath)
		}
	}
}

func expandBackwardMulti(node string, path []string,
	reverseMap map[string][]string,
	elementMap map[string]Element,
	visited map[string][][]string,
	queue *[][]string) {

	for ingredient, products := range reverseMap {
		for _, product := range products {
			if node != strings.ToLower(product) {
				continue
			}
			ie, iok := elementMap[ingredient]
			pe, pok := elementMap[product]
			if !iok || !pok || ie.Tier >= pe.Tier {
				continue // ❗ hanya tambahkan jika tier-nya valid
			}

			newPath := append([]string{ingredient}, path...)
			visited[ingredient] = append(visited[ingredient], newPath)
			*queue = append(*queue, newPath)
		}
	}
}

// Enhanced bidirectionalMultiple with multiple recipe paths
func bidirectionalMultiple(elementMap map[string]Element, target string, maxRecipes int) []TreeNode {
	target = strings.ToLower(target)

	// Handle basic elements directly
	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}
	}

	// Check if target exists in element map
	targetElem, exists := elementMap[target]
	if !exists || len(targetElem.Recipes) == 0 {
		return []TreeNode{}
	}

	// Store all valid recipes for the target
	var allResults []TreeNode
	targetTier := targetElem.Tier

	// Find multiple paths using bidirectional search
	paths := bidirectionalSearchMultiple(elementMap, target, maxRecipes)
	if len(paths) == 0 {
		return []TreeNode{}
	}

	// Process each path into a recipe tree
	for _, path := range paths {
		// Create recipe map from this path
		recipeMap := make(map[string][]string)

		// Process the path from end to beginning to build recipe map
		for i := len(path) - 1; i > 0; i-- {
			child := strings.ToLower(path[i])
			parent := strings.ToLower(path[i-1])

			// Find matching recipe
			elem, exists := elementMap[child]
			if !exists {
				continue
			}

			// Find a recipe that includes the parent
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}

				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])

				// If parent is part of this recipe
				if a == parent || b == parent {
					// Validasi tier
					ae, aok := elementMap[a]
					be, bok := elementMap[b]
					if !aok || !bok || ae.Tier >= elem.Tier || be.Tier >= elem.Tier {
						continue
					}
					// Find the other ingredient
					otherIngredient := b
					if a != parent {
						otherIngredient = a
					}

					// Add to recipe map
					recipeMap[child] = []string{parent, otherIngredient}
					break
				}
			}
		}

		// Expand recipe plan to include all necessary elements
		expandRecipePlan(recipeMap, elementMap, targetTier)

		// Build tree from this recipe map
		tree := buildRecipeTree(
			target,
			recipeMap,
			elementMap,
			make(map[string]bool),
			make(map[string]TreeNode),
		)

		// Avoid duplicate trees
		isDuplicate := false
		for _, existing := range allResults {
			if canonicalizeTree(existing) == canonicalizeTree(tree) {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			allResults = append(allResults, tree)
		}

		// Respect the max recipes limit
		if maxRecipes > 0 && len(allResults) >= maxRecipes {
			break
		}
	}

	return allResults
}
