package main

import "strings"

type DFSData struct {
	elementMap    map[string]Element
	initialTarget string
	maxRecipes    int
	cache map[string][]TreeNode
	nodeCounter int
}

func copyRecipe(originalMap map[string][]string) map[string][]string {
	newMap := make(map[string][]string, len(originalMap))
	for k, v := range originalMap {
		parentsCopy := make([]string, len(v))
		copy(parentsCopy, v)
		newMap[k] = parentsCopy
	}
	return newMap
}

func dfsMultiple(elementMap map[string]Element, target string, maxRecipes int) ([]TreeNode, int) {
	dfsData := DFSData{
		elementMap:    elementMap,
		initialTarget: strings.ToLower(target),
		maxRecipes:    maxRecipes,
		cache:         make(map[string][]TreeNode), // Cache stores []TreeNode
		nodeCounter: 0,
	}

	resultTreeNodes := dfsData.dfsRecursive(strings.ToLower(target))

	// Optional: Apply maxRecipes limit to the final list of trees
	// Your external de-duplication might happen before or after this.
	if maxRecipes > 0 && len(resultTreeNodes) > maxRecipes {
		return resultTreeNodes[:maxRecipes], dfsData.nodeCounter
	}
	return resultTreeNodes, dfsData.nodeCounter
}

func (d *DFSData) dfsRecursive(elementToMakeCurrently string) []TreeNode {
	d.nodeCounter++
	elementToMakeCurrently = strings.ToLower(elementToMakeCurrently)

	if cachedResult, found := d.cache[elementToMakeCurrently]; found {
		return cachedResult
	}

	elemDetails, exists := d.elementMap[elementToMakeCurrently]
	if !exists {
		d.cache[elementToMakeCurrently] = []TreeNode{}
		return []TreeNode{}
	}

	currentElementNameFormatted := elemDetails.Name


	if isBasicElement(elemDetails.Name) {
		leafNode := TreeNode{Name: currentElementNameFormatted}
		basicTreeList := []TreeNode{leafNode}
		d.cache[elementToMakeCurrently] = basicTreeList
		return basicTreeList
	}

	if len(elemDetails.Recipes) == 0 {
		leafNode := TreeNode{Name: currentElementNameFormatted}
		noRecipeTreeList := []TreeNode{leafNode}
		d.cache[elementToMakeCurrently] = noRecipeTreeList
		return noRecipeTreeList
	}

	var operationalLimit int
	isInitialTarget := (d.initialTarget == elementToMakeCurrently)

	if d.maxRecipes <= 0 {
		operationalLimit = 0
	} else if isInitialTarget {
		operationalLimit = d.maxRecipes
	} else {
		if d.maxRecipes < 10 {
			operationalLimit = 20
		} else {
			operationalLimit = d.maxRecipes * 1000
		}
	}

	allPossibleTreesForCurrentElement := make([]TreeNode, 0)
	productTier := elemDetails.Tier

recipePairLoop:
	for _, recipePair := range elemDetails.Recipes {
		if len(recipePair) != 2 {
			continue
		}
		parent1Name := strings.ToLower(recipePair[0])
		parent2Name := strings.ToLower(recipePair[1])

		elemParent1, p1Exists := d.elementMap[parent1Name]
		elemParent2, p2Exists := d.elementMap[parent2Name]

		if !p1Exists || !p2Exists {
			continue
		}
		if elemParent1.Tier >= productTier || elemParent2.Tier >= productTier {
			continue
		}
		if strings.Contains(elemParent1.Name, "fanon") || strings.Contains(elemParent2.Name, "fanon") {
			continue
		}

		subTreesForParent1 := d.dfsRecursive(parent1Name)
		if !isBasicElement(elemParent1.Name) && len(subTreesForParent1) == 0 {
			continue
		}

		subTreesForParent2 := d.dfsRecursive(parent2Name)
		if !isBasicElement(elemParent2.Name) && len(subTreesForParent2) == 0 {
			continue
		}

	combinationLoop:
		for _, treeP1 := range subTreesForParent1 { 
			for _, treeP2 := range subTreesForParent2 { 
				if operationalLimit > 0 && len(allPossibleTreesForCurrentElement) >= operationalLimit {
					break combinationLoop
				}

				newNode := TreeNode{
					Name:     currentElementNameFormatted,
					Children: []TreeNode{treeP1, treeP2},
				}
				allPossibleTreesForCurrentElement = append(allPossibleTreesForCurrentElement, newNode)
			}
		}

		if operationalLimit > 0 && len(allPossibleTreesForCurrentElement) >= operationalLimit {
			break recipePairLoop
		}
	}

	d.cache[elementToMakeCurrently] = allPossibleTreesForCurrentElement
	return allPossibleTreesForCurrentElement
}