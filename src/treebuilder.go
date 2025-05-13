package main

import (
	"strings"
)

func buildRecipeTree(elementName string, recipeSteps map[string][]string, elementMap map[string]Element, visitedInThisTree map[string]bool, memoizedTrees map[string]TreeNode) TreeNode {
	elementName = strings.ToLower(elementName)

	if !visitedInThisTree[elementName] {
		if cachedNode, found := memoizedTrees[elementName]; found {
			return cachedNode
		}
	}

	node := TreeNode{Name: capitalize(elementName)}
	if isBasicElement(elementName) || visitedInThisTree[elementName] {
		if isBasicElement(elementName) && !visitedInThisTree[elementName] {
			memoizedTrees[elementName] = node
		}
		return node
	}
	visitedInThisTree[elementName] = true
	defer delete(visitedInThisTree, elementName)

	parentsToUse, partOfThisSpecificRecipe := recipeSteps[elementName]

	if partOfThisSpecificRecipe && len(parentsToUse) == 2 {
		parent1 := strings.ToLower(parentsToUse[0])
		parent2 := strings.ToLower(parentsToUse[1])

		childNode1 := buildRecipeTree(parent1, recipeSteps, elementMap, visitedInThisTree, memoizedTrees)
		childNode2 := buildRecipeTree(parent2, recipeSteps, elementMap, visitedInThisTree, memoizedTrees)

		node.Children = append(node.Children, childNode1)
		node.Children = append(node.Children, childNode2)

	}
	memoizedTrees[elementName] = node
	return node
}