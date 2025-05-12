package main

import (
	"strings"
	"sync"
	"sync/atomic"
)

type BIDTreeData struct {
	target       	  string
	maxRecipes   	  int
	maxRecipesPerElmt int

	forwardQueue  [][]string        
	forwardTrees  map[string][]TreeNode
	forwardDepths map[string]int     

	backwardQueue     [][]string       
	backwardReached map[string]bool  
	backwardDepths  map[string]int     

	results        []TreeNode
	processedTrees map[string]bool
	resultMutex    sync.Mutex     

	maxDepth 	 int
	gotoEnd  	 bool
	nodesVisited int64
}

func (b *BIDTreeData) addResult(tree TreeNode) {
	if b.gotoEnd { return }

	b.resultMutex.Lock()
	defer b.resultMutex.Unlock()

	if b.gotoEnd || (b.maxRecipes > 0 && len(b.results) >= b.maxRecipes) {
		b.gotoEnd = true
		return
	}

	canonical := canonicalizeTree(tree)
	if !b.processedTrees[canonical] {
		b.results = append(b.results, tree)
		b.processedTrees[canonical] = true
		if b.maxRecipes > 0 && len(b.results) >= b.maxRecipes {
			b.gotoEnd = true
		}
	}
}

func initializeForwardSearch(b *BIDTreeData, elementMap map[string]Element) {
	var initialForward []string
	initialMap := make(map[string]bool)
	for elName := range elementMap {
		elNameLower := strings.ToLower(elName)
		if isBasicElement(elNameLower) {
			if _, ok := elementMap[elNameLower]; ok {
				tree := TreeNode{Name: capitalize(elNameLower)}
				if len(b.forwardTrees[elNameLower]) < b.maxRecipesPerElmt {
					atomic.AddInt64(&b.nodesVisited, 1)
					b.forwardTrees[elNameLower] = append(b.forwardTrees[elNameLower], tree)
					b.forwardDepths[elNameLower] = 0
					if !initialMap[elNameLower] {
						initialForward = append(initialForward, elNameLower)
						initialMap[elNameLower] = true
					}
				}
			}
		}
	}
	b.forwardQueue[0] = initialForward
}

func initializeBackwardSearch(b *BIDTreeData) {
	b.backwardQueue[0] = []string{b.target}
	b.backwardReached[b.target] = true
	b.backwardDepths[b.target] = 0
	atomic.AddInt64(&b.nodesVisited, 1)
}

func expandForwardLayer(b *BIDTreeData, fLayer int, elementMap map[string]Element) bool {
	if b.gotoEnd || fLayer >= len(b.forwardQueue) || len(b.forwardQueue[fLayer]) == 0 {
		return false
	}

	nextForwardLayerElements := make(map[string]bool)

	for potentialProductLower, productElem := range elementMap {
		if b.gotoEnd { break }

		if len(b.forwardTrees[potentialProductLower]) >= b.maxRecipesPerElmt {
			continue
		}
		if depth, processed := b.forwardDepths[potentialProductLower]; processed && depth < fLayer+1 && len(b.forwardTrees[potentialProductLower]) >= b.maxRecipesPerElmt {
			continue
		}

		productTier := productElem.Tier
		for _, recipe := range productElem.Recipes {
			if b.gotoEnd { break }
			if len(recipe) != 2 { continue }
			p1 := strings.ToLower(recipe[0])
			p2 := strings.ToLower(recipe[1])

			trees1, ok1 := b.forwardTrees[p1]
			trees2, ok2 := b.forwardTrees[p2]
			depth1, depthOk1 := b.forwardDepths[p1]
			depth2, depthOk2 := b.forwardDepths[p2]

			if ok1 && ok2 && depthOk1 && depthOk2 && (depth1 <= fLayer && depth2 <= fLayer) {
				eP1, p1Exists := elementMap[p1]
				eP2, p2Exists := elementMap[p2]
				if !p1Exists || !p2Exists || eP1.Tier >= productTier || eP2.Tier >= productTier {
					continue
				}

				existingTrees := b.forwardTrees[potentialProductLower]
				availableSlots := b.maxRecipesPerElmt - len(existingTrees)

				if availableSlots <= 0 {
					continue
				}

				var wg sync.WaitGroup
				combinedChan := make(chan TreeNode, availableSlots)
				nodesToCombine := len(trees1) * len(trees2)
				maxCombinations := availableSlots
				if nodesToCombine < maxCombinations {
					maxCombinations = nodesToCombine
				}
				var combinationCount int32

				wg.Add(len(trees1))
				for i := range trees1 {
					go func(t1 TreeNode) {
						defer wg.Done()
						for _, t2 := range trees2 {
							currentCount := atomic.AddInt32(&combinationCount, 1)
							if currentCount > int32(maxCombinations) {
								atomic.AddInt32(&combinationCount, -1)
								return
							}
							if b.gotoEnd { return }

							newNode := TreeNode{
								Name:     capitalize(potentialProductLower),
								Children: []TreeNode{t1, t2},
							}
							combinedChan <- newNode
						}
					}(trees1[i])
				}

				var combinedTrees []TreeNode
				collectorWg := sync.WaitGroup{}
				collectorWg.Add(1)
				go func() {
					defer collectorWg.Done()
					for i := 0; i < maxCombinations; i++ {
						tree, ok := <-combinedChan
						if !ok {
							break
						}
						combinedTrees = append(combinedTrees, tree)
					}
				}()

				wg.Wait()
				close(combinedChan)
				collectorWg.Wait()

				if len(combinedTrees) > 0 {
					b.forwardTrees[potentialProductLower] = append(existingTrees, combinedTrees...)

					if _, depthExists := b.forwardDepths[potentialProductLower]; !depthExists || b.forwardDepths[potentialProductLower] > fLayer+1 {
						atomic.AddInt64(&b.nodesVisited, 1)
						b.forwardDepths[potentialProductLower] = fLayer + 1
					}
					if len(b.forwardTrees[potentialProductLower]) > 0 {
						nextForwardLayerElements[potentialProductLower] = true
					}

					if b.backwardReached[potentialProductLower] {
						if potentialProductLower == b.target {
							for _, t := range combinedTrees {
								b.addResult(t)
							}
						}
					}
				}
			}
		}
	}

	var nextQueue []string
	for elem := range nextForwardLayerElements {
		if len(b.forwardTrees[elem]) > 0 {
			nextQueue = append(nextQueue, elem)
		}
	}
	if len(nextQueue) > 0 {
		b.forwardQueue = append(b.forwardQueue, nextQueue)
		return true
	}
	return false
}

func expandBackwardLayer(b *BIDTreeData, bLayer int, elementMap map[string]Element) bool {
	if b.gotoEnd || bLayer >= len(b.backwardQueue) || len(b.backwardQueue[bLayer]) == 0 {
		return false
	}

	nextBackwardLayerElements := make(map[string]bool)
	currentLayerElements := b.backwardQueue[bLayer]

	for _, elemToExpand := range currentLayerElements {
		if b.gotoEnd { break }
		elemDetails, ok := elementMap[elemToExpand]
		if !ok { continue }

		for _, recipe := range elemDetails.Recipes {
			if b.gotoEnd { break }
			if len(recipe) != 2 { continue }
			p1 := strings.ToLower(recipe[0])
			p2 := strings.ToLower(recipe[1])

			eP1, ok1 := elementMap[p1]
			eP2, ok2 := elementMap[p2]
			if !ok1 || !ok2 || eP1.Tier >= elemDetails.Tier || eP2.Tier >= elemDetails.Tier {
				continue
			}

			processBackwardIngredient := func(ing string) {
				if !b.backwardReached[ing] {
					atomic.AddInt64(&b.nodesVisited, 1)
					b.backwardReached[ing] = true
					b.backwardDepths[ing] = bLayer + 1
					nextBackwardLayerElements[ing] = true
					if trees, found := b.forwardTrees[ing]; found {
						if ing == b.target {
							for _, t := range trees {
								b.addResult(t)
							}
						}
					}
				}
			}

			processBackwardIngredient(p1)
			processBackwardIngredient(p2)
		}
	}

	var nextQueue []string
	for elem := range nextBackwardLayerElements {
		nextQueue = append(nextQueue, elem)
	}
	if len(nextQueue) > 0 {
		b.backwardQueue = append(b.backwardQueue, nextQueue)
		return true
	}
	return false
}

func bidirectionalMultiple(target string, maxRecipes int, maxRecipesPerElmt int) ([]TreeNode, int) {
	targetLower := strings.ToLower(target)

	if isBasicElement(targetLower) {
		return []TreeNode{{Name: capitalize(targetLower)}}, 1
	}
	targetElem, exists := elementMap[targetLower]
	if !exists || len(targetElem.Recipes) == 0 {
		return []TreeNode{}, 0
	}

	BIDData := &BIDTreeData{
		target:            targetLower,
		maxRecipes:        maxRecipes,
		maxRecipesPerElmt: maxRecipesPerElmt,
		forwardQueue:      make([][]string, 1),
		forwardTrees:      make(map[string][]TreeNode),
		forwardDepths:     make(map[string]int),
		backwardQueue:     make([][]string, 1),
		backwardReached:   make(map[string]bool),
		backwardDepths:    make(map[string]int),
		results:           make([]TreeNode, 0),
		processedTrees:    make(map[string]bool),
		maxDepth:          20,
		gotoEnd:           false,
	}

	initializeForwardSearch(BIDData, elementMap)
	initializeBackwardSearch(BIDData)

	fLayer, bLayer := 0, 0
	for fLayer < BIDData.maxDepth && bLayer < BIDData.maxDepth {
		if BIDData.gotoEnd {
			break
		}

		forwardExpanded := expandForwardLayer(BIDData, fLayer, elementMap)
		if forwardExpanded {
			fLayer++
		}

		if BIDData.gotoEnd {
			break
		}

		backwardExpanded := expandBackwardLayer(BIDData, bLayer, elementMap)
		if backwardExpanded {
			bLayer++
		}

		if !forwardExpanded && !backwardExpanded &&
		   (fLayer >= len(BIDData.forwardQueue) || len(BIDData.forwardQueue[fLayer]) == 0) &&
		   (bLayer >= len(BIDData.backwardQueue) || len(BIDData.backwardQueue[bLayer]) == 0) {
			break
		}
	}

	return BIDData.results, int(BIDData.nodesVisited)
}