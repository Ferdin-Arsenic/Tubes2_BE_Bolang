package main

import (
	"encoding/json"
	"fmt"
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

var basicElements = []string{"air", "earth", "fire", "water"}

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
