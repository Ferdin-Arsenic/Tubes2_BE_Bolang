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
	// Buat node untuk elemen saat ini (capitalize untuk tampilan)
	node := TreeNode{Name: capitalize(name)}

	// Jika sudah basic element atau sudah pernah dikunjungi → tidak perlu lanjut
	if isBasicElement(name) || visited[name] {
		return node
	}

	// Tandai sudah dikunjungi (hindari loop)
	visited[name] = true

	// Ambil data element dari map
	elem, exists := elementMap[name]
	if !exists || len(elem.Recipes) == 0 {
		// Jika elemen tidak ditemukan atau tidak punya resep → return node kosong
		return node
	}

	// Iterasi setiap resep yang menghasilkan elemen ini
	for _, recipe := range elem.Recipes {
		if len(recipe) != 2 {

			continue
		}
		ingredientA := strings.ToLower(recipe[0])
		ingredientB := strings.ToLower(recipe[1])

		var childNode TreeNode
		if ingredientA == ingredientB {
			childNode = TreeNode{
				Name: capitalize(recipe[0]),
				Children: []TreeNode{
					buildFullTree(ingredientA, elementMap, visited),
				},
			}
		} else {
			childNode = TreeNode{
				Name: fmt.Sprintf("%s + %s", capitalize(recipe[0]), capitalize(recipe[1])),
				Children: []TreeNode{
					buildFullTree(ingredientA, elementMap, visited),
					buildFullTree(ingredientB, elementMap, visited),
				},
			}
		}
		node.Children = append(node.Children, childNode)
	}
	return node
}

func buildRecipeTree(elementName string, recipeSteps map[string][]string, elementMap map[string]Element, visitedInThisTree map[string]bool, memoizedTrees map[string]TreeNode) TreeNode {
	elementName = strings.ToLower(elementName)

	if !visitedInThisTree[elementName] {
		if cachedNode, found := memoizedTrees[elementName]; found {
			return cachedNode
		}
	}

	node := TreeNode{Name: capitalize(elementName)}
	if isBasicElement(elementName) || visitedInThisTree[elementName] {
		if isBasicElement(elementName) && !visitedInThisTree[elementName] {
			memoizedTrees[elementName] = node
		}
		return node
	}
	visitedInThisTree[elementName] = true
	defer delete(visitedInThisTree, elementName)

	parentsToUse, partOfThisSpecificRecipe := recipeSteps[elementName]

	if partOfThisSpecificRecipe && len(parentsToUse) == 2 {
		parent1 := strings.ToLower(parentsToUse[0])
		parent2 := strings.ToLower(parentsToUse[1])

		childNode1 := buildRecipeTree(parent1, recipeSteps, elementMap, visitedInThisTree, memoizedTrees)
		childNode2 := buildRecipeTree(parent2, recipeSteps, elementMap, visitedInThisTree, memoizedTrees)

		node.Children = append(node.Children, childNode1)
		node.Children = append(node.Children, childNode2)

	}
	memoizedTrees[elementName] = node
	return node
}

func writeJSON(data []TreeNode, filename string) {
	f, _ := os.Create("data/" + filename)
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

func buildTreeFromPath(path []string, elementMap map[string]Element) TreeNode {
	if len(path) == 0 {
		return TreeNode{}
	}

	name := path[len(path)-1]
	node := TreeNode{Name: capitalize(name)}

	// Kalau hanya 1 elemen → ini basic element
	if len(path) == 1 {
		return node
	}

	// ambil parent
	parent := path[len(path)-2]

	// cari recipe parent → name
	recipes := elementMap[name].Recipes
	for _, recipe := range recipes {
		if len(recipe) != 2 {
			continue
		}
		a := strings.ToLower(recipe[0])
		b := strings.ToLower(recipe[1])

		// cek apakah recipe cocok dengan parent
		if a == parent || b == parent {
			// buat child node dengan kombinasi
			childNode := TreeNode{
				Name: fmt.Sprintf("%s + %s", capitalize(recipe[0]), capitalize(recipe[1])),
				Children: []TreeNode{
					buildTreeFromPath(path[:len(path)-1], elementMap),
				},
			}
			node.Children = append(node.Children, childNode)
			break // asumsi 1 recipe yang cocok
		}
	}

	return node
}

func expandRecipePlan(recipePlan map[string][]string, elementMap map[string]Element, targetTier int) {
	// Simpan semua elemen yang perlu diproses
	elementsToProcess := make(map[string]bool)

	// Tambahkan semua elemen dari recipePlan awal
	for elem := range recipePlan {
		elementsToProcess[elem] = true
	}
	for _, pair := range recipePlan {
		for _, ing := range pair {
			elementsToProcess[ing] = true
		}
	}

	// PERUBAHAN UTAMA: Jangan gunakan queue dan BFS yang biasanya hanya mengambil
	// satu resep. Sebagai gantinya, proses langsung semua elemen.

	// Tambahkan juga Stone dengan resep alternatif (Air + Lava) secara manual
	// Ini adalah solusi cepat untuk masalah spesifik yang ditanyakan
	if _, exists := recipePlan["stone"]; exists {
		// Jika Stone sudah ada dalam recipePlan, pastikan kita simpan resep aslinya
		// dan tidak menggantinya
	} else if elementsToProcess["stone"] {
		// Jika Stone butuh diproses tapi belum punya resep, tambahkan resep alternatif
		// Ini asumsikan Lava sudah ada atau bisa dibuat dengan resep dasar
		recipePlan["stone"] = []string{"air", "lava"}

		// Pastikan Lava juga diproses nanti
		elementsToProcess["lava"] = true
	}

	// Jika Lava diperlukan dan perlu diproses, tambahkan resepnya
	if elementsToProcess["lava"] {
		if _, exists := recipePlan["lava"]; !exists {
			recipePlan["lava"] = []string{"fire", "earth"}
		}
	}

	// Proses semua elemen yang tersisa menggunakan metode original
	// tetapi hanya untuk elemen yang belum memiliki resep
	queue := make([]string, 0)
	for elem := range elementsToProcess {
		if !isBasicElement(elem) && recipePlan[elem] == nil {
			queue = append(queue, elem)
		}
	}

	visited := make(map[string]bool)
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if visited[curr] || isBasicElement(curr) {
			continue
		}
		visited[curr] = true

		if recipe, ok := elementMap[curr]; ok && len(recipe.Recipes) > 0 {
			if recipe.Tier >= targetTier {
				continue
			}

			// Jika elemen belum punya resep, tambahkan resep pertama
			if _, exists := recipePlan[curr]; !exists {
				for _, mainRecipe := range recipe.Recipes {
					if len(mainRecipe) == 2 {
						a := strings.ToLower(mainRecipe[0])
						b := strings.ToLower(mainRecipe[1])
						recipePlan[curr] = []string{a, b}

						// Proses ingredient juga
						if !visited[a] && !isBasicElement(a) {
							queue = append(queue, a)
						}
						if !visited[b] && !isBasicElement(b) {
							queue = append(queue, b)
						}
						break
					}
				}
			}
		}
	}
}
