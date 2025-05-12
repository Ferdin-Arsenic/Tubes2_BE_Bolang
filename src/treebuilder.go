package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sort"
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
	queue := make([]string, 0)
	visited := make(map[string]bool)

	// Masukkan semua elemen dari awal plan ke queue
	for elem := range recipePlan {
		queue = append(queue, elem)
		visited[elem] = true
	}
	for _, pair := range recipePlan {
		for _, ing := range pair {
			if !visited[ing] {
				queue = append(queue, ing)
				visited[ing] = true
			}
		}
	}

	// Jalankan BFS untuk melengkapi semua dependency
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if isBasicElement(curr) {
			continue
		}

		if recipe, ok := elementMap[curr]; ok && len(recipe.Recipes) > 0 {
			if recipe.Tier >= targetTier {
				continue
			}

			// Ambil resep pertama saja (atau bisa diatur)
			mainRecipe := recipe.Recipes[0]
			if len(mainRecipe) == 2 {
				a := strings.ToLower(mainRecipe[0])
				b := strings.ToLower(mainRecipe[1])
				if _, exists := recipePlan[curr]; !exists {
					recipePlan[curr] = []string{a, b}
				}
				// Tambahkan ingredient ke queue jika belum
				if !visited[a] {
					queue = append(queue, a)
					visited[a] = true
				}
				if !visited[b] {
					queue = append(queue, b)
					visited[b] = true
				}
			}
		}
	}
}

func generateTreeSignature(node TreeNode) string {
	var sb strings.Builder
	sb.WriteString("N<") // Node Start
	sb.WriteString(node.Name)
	sb.WriteString(">")

	if len(node.Children) > 0 {
		sb.WriteString("C<") // Children Start
		childSignatures := make([]string, len(node.Children))
		for i, child := range node.Children {
			childSignatures[i] = generateTreeSignature(child)
		}
		sort.Strings(childSignatures) // Sort child signatures for canonical order

		for i, sig := range childSignatures {
			sb.WriteString(sig)
			if i < len(childSignatures)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteString(">") // Children End
	}
	return sb.String()
}

// CountUniqueTrees takes a slice of TreeNodes and returns the count of unique tree structures.
func CountUniqueTrees(trees []TreeNode) int {
	if len(trees) == 0 {
		return 0
	}

	seenSignatures := make(map[string]struct{})

	for _, tree := range trees {
		signature := generateTreeSignature(tree)
		// Add the signature to the set. The length of the set at the end
		// will be the count of unique signatures (and thus unique trees).
		seenSignatures[signature] = struct{}{}
	}
	return len(seenSignatures)
}