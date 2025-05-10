package main

import (
	"strings"
	"log"
	"github.com/gorilla/websocket"
	"time"
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

func BfsMultiple(elementMap map[string]Element, target string, maxRecipe int, conn *websocket.Conn) [][]string {
	queue := [][]string{}
	visited := make(map[string]int)
	var results [][]string

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

		if node == target {
			results = append(results, path)

			visitedTree := make(map[string]bool)
			tree := buildFullTree(target, elementMap, visitedTree)
			tree.Highlight = true
			time.Sleep(time.Second)


			err := conn.WriteJSON(map[string]interface{}{
				"treeData": []TreeNode{tree},
			})
			if err != nil {
				log.Printf("Error sending live tree: %v", err)
			}

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
	return results
}