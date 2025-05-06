package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

type Recipe struct {
	Ingredient1 string
	Ingredient2 string
}

type Node struct {
	Name     string
	Recipes  []Recipe
	Products []string
}

type Step struct {
	Result string
	From   [2]string
}

type State struct {
	Inventory map[string]bool
	Steps     []Step
}

func main() {
	elements, err := loadAndCleanElements("elements.json")
	if err != nil {
		fmt.Println("Gagal load data:", err)
		return
	}

	graph := buildGraph(elements)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Masukkan elemen tujuan: ")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)

	path := bfs(graph, target)
	if len(path) == 0 {
		fmt.Println("Tidak ditemukan jalur dari elemen dasar ke", target)
	} else {
		fmt.Println("Jalur kombinasi:")
		for _, step := range path {
			fmt.Println("- " + step)
		}
	}
}

func bfs(graph map[string]*Node, target string) []string {
	type Step struct {
		Result string
		From   [2]string
	}

	type State struct {
		Inventory map[string]bool
		Steps     []Step
	}

	// Inisialisasi dengan elemen dasar
	startInventory := map[string]bool{
		"air":   true,
		"earth": true,
		"fire":  true,
		"water": true,
	}
	visited := make(map[string]bool)
	queue := []State{{
		Inventory: copyMap(startInventory),
		Steps:     []Step{},
	}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.Inventory[target] {
			var result []string
			for _, step := range curr.Steps {
				result = append(result, fmt.Sprintf("%s + %s → %s", step.From[0], step.From[1], step.Result))
			}
			return result
		}

		// Ambil semua kombinasi 2 elemen dari inventory saat ini
		var elements []string
		for k := range curr.Inventory {
			elements = append(elements, k)
		}

		for i := 0; i < len(elements); i++ {
			for j := i; j < len(elements); j++ {
				a, b := elements[i], elements[j]
				if a > b {
					a, b = b, a
				}
				key := a + "+" + b
				if visited[key] {
					continue
				}
				visited[key] = true

				// Cek apakah kombinasi ini bisa membuat elemen baru
				for result, node := range graph {
					for _, r := range node.Recipes {
						ra, rb := r.Ingredient1, r.Ingredient2
						if (a == ra && b == rb) || (a == rb && b == ra) {
							if curr.Inventory[result] {
								continue
							}
							newInventory := copyMap(curr.Inventory)
							newInventory[result] = true
							newSteps := append([]Step{}, curr.Steps...)
							newSteps = append(newSteps, Step{
								Result: result,
								From:   [2]string{a, b},
							})
							queue = append(queue, State{
								Inventory: newInventory,
								Steps:     newSteps,
							})
						}
					}
				}
			}
		}
	}

	return []string{}
}

func copyMap(original map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range original {
		newMap[k] = v
	}
	return newMap
}

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

func buildGraph(elements []Element) map[string]*Node {
	nodes := make(map[string]*Node)
	seenGlobal := make(map[string]map[string]bool)

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
			if ing1 > ing2 {
				ing1, ing2 = ing2, ing1
			}
			key := ing1 + "+" + ing2
			if _, ok := seenGlobal[e.Name]; !ok {
				seenGlobal[e.Name] = make(map[string]bool)
			}
			if seenGlobal[e.Name][key] {
				continue
			}
			seenGlobal[e.Name][key] = true

			node.Recipes = append(node.Recipes, Recipe{ing1, ing2})

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
