package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
)

func main() {
	data, err := ioutil.ReadFile("data/elements.json")
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

	var target string
	fmt.Print("Masukkan target element: ")
	fmt.Scanln(&target)
	target = strings.ToLower(target)

	fmt.Print("Pilih mode (1 = shortest/single, 2 = multiple): ")
	var mode int
	fmt.Scanln(&mode)

	// 🔥 TAMBahkan opsi algoritma ke-3 (bidirectional)
	fmt.Print("Pilih algoritma (1 = bfs, 2 = dfs, 3 = bidirectional): ")
	var algo int
	fmt.Scanln(&algo)

	startTime := time.Now()
	if mode == 1 {
		var recipePlan map[string][]string
		if algo == 1 {
			fmt.Println("BFS untuk mode shortest belum diimplementasikan dengan struktur resep baru.")
			return
		} else {
			foundRecipePlans := dfsMultiple(elementMap, target, 1)
			if len(foundRecipePlans) > 0 {
				recipePlan = foundRecipePlans[0]
			}
		}

		if recipePlan == nil {
			fmt.Println("Tidak ditemukan resep untuk", target)
			return
		}

		fmt.Println("Resep ditemukan (via DFS):")
		visited := make(map[string]bool)
		memoCache := make(map[string]TreeNode)
		tree := buildRecipeTree(target, recipePlan, elementMap, visited, memoCache)
		tree.Highlight = true
		writeJSON([]TreeNode{tree}, target+"_single_dfs.json")
		fmt.Println("Tree saved to", target+"_single_dfs.json")

	} else if mode == 2 {
		var recipePlans []map[string][]string
		var maxRecipeInput int
		fmt.Print("Masukkan maksimal recipe: ")
		fmt.Scanln(&maxRecipeInput)

		if algo == 1 {
			fmt.Println("BFS untuk mode multiple belum diimplementasikan dengan struktur resep baru.")
			return
		} else {
			recipePlans = dfsMultiple(elementMap, target, maxRecipeInput)
		}

		fmt.Println("Ditemukan", len(recipePlans), "resep via DFS.")
		if len(recipePlans) == 0 {
			return
		}

		// for i := range len(recipePlans) {
		// 	recipePrinter(recipePlans[i])
		// }

		var wg sync.WaitGroup
		treeChan := make(chan TreeNode, len(recipePlans))

		for _, plan := range recipePlans {
			wg.Add(1)
			go func(p map[string][]string) {
				defer wg.Done()
				localVisited := make(map[string]bool)
				memoCache := make(map[string]TreeNode)

				tree := buildRecipeTree(target, p, elementMap, localVisited, memoCache)
				tree.Highlight = true
				treeChan <- tree
			}(plan)
		}
		fmt.Println("Waktu eksekusi: ", time.Since(startTime))
		wg.Wait()
		close(treeChan)

		var allTrees []TreeNode
		for t := range treeChan {
			allTrees = append(allTrees, t)
		}

		if len(allTrees) > 0 {
			writeJSON(allTrees, target+"_multiple_dfs.json")
			fmt.Println("Semua tree tersimpan di", target+"_multiple_dfs.json")
		} else {
			fmt.Println("Tidak ada tree yang dihasilkan.")
		}

	} else {
		fmt.Println("Mode tidak dikenali.")
	}
}

func recipePrinter(recipe map[string][]string) {
	// Enhanced printing with better formatting
	fmt.Println("Printing recipe:")

	for element, ingredients := range recipe {
		fmt.Printf("  To make %s, combine:\n", element)

		if len(ingredients) == 2 {
			fmt.Printf("    - %s\n    - %s\n", ingredients[0], ingredients[1])
		} else {
			fmt.Println("    (Invalid recipe format)")
		}
	}
	fmt.Println() // Empty line after recipe
}
