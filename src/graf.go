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
}

type Node struct {
	Name     string
	Recipes  []Recipe
	Products []string
}

type Recipe struct {
	Ingredient1 string
	Ingredient2 string
}

func main() {
	elements, err := loadAndCleanElements("elements.json")
	if err != nil {
		fmt.Println("Gagal memuat elements.json:", err)
		return
	}

	graph := buildGraph(elements)
	err = exportGraphToFile(graph, "graph_output.txt")
	if err != nil {
		fmt.Println("Gagal menyimpan graf:", err)
	} else {
		fmt.Println("Graf disimpan di graph_output.txt")
	}

	fmt.Printf("Graf berhasil dibangun dengan %d simpul\n", len(graph))

	// Contoh: tampilkan 10 node pertama dan resepnya
	i := 0
	for name, node := range graph {
		fmt.Println("Element:", name)
		for _, r := range node.Recipes {
			fmt.Printf("  - %s + %s → %s\n", r.Ingredient1, r.Ingredient2, name)
		}
		i++
		if i >= 10 {
			break
		}
	}
}

// loadAndCleanElements membaca dan merapikan string resep
func loadAndCleanElements(filename string) ([]Element, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var raw []Element
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	var cleaned []Element
	for _, e := range raw {
		newRecipes := [][]string{}
		for _, r := range e.Recipes {
			if len(r) != 2 {
				continue
			}
			lines := strings.Split(r[1], "\n")
			for _, line := range lines {
				parts := strings.Split(line, "+")
				if len(parts) == 2 {
					ing1 := strings.TrimSpace(parts[0])
					ing2 := strings.TrimSpace(parts[1])
					if ing1 != "" && ing2 != "" {
						newRecipes = append(newRecipes, []string{ing1, ing2})
					}
				}
			}
		}
		cleaned = append(cleaned, Element{
			Name:    e.Name,
			Recipes: newRecipes,
		})
	}

	return cleaned, nil
}

// buildGraph membangun peta node dari elemen dan resep
func buildGraph(elements []Element) map[string]*Node {
	nodes := make(map[string]*Node)
	seenGlobal := make(map[string]map[string]bool) // NEW: Track per-output-element

	for _, e := range elements {
		if _, ok := nodes[e.Name]; !ok {
			nodes[e.Name] = &Node{
				Name:     e.Name,
				Recipes:  []Recipe{},
				Products: []string{},
			}
		}
		node := nodes[e.Name]
		for _, r := range e.Recipes {
			if len(r) != 2 {
				continue
			}
			ing1 := strings.TrimSpace(r[0])
			ing2 := strings.TrimSpace(r[1])

			// Normalize order
			if ing1 > ing2 {
				ing1, ing2 = ing2, ing1
			}
			key := ing1 + "+" + ing2

			// Cek global map per node
			if _, ok := seenGlobal[e.Name]; !ok {
				seenGlobal[e.Name] = make(map[string]bool)
			}
			if seenGlobal[e.Name][key] {
				continue // Sudah ada sebelumnya
			}
			seenGlobal[e.Name][key] = true

			// Simpan resep
			node.Recipes = append(node.Recipes, Recipe{ing1, ing2})

			// Tambahkan node bahan jika belum ada
			for _, ing := range []string{ing1, ing2} {
				if _, ok := nodes[ing]; !ok {
					nodes[ing] = &Node{
						Name:     ing,
						Recipes:  []Recipe{},
						Products: []string{},
					}
				}
				nodes[ing].Products = append(nodes[ing].Products, e.Name)
			}
		}
	}

	return nodes
}

func exportGraphToFile(graph map[string]*Node, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for name, node := range graph {
		if len(node.Recipes) > 0 {
			fmt.Fprintf(file, "Element: %s\n", name)
			for _, r := range node.Recipes {
				fmt.Fprintf(file, "  - %s + %s → %s\n", r.Ingredient1, r.Ingredient2, name)
			}
			fmt.Fprintln(file)
		}
	}

	return nil
}
