package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
	Tier    int        `json:"tier"`
}

type TreeNode struct {
	Name      string     `json:"name"`
	Children  []TreeNode `json:"children,omitempty"`
	Highlight bool       `json:"highlight,omitempty"`
}

var basicElements = []string{"air", "earth", "fire", "water", "time"}

func main() {
	data, err := ioutil.ReadFile("elements.json")
	if err != nil {
		log.Fatalf("Failed to read elements.json: %v", err)
	}

	var elements []Element
	if err := json.Unmarshal(data, &elements); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	elementMap := make(map[string]Element)
	for _, e := range elements {
		elementMap[strings.ToLower(e.Name)] = e
	}

	// Input
	var target string
	fmt.Print("Masukkan target element: ")
	fmt.Scanln(&target)
	target = strings.ToLower(target)

	fmt.Print("Pilih mode (1 = shortest, 2 = multiple): ")
	var mode int
	fmt.Scanln(&mode)

	if mode == 1 {
		// ===== Shortest recipe mode =====
		path := bfsShortest(elementMap, target)
		if len(path) == 0 {
			fmt.Println("Tidak ditemukan jalur ke", target)
			return
		}

		fmt.Println("Shortest path ditemukan:", path)

		visited := make(map[string]bool)
		tree := buildFullTree(target, elementMap, visited)
		tree.Highlight = true

		writeJSON([]TreeNode{tree}, target+"_shortest.json")
		fmt.Println("Tree saved to", target+"_shortest.json")

	} else if mode == 2 {
		// ===== Multiple recipe mode =====
		var maxRecipe int
		fmt.Print("Masukkan maksimal recipe: ")
		fmt.Scanln(&maxRecipe)

		paths := bfsMultiple(elementMap, target, maxRecipe)
		fmt.Println("Ditemukan", len(paths), "recipe")

		var wg sync.WaitGroup
		treeChan := make(chan TreeNode, len(paths))

		for _, path := range paths {
			wg.Add(1)
			go func(p []string) {
				defer wg.Done()
				visited := make(map[string]bool)
				tree := buildFullTree(p[len(p)-1], elementMap, visited)
				tree.Highlight = true
				treeChan <- tree
			}(path)
		}

		wg.Wait()
		close(treeChan)

		unique := make(map[string]bool)
		var allTrees []TreeNode

		for t := range treeChan {
			jsonBytes, _ := json.Marshal(t)
			key := string(jsonBytes)
			if !unique[key] {
				allTrees = append(allTrees, t)
				unique[key] = true
			}
		}

		writeJSON(allTrees, target+"_multiple.json")
		fmt.Println("Semua tree tersimpan di", target+"_multiple.json")

	} else {
		fmt.Println("Mode tidak dikenali.")
	}
}

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

func bfsMultiple(elementMap map[string]Element, target string, maxRecipe int) [][]string {
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

func buildFullTree(name string, elementMap map[string]Element, visited map[string]bool) TreeNode {
	node := TreeNode{Name: capitalize(name)}

	if isBasicElement(name) || visited[name] {
		return node
	}

	visited[name] = true

	elem, exists := elementMap[name]
	if !exists || len(elem.Recipes) == 0 {
		return node
	}

	for _, recipe := range elem.Recipes {
		if len(recipe) != 2 {
			continue
		}
		childNode := TreeNode{
			Name: fmt.Sprintf("%s + %s", capitalize(recipe[0]), capitalize(recipe[1])),
			Children: []TreeNode{
				buildFullTree(strings.ToLower(recipe[0]), elementMap, visited),
				buildFullTree(strings.ToLower(recipe[1]), elementMap, visited),
			},
		}
		node.Children = append(node.Children, childNode)
	}
	return node
}

func writeJSON(data []TreeNode, filename string) {
	f, _ := os.Create(filename)
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

func isBasicElement(name string) bool {
	for _, b := range basicElements {
		if strings.ToLower(name) == strings.ToLower(b) {
			return true
		}
	}
	return false
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
