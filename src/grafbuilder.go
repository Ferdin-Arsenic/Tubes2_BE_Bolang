package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
	Tier    int        `json:"tier"`
}

var basicElements = []string{"Fire", "Water", "Air", "Earth"}

func BuildGraph(filename string) map[string][]string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Gagal membuka file: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Gagal membaca file: %v", err)
	}

	var elements []Element
	if err := json.Unmarshal(bytes, &elements); err != nil {
		log.Fatalf("Gagal parse JSON: %v", err)
	}

	graph := make(map[string][]string)

	for _, elem := range elements {
		output := elem.Name
		for _, recipe := range elem.Recipes {
			if len(recipe) != 2 {
				continue
			}
			ing1 := strings.TrimSpace(recipe[0])
			ing2 := strings.TrimSpace(recipe[1])

			graph[ing1] = append(graph[ing1], output)
			graph[ing2] = append(graph[ing2], output)

			fmt.Printf("Edge: %s → %s\n", ing1, output)
			fmt.Printf("Edge: %s → %s\n", ing2, output)
		}
	}

	return graph
}

func BFS(graph map[string][]string, starts []string, target string) []string {
	visited := make(map[string]bool)
	queue := [][]string{}

	for _, start := range starts {
		queue = append(queue, []string{start})
	}

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		node := path[len(path)-1]

		if visited[node] {
			continue
		}
		visited[node] = true

		if strings.EqualFold(node, target) {
			return path
		}

		for _, neighbor := range graph[node] {
			newPath := append([]string{}, path...)
			newPath = append(newPath, neighbor)
			queue = append(queue, newPath)
		}
	}
	return nil
}

func DFS(graph map[string][]string, starts []string, target string) []string {
	visited := make(map[string]bool)
	var result []string

	var dfsHelper func(path []string, node string) bool
	dfsHelper = func(path []string, node string) bool {
		if visited[node] {
			return false
		}
		visited[node] = true
		path = append(path, node)

		if strings.EqualFold(node, target) {
			result = path
			return true
		}

		for _, neighbor := range graph[node] {
			if dfsHelper(append([]string{}, path...), neighbor) {
				return true
			}
		}
		return false
	}

	for _, start := range starts {
		if dfsHelper([]string{}, start) {
			return result
		}
	}
	return nil
}

func main() {
	graph := BuildGraph("elements.json")

	// sort neighbor list
	for node := range graph {
		sort.Strings(graph[node])
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Program Pencarian Elemen dari Elemen Dasar ===")

	fmt.Print("Masukkan elemen target: ")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)

	fmt.Print("Pilih metode pencarian (BFS/DFS): ")
	method, _ := reader.ReadString('\n')
	method = strings.ToUpper(strings.TrimSpace(method))

	var path []string
	if method == "BFS" {
		path = BFS(graph, basicElements, target)
	} else if method == "DFS" {
		path = DFS(graph, basicElements, target)
	} else {
		fmt.Println("Metode tidak dikenali. Gunakan 'BFS' atau 'DFS'.")
		return
	}

	if path != nil {
		fmt.Println("Jalur ditemukan:")
		fmt.Println(strings.Join(path, " -> "))
	} else {
		fmt.Println("Tidak ditemukan jalur dari elemen dasar ke", target)
	}
}
