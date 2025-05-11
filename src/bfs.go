package main

import (
	"strings"
)

// bfsShortest finds a single shortest path from basic elements to target
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

// bfsMultiple returns multiple recipe paths in the format required by the tree builder
func bfsMultiple(elementMap map[string]Element, target string, maxRecipes int) []TreeNode {
	target = strings.ToLower(target)

	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}
	}

	// Dapatkan semua resep yang bisa membuat target
	targetRecipes := [][]string{}
	if elem, exists := elementMap[target]; exists {
		for _, recipe := range elem.Recipes {
			if len(recipe) == 2 {
				targetRecipes = append(targetRecipes, []string{
					strings.ToLower(recipe[0]),
					strings.ToLower(recipe[1]),
				})
			}
		}
	}

	if len(targetRecipes) == 0 {
		return []TreeNode{}
	}
	targetTier := elementMap[target].Tier

	var allResults []TreeNode

	for _, recipe := range targetRecipes {
		ingredient1 := recipe[0]
		ingredient2 := recipe[1]

		paths1 := [][]string{}
		if !isBasicElement(ingredient1) {
			paths1 = bfsGetPaths(elementMap, ingredient1, maxRecipes, targetTier)
		} else {
			paths1 = [][]string{{ingredient1}}
		}

		paths2 := [][]string{}
		if !isBasicElement(ingredient2) {
			paths2 = bfsGetPaths(elementMap, ingredient2, maxRecipes, targetTier)
		} else {
			paths2 = [][]string{{ingredient2}}
		}

		for _, path1 := range paths1 {
			for _, path2 := range paths2 {
				if maxRecipes > 0 && len(allResults) >= maxRecipes {
					return allResults
				}

				// gabung path jadi recipeMap, lalu build tree
				recipeMap := combinePathsToRecipe(path1, path2, target, recipe, elementMap, targetTier)
				if len(recipeMap) > 0 {
					expandRecipePlan(recipeMap, elementMap, targetTier)
					tree := buildRecipeTree(
						target,
						recipeMap,
						elementMap,
						make(map[string]bool),
						make(map[string]TreeNode),
					)
					allResults = append(allResults, tree)
				}
			}
		}
	}

	if maxRecipes > 0 && len(allResults) > maxRecipes {
		return allResults[:maxRecipes]
	}
	return allResults
}

// bfsGetPaths mencari jalur BFS dari elemen dasar ke target
func bfsGetPaths(elementMap map[string]Element, target string, maxPaths int, targetTier int) [][]string {

	queue := [][]string{}
	visited := make(map[string]bool)
	var results [][]string

	target = strings.ToLower(target)

	// Mulai dari elemen dasar
	for _, basic := range basicElements {
		queue = append(queue, []string{basic})
	}

	// Track level untuk memastikan kita mendapatkan jalur terpendek
	currentLevel := 1
	itemsAtCurrentLevel := len(queue)
	itemsAtNextLevel := 0
	foundInCurrentLevel := false

	for len(queue) > 0 && (maxPaths <= 0 || len(results) < maxPaths) {
		// Proses node saat ini
		path := queue[0]
		queue = queue[1:]
		itemsAtCurrentLevel--

		node := strings.ToLower(path[len(path)-1])

		// Skip jika sudah dikunjungi
		if visited[node] {
			continue
		}
		visited[node] = true

		// Jika menemukan target, tambahkan ke hasil
		if node == target {
			results = append(results, path)
			foundInCurrentLevel = true
			continue
		}

		// Cari kombinasi yang menghasilkan elemen baru
		for name, elem := range elementMap {
			for _, recipe := range elem.Recipes {
				if len(recipe) != 2 {
					continue
				}
				if elem.Tier >= targetTier {
					continue
				}

				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])

				// Jika elemen saat ini bisa digunakan untuk membuat elemen baru
				if a == node || b == node {
					newPath := append([]string{}, path...)
					newPath = append(newPath, name)
					queue = append(queue, newPath)
					itemsAtNextLevel++
				}
			}
		}

		// Cek jika level saat ini telah selesai diproses
		if itemsAtCurrentLevel == 0 {
			// Jika sudah menemukan target di level ini dan punya hasil, berhenti
			if foundInCurrentLevel && len(results) > 0 {
				break
			}

			// Pindah ke level berikutnya
			currentLevel++
			itemsAtCurrentLevel = itemsAtNextLevel
			itemsAtNextLevel = 0
			foundInCurrentLevel = false
		}
	}

	return results
}

// combinePathsToRecipe menggabungkan dua jalur dan membuat resep map
func combinePathsToRecipe(path1, path2 []string, target string, targetIngredients []string, elementMap map[string]Element, targetTier int) map[string][]string {

	recipeMap := make(map[string][]string)

	// Tambahkan target dengan bahannya
	recipeMap[target] = targetIngredients

	// Proses path1
	for i := 1; i < len(path1); i++ {
		elem := strings.ToLower(path1[i])
		if isBasicElement(elem) {
			continue
		}
		if elementMap[elem].Tier >= targetTier {
			continue
		}

		// Cari recipe untuk elem ini
		for _, recipe := range elementMap[elem].Recipes {
			if len(recipe) != 2 {
				continue
			}

			ingredient1 := strings.ToLower(recipe[0])
			ingredient2 := strings.ToLower(recipe[1])

			// Periksa jika bahan-bahan ada di path sebelumnya
			found1 := false
			found2 := false

			for j := 0; j < i; j++ {
				if strings.ToLower(path1[j]) == ingredient1 {
					found1 = true
				}
				if strings.ToLower(path1[j]) == ingredient2 {
					found2 = true
				}
			}

			// Tambahkan ke recipe jika setidaknya satu bahan ditemukan
			// atau jika itu elemen dasar
			if (found1 || isBasicElement(ingredient1)) &&
				(found2 || isBasicElement(ingredient2)) {
				recipeMap[elem] = []string{ingredient1, ingredient2}
				break
			}
		}
	}

	// Proses path2
	for i := 1; i < len(path2); i++ {
		elem := strings.ToLower(path2[i])
		if isBasicElement(elem) || recipeMap[elem] != nil {
			continue
		}
		if elementMap[elem].Tier >= targetTier {
			continue
		}

		// Cari recipe untuk elem ini
		for _, recipe := range elementMap[elem].Recipes {
			if len(recipe) != 2 {
				continue
			}

			ingredient1 := strings.ToLower(recipe[0])
			ingredient2 := strings.ToLower(recipe[1])

			// Periksa jika bahan-bahan ada di path sebelumnya
			found1 := false
			found2 := false

			for j := 0; j < i; j++ {
				if strings.ToLower(path2[j]) == ingredient1 {
					found1 = true
				}
				if strings.ToLower(path2[j]) == ingredient2 {
					found2 = true
				}
			}

			// Tambahkan ke recipe jika setidaknya satu bahan ditemukan
			// atau jika itu elemen dasar
			if (found1 || isBasicElement(ingredient1)) &&
				(found2 || isBasicElement(ingredient2)) {
				recipeMap[elem] = []string{ingredient1, ingredient2}
				break
			}
		}
	}

	return recipeMap
}

// Fungsi asli untuk konversi jalur ke resep map
func convertPathsToRecipeMaps(paths [][]string, target string, elementMap map[string]Element) []map[string][]string {
	var recipeMaps []map[string][]string

	for _, path := range paths {
		// Skip if the path doesn't lead to the target
		if len(path) == 0 || strings.ToLower(path[len(path)-1]) != target {
			continue
		}

		recipeMap := make(map[string][]string)

		// Build the recipe map from the path
		for i := 1; i < len(path); i++ {
			currentElem := strings.ToLower(path[i])

			// Find which elements combined to create this element
			for _, recipe := range elementMap[currentElem].Recipes {
				if len(recipe) != 2 {
					continue
				}

				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])

				// Check if either ingredient is in our path
				aInPath := false
				bInPath := false

				for j := 0; j < i; j++ {
					if strings.ToLower(path[j]) == a {
						aInPath = true
					}
					if strings.ToLower(path[j]) == b {
						bInPath = true
					}
				}

				// If both ingredients are in our path or at least one is in our path
				// (this is a simplification, may need more sophisticated logic)
				if aInPath || bInPath || isBasicElement(a) || isBasicElement(b) {
					recipeMap[currentElem] = []string{a, b}
					break
				}
			}
		}

		// Only add the recipe map if it's not empty
		if len(recipeMap) > 0 {
			recipeMaps = append(recipeMaps, recipeMap)
		}
	}

	return recipeMaps
}
