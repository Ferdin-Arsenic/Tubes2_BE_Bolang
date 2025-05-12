package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type RecipeStep struct {
	Element     string
	Ingredients []string
}

type BuildQueueItem struct {
	Path  []RecipeStep
	Open  map[string]bool // elements that still need recipes
	Depth int
}

const (
	bfsMaxDepth     = 50
	bfsMaxQueueSize = 50000
)

var seenPathKeys = make(map[string]bool)

func bfsMultiple(elementMap map[string]Element, target string, maxRecipes int) []TreeNode {
	target = strings.ToLower(target)
	seenPathKeys = make(map[string]bool)

	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}
	}
	if elem, ok := elementMap[target]; !ok || len(elem.Recipes) == 0 {
		return []TreeNode{}
	}

	return bfsBuildRecipeTrees(target, elementMap, maxRecipes)
}

func bfsBuildRecipeTrees(target string, elementMap map[string]Element, maxRecipes int) []TreeNode {
	queue := createInitialQueueItems(target, elementMap)
	results := []TreeNode{}
	fingerprints := make(map[string]bool)

	for len(queue) > 0 && len(results) < maxRecipes {
		curr := queue[0]
		queue = queue[1:]

		if curr.Depth > bfsMaxDepth {
			continue
		}

		if len(curr.Open) == 0 {
			key := pathToStringKey(curr.Path)
			if seenPathKeys[key] || isStructuralDuplicate(curr.Path, elementMap) {
				continue
			}
			seenPathKeys[key] = true

			tree := buildTreeFromSteps(target, curr.Path, elementMap)
			fp := canonicalizeTree(tree)
			if !fingerprints[fp] {
				fingerprints[fp] = true
				results = append(results, tree)
			}
			continue
		}

		for openElem := range curr.Open {
			queue = append(queue, expandOpenElement(openElem, curr, elementMap)...)
			break
		}

		if len(queue) > bfsMaxQueueSize {
			sort.Slice(queue, func(i, j int) bool {
				return queue[i].Depth < queue[j].Depth
			})
			queue = queue[:bfsMaxQueueSize]
		}
	}

	return results
}

func isValidRecipe(a, b string, targetTier int, elementMap map[string]Element) bool {
	elemA, okA := elementMap[a]
	elemB, okB := elementMap[b]

	if !okA || !okB {
		return false
	}
	// ❗ Jika tier belum dikenali (-1), masih bisa dianggap valid
	if elemA.Tier < 0 || elemB.Tier < 0 {
		return true
	}
	return elemA.Tier < targetTier && elemB.Tier < targetTier
}

func createInitialQueueItems(target string, elementMap map[string]Element) []BuildQueueItem {
	queue := []BuildQueueItem{}
	targetTier := elementMap[target].Tier

	for _, recipe := range elementMap[target].Recipes {
		if len(recipe) != 2 {
			continue
		}
		a := strings.ToLower(recipe[0])
		b := strings.ToLower(recipe[1])
		if !isValidRecipe(a, b, targetTier, elementMap) {
			continue
		}

		step := RecipeStep{Element: target, Ingredients: []string{a, b}}
		open := map[string]bool{}
		if !isBasicElement(a) {
			open[a] = true
		}
		if !isBasicElement(b) {
			open[b] = true
		}

		queue = append(queue, BuildQueueItem{
			Path:  []RecipeStep{step},
			Open:  open,
			Depth: 1,
		})
	}
	return queue
}

func expandOpenElement(openElem string, curr BuildQueueItem, elementMap map[string]Element) []BuildQueueItem {
	newItems := []BuildQueueItem{}
	elemTier := elementMap[openElem].Tier

	for _, recipe := range elementMap[openElem].Recipes {
		if len(recipe) != 2 {
			continue
		}
		a := strings.ToLower(recipe[0])
		b := strings.ToLower(recipe[1])
		if !isValidRecipe(a, b, elemTier, elementMap) {
			continue
		}

		newStep := RecipeStep{Element: openElem, Ingredients: []string{a, b}}
		newOpen := copyOpenMap(curr.Open)
		delete(newOpen, openElem)
		if !isBasicElement(a) {
			newOpen[a] = true
		}
		if !isBasicElement(b) {
			newOpen[b] = true
		}

		newPath := append([]RecipeStep{}, curr.Path...)
		newPath = append(newPath, newStep)

		newItems = append(newItems, BuildQueueItem{
			Path:  newPath,
			Open:  newOpen,
			Depth: curr.Depth + 1,
		})
	}

	return newItems
}

func copyOpenMap(orig map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range orig {
		newMap[k] = v
	}
	return newMap
}

func buildTreeFromSteps(root string, steps []RecipeStep, elementMap map[string]Element) TreeNode {
	recipeMap := make(map[string][]string)
	for _, step := range steps {
		recipeMap[step.Element] = step.Ingredients
	}
	tree := buildRecipeTree(root, recipeMap, elementMap, make(map[string]bool), make(map[string]TreeNode))
	sortTreeChildren(&tree)
	return tree
}

func sortTreeChildren(node *TreeNode) {
	if len(node.Children) > 0 {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})
		for i := range node.Children {
			sortTreeChildren(&node.Children[i])
		}
	}
}

func isStructuralDuplicate(steps []RecipeStep, elementMap map[string]Element) bool {
	var normalized []string
	for _, step := range steps {
		a := strings.ToLower(step.Ingredients[0])
		b := strings.ToLower(step.Ingredients[1])
		e := strings.ToLower(step.Element)

		aTier := elementMap[a].Tier
		bTier := elementMap[b].Tier
		eTier := elementMap[e].Tier

		if a > b {
			a, b = b, a
			aTier, bTier = bTier, aTier
		}

		normalized = append(normalized,
			fmt.Sprintf("%s(%d):%s(%d)+%s(%d)", e, eTier, a, aTier, b, bTier))
	}

	sort.Strings(normalized)
	key := strings.Join(normalized, "|")

	if seenPathKeys[key] {
		return true
	}
	seenPathKeys[key] = true
	return false
}

func pathToStringKey(steps []RecipeStep) string {
	keys := make([]string, 0, len(steps))
	for _, step := range steps {
		a := strings.ToLower(step.Ingredients[0])
		b := strings.ToLower(step.Ingredients[1])
		if a > b {
			a, b = b, a
		}
		keys = append(keys, fmt.Sprintf("%s:%s+%s", step.Element, a, b))
	}
	sort.Strings(keys)
	return strings.Join(keys, "|")
}

func canonicalizeTree(node TreeNode) string {
	if len(node.Children) == 0 {
		return node.Name
	}
	var childrenStr []string
	for _, child := range node.Children {
		childrenStr = append(childrenStr, canonicalizeTree(child))
	}
	sort.Strings(childrenStr)
	return fmt.Sprintf("%s(%s)", node.Name, strings.Join(childrenStr, ","))
}

func bfsMultipleLive(elementMap map[string]Element, target string, maxRecipes int, delay int, conn *websocket.Conn) []TreeNode {
	target = strings.ToLower(target)
	seenPathKeys = make(map[string]bool)

	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}
	}
	if elem, ok := elementMap[target]; !ok || len(elem.Recipes) == 0 {
		return []TreeNode{}
	}

	queue := createInitialQueueItems(target, elementMap)
	results := []TreeNode{}
	fingerprints := make(map[string]bool)

	for len(queue) > 0 && len(results) < maxRecipes {
		curr := queue[0]
		queue = queue[1:]

		if curr.Depth > bfsMaxDepth {
			continue
		}

		if len(curr.Open) == 0 {
			key := pathToStringKey(curr.Path)
			if seenPathKeys[key] || isStructuralDuplicate(curr.Path, elementMap) {
				continue
			}
			seenPathKeys[key] = true

			tree := buildTreeFromSteps(target, curr.Path, elementMap)
			fp := canonicalizeTree(tree)
			if !fingerprints[fp] {
				fingerprints[fp] = true
				results = append(results, tree)

				conn.WriteJSON(map[string]interface{}{
					"status":       "Progress",
					"message":      "Finding trees",
					"treeData":     []TreeNode{tree},
					"nodesVisited": 0,
				})
				fmt.Println("Live update sent")
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
			continue
		}

		for openElem := range curr.Open {
			queue = append(queue, expandOpenElement(openElem, curr, elementMap)...)
			break
		}

		if len(queue) > bfsMaxQueueSize {
			sort.Slice(queue, func(i, j int) bool {
				return queue[i].Depth < queue[j].Depth
			})
			queue = queue[:bfsMaxQueueSize]
		}
	}

	return results
}
