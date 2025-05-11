package main

import "strings"

type DFSData struct {
	elementMap      map[string]Element
    initialTarget   string
    maxRecipes      int
    cache         map[string][]map[string][]string
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

func dfsMultiple(elementMap map[string]Element, target string, maxRecipes int) []map[string][]string {
	dfsData := DFSData{
		elementMap:    elementMap,
		initialTarget: strings.ToLower(target),
		maxRecipes:    maxRecipes,
		cache:         make(map[string][]map[string][]string),
	}

	resultRecipes := dfsData.dfsRecursive(strings.ToLower(target))

	if maxRecipes > 0 && len(resultRecipes) > maxRecipes {
		return resultRecipes[:maxRecipes]
	}
	return resultRecipes
}

func (d *DFSData) dfsRecursive(elementToMakeCurrently string) []map[string][]string {
	elementToMakeCurrently = strings.ToLower(elementToMakeCurrently)

	if cachedResult, found := d.cache[elementToMakeCurrently]; found {
		return cachedResult
	}

	elemDetails, exists := d.elementMap[elementToMakeCurrently]
	if !exists {
		d.cache[elementToMakeCurrently] = []map[string][]string{}
		return []map[string][]string{}
	}

	if isBasicElement(elemDetails.Name) {
		basicRecipeList := []map[string][]string{make(map[string][]string)}
		d.cache[elementToMakeCurrently] = basicRecipeList
		return basicRecipeList
	}

	if len(elemDetails.Recipes) == 0 {
		d.cache[elementToMakeCurrently] = []map[string][]string{}
		return []map[string][]string{}
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
			operationalLimit = d.maxRecipes * 10
		}
	}

	allRecipesForCurrentElement := make([]map[string][]string, 0)
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

		pathsForParent1 := d.dfsRecursive(parent1Name)
		if !isBasicElement(elemParent1.Name) && len(pathsForParent1) == 0 {
			continue
		}

		pathsForParent2 := d.dfsRecursive(parent2Name)
		if !isBasicElement(elemParent2.Name) && len(pathsForParent2) == 0 {
			continue
		}

	combinationLoop:
		for _, pathP1 := range pathsForParent1 {
			for _, pathP2 := range pathsForParent2 {
				if operationalLimit > 0 && len(allRecipesForCurrentElement) >= operationalLimit {
					break combinationLoop
				}

				newRecipe := make(map[string][]string, len(pathP1)+len(pathP2)+1)
				for el, p := range pathP1 {
					newRecipe[el] = p
				}
				for el, p := range pathP2 {
					newRecipe[el] = p
				}
				newRecipe[elementToMakeCurrently] = []string{parent1Name, parent2Name}
				allRecipesForCurrentElement = append(allRecipesForCurrentElement, newRecipe)
			}
		}
		if operationalLimit > 0 && len(allRecipesForCurrentElement) >= operationalLimit {
			break recipePairLoop
		}
	}

	d.cache[elementToMakeCurrently] = allRecipesForCurrentElement
	return allRecipesForCurrentElement
}