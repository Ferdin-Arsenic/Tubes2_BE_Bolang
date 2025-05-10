package main

import (
	"strings"
	"github.com/gorilla/websocket"
	"time"
	"fmt"
	"encoding/json"
	"log"
	"sync"
)

func DfsMultiple(elementMap map[string]Element, basicElements []string, target string, maxPaths int, conn *websocket.Conn) int {
	foundPaths := [][]string{}
	nodeVisited := 0
	targetNormalized := strings.ToLower(target)

	for _, startElement := range basicElements {
		if len(foundPaths) >= maxPaths {
			break
		}
		visitedOnCurrentPath := make(map[string]bool)
		initialPath := []string{startElement}
		nodeVisited++

		dfsHelper(
			elementMap,
			startElement,
			targetNormalized,
			visitedOnCurrentPath,
			initialPath,
			&foundPaths,
			maxPaths,
		)
	}

	var wg sync.WaitGroup
	treeChan := make(chan TreeNode, len(foundPaths))

	for _, path := range foundPaths {
		wg.Add(1)
		go func(p []string) {
			defer wg.Done()
			visited := make(map[string]bool)
			tree := buildFullTree(p[len(p)-1], elementMap, visited)
			tree.Highlight = true
			treeChan <- tree
		}(path)
	}

	go func() {
		wg.Wait()
		close(treeChan)
	}()

	unique := make(map[string]bool)

	for tree := range treeChan {
		jsonBytes, _ := json.Marshal(tree)
		key := string(jsonBytes)

		if !unique[key] {
			unique[key] = true
			err := conn.WriteJSON(map[string]interface{}{
				"status":   "Tree Update",
				"message":  fmt.Sprintf("Streaming %d of %d trees (Nodes visited: %d)", len(unique), len(foundPaths), nodeVisited),
				"treeData": []TreeNode{tree},
			})
			if err != nil {
				log.Printf("Error sending tree: %v", err)
				break
			}
			log.Printf("Sent tree %d of %d (Nodes visited: %d)", len(unique), len(foundPaths), nodeVisited)
			time.Sleep(2 * time.Second)
		}
	}

	return nodeVisited
}


// Recursive DFS helper function
func dfsHelper(elementMap map[string]Element, currentElement string, targetNode string, visitedOnCurrentPath map[string]bool, currentPathInProgress []string, foundPaths *[][]string, maxPaths int) {
	if len(*foundPaths) >= maxPaths {return}

	currentElementLower := strings.ToLower(currentElement)

	if visitedOnCurrentPath[currentElementLower] {return}

	visitedOnCurrentPath[currentElementLower] = true
	
	defer func() { delete(visitedOnCurrentPath, currentElementLower) }()

	if currentElementLower == targetNode {
		pathCopy := make([]string, len(currentPathInProgress))
		copy(pathCopy, currentPathInProgress)
		*foundPaths = append(*foundPaths, pathCopy)
		return
	}

	for potentialNextProduct, details := range elementMap {
		if len(*foundPaths) >= maxPaths {
			return
		}

		for _, recipe := range details.Recipes {
			if len(recipe) != 2 {
				continue
			}
			ing1 := strings.ToLower(recipe[0])
			ing2 := strings.ToLower(recipe[1])

			if ing1 == currentElementLower || ing2 == currentElementLower {
				nextPath := append([]string{}, currentPathInProgress...)
				nextPath = append(nextPath, potentialNextProduct)

				dfsHelper( elementMap, potentialNextProduct, targetNode, visitedOnCurrentPath, nextPath, foundPaths, maxPaths)

				if len(*foundPaths) >= maxPaths {return}
				break
			}
		}
	}
}