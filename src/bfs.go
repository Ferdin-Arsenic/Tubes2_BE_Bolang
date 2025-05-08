package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
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
	// Load elements.json
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

	// User input
	var target string
	fmt.Print("Masukkan target element: ")
	fmt.Scanln(&target)
	target = strings.ToLower(target)

	// Cari shortest path pakai BFS
	path := bfs(elementMap, target)
	if len(path) == 0 {
		fmt.Println("Tidak ditemukan jalur ke", target)
		return
	}

	fmt.Println("Shortest path ditemukan:", path)

	// Bangun tree rekursif
	visited := make(map[string]bool)
	tree := buildFullTree(target, elementMap, visited)
	tree.Highlight = true

	// Tandai node target sebagai highlight
	tree.Highlight = true

	// Simpan ke JSON
	outFile := "tree_" + target + ".json"
	f, _ := os.Create(outFile)
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(tree)

	fmt.Println("Tree saved to", outFile)
}

// BFS mencari shortest path (list of nodes)
func bfs(elementMap map[string]Element, target string) []string {
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

// Build tree rekursif (sampai basic element)
func buildFullTree(name string, elementMap map[string]Element, visited map[string]bool) TreeNode {
	node := TreeNode{Name: capitalize(name)}

	if isBasicElement(name) || visited[name] {
		// Kalau basic element, atau sudah pernah dikunjungi → jangan expand lagi
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
