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

	var target string
	fmt.Print("Masukkan target element: ")
	fmt.Scanln(&target)
	target = strings.ToLower(target)

	fmt.Print("Pilih mode (1 = shortest/single, 2 = multiple): ")
	var mode int
	fmt.Scanln(&mode)

	fmt.Print("Pilih algoritma (1 = bfs, 2 = dfs): ")
	var algo int
	fmt.Scanln(&algo)

	if mode == 1 {
		var singleRecipePlan map[string][]string
		if algo == 1 {
			fmt.Println("BFS untuk mode shortest belum diimplementasikan dengan struktur resep baru.")
			return
		} else {
			foundRecipePlans := dfsMultiple(elementMap, target,  1)
			if len(foundRecipePlans) > 0 {
				singleRecipePlan = foundRecipePlans[0]
			}
		}

		if singleRecipePlan == nil {
			fmt.Println("Tidak ditemukan resep untuk", target)
			return
		}

		fmt.Println("Resep ditemukan (via DFS):")
		visited := make(map[string]bool)
		tree := buildRecipeTree(target, singleRecipePlan, elementMap, visited)
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
			recipePlans = dfsMultiple(elementMap, target,  maxRecipeInput)
		}

		fmt.Println("Ditemukan", len(recipePlans), "resep via DFS.")
		if len(recipePlans) == 0 {
			return
		}

		var wg sync.WaitGroup
		treeChan := make(chan TreeNode, len(recipePlans))

		for _, plan := range recipePlans {
			wg.Add(1)
			go func(p map[string][]string) {
				defer wg.Done()
				localVisited := make(map[string]bool)
				tree := buildRecipeTree(target, p, elementMap, localVisited)
				tree.Highlight = true
				treeChan <- tree
			}(plan)
		}

		wg.Wait()
		close(treeChan)

		var allTrees []TreeNode
		// Deduplikasi tidak lagi diperlukan jika setiap 'plan' unik dan buildRecipeTree menghasilkan tree unik per plan.
		// Namun, jika 'dfsMultiple' bisa menghasilkan plan identik (meski seharusnya tidak jika logikanya benar),
		// deduplikasi berdasarkan struktur tree mungkin masih berguna. Untuk saat ini, kita asumsikan plan unik.
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