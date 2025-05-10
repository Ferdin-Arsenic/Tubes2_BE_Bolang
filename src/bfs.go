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

func bfsShortest(elementMap map[string]Element, target string) []string {
	queue := [][]string{}
	visited := make(map[string]bool)

	for _, basic := range basicElements {
		queue = append(queue, []string{basic})
	}

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		node := strings.ToLower(path[len(path)-1])

		if visited[node] {
			continue
		}
		visited[node] = true

		if node == target {
			return path
		}

		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				if a == node || b == node {
					newPath := append([]string{}, path...)
					newPath = append(newPath, name)
					queue = append(queue, newPath)
				}
			}
		}
	}
	return nil
}

func BfsMultiple(elementMap map[string]Element, target string, maxRecipe int, conn *websocket.Conn) int {
	queue := [][]string{}
	visited := make(map[string]int)
	var results [][]string
	nodeVisited := 0

	for _, basic := range basicElements {
		queue = append(queue, []string{basic})
	}

	for len(queue) > 0 && len(results) < maxRecipe {
		path := queue[0]
		queue = queue[1:]
		node := strings.ToLower(path[len(path)-1])

		if visited[node] >= maxRecipe {
			continue
		}
		visited[node]++
		nodeVisited++

		if node == target {
			results = append(results, path)
			continue
		}

		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				if a == node || b == node {
					newPath := append([]string{}, path...)
					newPath = append(newPath, name)
					queue = append(queue, newPath)
				}
			}
		}
	}

	var wg sync.WaitGroup
	treeChan := make(chan TreeNode, len(results))

	for _, path := range results {
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
				"message":  fmt.Sprintf("Streaming %d of %d trees (Nodes visited: %d)", len(unique), len(results), nodeVisited),
				"treeData": []TreeNode{tree},
			})
			if err != nil {
				log.Printf("Error sending tree: %v", err)
				break
			}
			log.Printf("Sent tree %d of %d (Nodes visited: %d)", len(unique), len(results), nodeVisited)
			time.Sleep(2 * time.Second)
		}
	}

	return nodeVisited
}