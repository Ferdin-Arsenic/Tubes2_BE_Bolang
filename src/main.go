package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"log"
)

func main() {
	// mainBFS();
	mainDFS();
}

func mainDFS(){
	data, err := ioutil.ReadFile("elements.json")
	if err != nil {log.Fatalf("Failed to read elements.json: %v", err)}

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

	fmt.Print("Pilih algoritma (1 = bfs, 2 = dfs): ")
	var algo int
	fmt.Scanln(&algo)


	var path []string
	if mode == 1 {
		// ===== Shortest recipe mode =====
		if algo == 1 {
			path = bfsShortest(elementMap, target)
		} else {
			path = dfsShortest()
		}
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
		var paths [][]string
		var maxRecipe int
		fmt.Print("Masukkan maksimal recipe: ")
		fmt.Scanln(&maxRecipe)

		if algo == 1 {
			paths = bfsMultiple(elementMap, target, maxRecipe)
		
		} else {
			paths = dfsMultiple(elementMap, basicElements,target, maxRecipe)
		}
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

func mainBFS(){
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