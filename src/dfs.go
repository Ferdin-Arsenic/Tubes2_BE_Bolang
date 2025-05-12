package main

import (
	"strings"
	"github.com/gorilla/websocket"
	"time"
)

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

func dfsMultiple(elementMap map[string]Element, target string, maxRecipes int, cached bool) ([]TreeNode, int) {
	dfsData := DFSData{
		elementMap:    elementMap,
		initialTarget: strings.ToLower(target),
		maxRecipes:    maxRecipes,
		cache:         make(map[string][]TreeNode), // Cache stores []TreeNode
		nodeCounter: 0,
	}

	var resultTrees []TreeNode
	if cached{
		resultTrees = dfsData.dfsRecursiveCached(strings.ToLower(target))
	} else {
		resultTrees = dfsData.dfsRecursiveMultithread(strings.ToLower(target))
	}

	if maxRecipes > 0 && len(resultTrees) > maxRecipes {
		return resultTrees[:maxRecipes], dfsData.nodeCounter
	}
	return resultTrees, dfsData.nodeCounter
}

func (d *DFSData) dfsRecursiveCached(currElement string) []TreeNode {
	d.nodeCounter++
	currElement = strings.ToLower(currElement)

	if cachedResult, found := d.cache[currElement]; found {
		return cachedResult
	}

	elemDetails, exists := d.elementMap[currElement]
	if !exists {
		d.cache[currElement] = []TreeNode{}
		return []TreeNode{}
	}

	if isBasicElement(elemDetails.Name) {
		leafNode := TreeNode{Name: elemDetails.Name}
		basicTreeList := []TreeNode{leafNode}
		d.cache[currElement] = basicTreeList
		return basicTreeList
	}

	if len(elemDetails.Recipes) == 0 {
		leafNode := TreeNode{Name: elemDetails.Name}
		noRecipeTreeList := []TreeNode{leafNode}
		d.cache[currElement] = noRecipeTreeList
		return noRecipeTreeList
	}

	var operationalLimit int
	if d.maxRecipes <= 0 {
		operationalLimit = 0
	} else {
		operationalLimit = d.maxRecipes
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

		subTreesForParent1 := d.dfsRecursiveCached(parent1Name)
		if !isBasicElement(elemParent1.Name) && len(subTreesForParent1) == 0 {
			continue
		}

		subTreesForParent2 := d.dfsRecursiveCached(parent2Name)
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
					Name:     elemDetails.Name,
					Children: []TreeNode{treeP1, treeP2},
				}
				allPossibleTreesForCurrentElement = append(allPossibleTreesForCurrentElement, newNode)
			}
		}

		if operationalLimit > 0 && len(allPossibleTreesForCurrentElement) >= operationalLimit {
			break recipePairLoop
		}
	}

	d.cache[currElement] = allPossibleTreesForCurrentElement
	return allPossibleTreesForCurrentElement
}

func (d *DFSData) dfsRecursiveMultithread(currElement string) []TreeNode {
	d.nodeCounter++
	currElement = strings.ToLower(currElement)

	elemDetails, exists := d.elementMap[currElement]
	if !exists {
		return []TreeNode{}
	}

	if isBasicElement(elemDetails.Name) {
		leafNode := TreeNode{Name: elemDetails.Name}
		basicTreeList := []TreeNode{leafNode}
		return basicTreeList
	}

	if len(elemDetails.Recipes) == 0 {
		leafNode := TreeNode{Name: elemDetails.Name}
		noRecipeTreeList := []TreeNode{leafNode}
		return noRecipeTreeList
	}

	var operationalLimit int

	if d.maxRecipes <= 0 {
		operationalLimit = 0
	} else {
		operationalLimit = d.maxRecipes
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

		subTreesForParent1 := d.dfsRecursiveMultithread(parent1Name)
		if !isBasicElement(elemParent1.Name) && len(subTreesForParent1) == 0 {
			continue
		}

		subTreesForParent2 := d.dfsRecursiveMultithread(parent2Name)
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
					Name:     elemDetails.Name,
					Children: []TreeNode{treeP1, treeP2},
				}
				allPossibleTreesForCurrentElement = append(allPossibleTreesForCurrentElement, newNode)
			}
		}

		if operationalLimit > 0 && len(allPossibleTreesForCurrentElement) >= operationalLimit {
			break recipePairLoop
		}
	}

	return allPossibleTreesForCurrentElement
}

func dfsMultipleLive(elementMap map[string]Element, target string, maxRecipes int, delay int, conn *websocket.Conn) ([]TreeNode, int) {
	dfsData := DFSData{
		elementMap:    elementMap,
		initialTarget: strings.ToLower(target),
		maxRecipes:    maxRecipes,
		cache:         make(map[string][]TreeNode), // Cache stores []TreeNode
		nodeCounter: 0,
	}

	resultTrees := dfsData.dfsRecursiveLive(strings.ToLower(target), delay, conn)

	// Optional: Apply maxRecipes limit to the final list of trees
	// Your external de-duplication might happen before or after this.
	if maxRecipes > 0 && len(resultTrees) > maxRecipes {
		return resultTrees[:maxRecipes], dfsData.nodeCounter
	}
	return resultTrees, dfsData.nodeCounter
}

func (d *DFSData) dfsRecursiveLive(currElement string, delay int, conn *websocket.Conn) []TreeNode {
	d.nodeCounter++
	currElement = strings.ToLower(currElement)

	if cachedResult, found := d.cache[currElement]; found {
		return cachedResult
	}

	elemDetails, exists := d.elementMap[currElement]
	if !exists {
		d.cache[currElement] = []TreeNode{}
		return []TreeNode{}
	}

	if isBasicElement(elemDetails.Name) {
		leafNode := TreeNode{Name: elemDetails.Name}
		basicTreeList := []TreeNode{leafNode}
		d.cache[currElement] = basicTreeList
		return basicTreeList
	}

	if len(elemDetails.Recipes) == 0 {
		leafNode := TreeNode{Name: elemDetails.Name}
		noRecipeTreeList := []TreeNode{leafNode}
		d.cache[currElement] = noRecipeTreeList
		return noRecipeTreeList
	}

	var operationalLimit int
	isInitialTarget := (d.initialTarget == currElement)

	if d.maxRecipes <= 0 {
		operationalLimit = 0
	} else if isInitialTarget {
		operationalLimit = d.maxRecipes
	} else {
		if d.maxRecipes < 10 {
			operationalLimit = 20
		} else {
			operationalLimit = d.maxRecipes
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

		subTreesForParent1 := d.dfsRecursiveLive(parent1Name, delay, conn)
		if !isBasicElement(elemParent1.Name) && len(subTreesForParent1) == 0 {
			continue
		}

		subTreesForParent2 := d.dfsRecursiveLive(parent2Name, delay, conn)
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
					Name:     elemDetails.Name,
					Children: []TreeNode{treeP1, treeP2},
				}
				allPossibleTreesForCurrentElement = append(allPossibleTreesForCurrentElement, newNode)
			}
		}

		if operationalLimit > 0 && len(allPossibleTreesForCurrentElement) >= operationalLimit {
			break recipePairLoop
		}
	}

	conn.WriteJSON(map[string]interface{}{
		"status":   "Progress",
		"message":  "Finding " + elemDetails.Name + " trees",
		"duration": 0,
		"treeData": allPossibleTreesForCurrentElement,
		"nodesVisited": d.nodeCounter,
	})
	time.Sleep(time.Duration(delay) * time.Millisecond)

	d.cache[currElement] = allPossibleTreesForCurrentElement
	return allPossibleTreesForCurrentElement
}