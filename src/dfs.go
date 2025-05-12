package main

import (
	"strings"
	"github.com/gorilla/websocket"
	"time"
	"sync/atomic"
	"sync"
)

type AlgoData struct {
	initialTarget string
	maxRecipes    int
	cache map[string][]TreeNode
	nodeCounter int64
}

func dfsMultiple(target string, maxRecipes int) ([]TreeNode, int) {
	AlgoData := AlgoData{
		initialTarget: strings.ToLower(target),
		maxRecipes:    maxRecipes,
		nodeCounter:   0,
	}

	var resultTrees []TreeNode
	resultTrees = AlgoData.dfsRecursive(strings.ToLower(target))

	if maxRecipes > 0 && len(resultTrees) > maxRecipes {
		return resultTrees[:maxRecipes], int(AlgoData.nodeCounter)
	}
	return resultTrees, int(AlgoData.nodeCounter)
}


func (d *AlgoData) dfsRecursive(currElement string) []TreeNode {
	atomic.AddInt64(&d.nodeCounter, 1)
	currElement = strings.ToLower(currElement)

	elemDetails, exists := elementMap[currElement]
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

	currTreeCombinations := make([]TreeNode, 0)
	productTier := elemDetails.Tier

	recipePairLoop:
	for _, recipePair := range elemDetails.Recipes {
		if len(recipePair) != 2 {
			continue
		}
		parent1Name := strings.ToLower(recipePair[0])
		parent2Name := strings.ToLower(recipePair[1])

		elemParent1, p1Exists := elementMap[parent1Name]
		elemParent2, p2Exists := elementMap[parent2Name]

		if !p1Exists || !p2Exists {
			continue
		}
		if elemParent1.Tier >= productTier || elemParent2.Tier >= productTier {
			continue
		}

		var subTreesForParent1 []TreeNode
		var subTreesForParent2 []TreeNode
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			subTreesForParent1 = d.dfsRecursive(parent1Name)
		}()

		go func() {
			defer wg.Done()
			subTreesForParent2 = d.dfsRecursive(parent2Name)
		}()
		
		wg.Wait()
		if !isBasicElement(elemParent1.Name) && len(subTreesForParent1) == 0 {continue}
		if !isBasicElement(elemParent2.Name) && len(subTreesForParent2) == 0 {continue}

		combinationLoop:
		for _, treeP1 := range subTreesForParent1 { 
			for _, treeP2 := range subTreesForParent2 { 
				if d.maxRecipes > 0 && len(currTreeCombinations) >= d.maxRecipes {
					break combinationLoop
				}

				newNode := TreeNode{
					Name:     elemDetails.Name,
					Children: []TreeNode{treeP1, treeP2},
				}
				currTreeCombinations = append(currTreeCombinations, newNode)
			}
		}

		if d.maxRecipes > 0 && len(currTreeCombinations) >= d.maxRecipes {
			break recipePairLoop
		}
	}
	return currTreeCombinations
}

func dfsMultipleLive(target string, maxRecipes int, delay int, conn *websocket.Conn) ([]TreeNode, int) {
	AlgoData := AlgoData{
		initialTarget: strings.ToLower(target),
		maxRecipes:    maxRecipes,
		cache:         make(map[string][]TreeNode),
		nodeCounter: 0,
	}

	resultTrees := AlgoData.dfsRecursiveLive(strings.ToLower(target), delay, conn)

	if maxRecipes > 0 && len(resultTrees) > maxRecipes {
		return resultTrees[:maxRecipes], int(AlgoData.nodeCounter)
	}
	return resultTrees, int(AlgoData.nodeCounter)
}

func (d *AlgoData) dfsRecursiveLive(currElement string, delay int, conn *websocket.Conn) []TreeNode {
	d.nodeCounter++
	currElement = strings.ToLower(currElement)

	if cachedResult, found := d.cache[currElement]; found {
		return cachedResult
	}

	elemDetails, exists := elementMap[currElement]
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

	currTreeCombinations := make([]TreeNode, 0)
	productTier := elemDetails.Tier

recipePairLoop:
	for _, recipePair := range elemDetails.Recipes {
		if len(recipePair) != 2 {
			continue
		}
		parent1Name := strings.ToLower(recipePair[0])
		parent2Name := strings.ToLower(recipePair[1])

		elemParent1, p1Exists := elementMap[parent1Name]
		elemParent2, p2Exists := elementMap[parent2Name]

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
				if operationalLimit > 0 && len(currTreeCombinations) >= operationalLimit {
					break combinationLoop
				}

				newNode := TreeNode{
					Name:     elemDetails.Name,
					Children: []TreeNode{treeP1, treeP2},
				}
				currTreeCombinations = append(currTreeCombinations, newNode)
			}
		}

		if operationalLimit > 0 && len(currTreeCombinations) >= operationalLimit {
			break recipePairLoop
		}
	}

	conn.WriteJSON(map[string]interface{}{
		"status":   "Progress",
		"message":  "Finding " + elemDetails.Name + " trees",
		"duration": 0,
		"treeData": currTreeCombinations,
		"nodesVisited": d.nodeCounter,
	})
	time.Sleep(time.Duration(delay) * time.Millisecond)

	d.cache[currElement] = currTreeCombinations
	return currTreeCombinations
}