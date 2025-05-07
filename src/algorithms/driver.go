package main

import "fmt"

func main() {
	// Example Integer Graph
	graph := Graph{}

	edgeList := [][2]int {
		{0, 2},
		{0, 4},
		{0, 5},
		{1, 5},
		{2, 3},
		{2, 4},
		{4, 8},
		{5, 6},
		{6, 8},
		{7, 0},
		{7, 4},
		{7, 6},
	}

	for _, edge := range edgeList {
		start, end := edge[0], edge[1]
		graph[start] = append(graph[start], end)
		graph[end] = append(graph[end], start)
	}

	paths := MultiplePathDFS(graph, 1, 6, 10)
	for i, path := range paths {
		fmt.Printf("Path %d: %v\n", i+1, path)
	}

	fmt.Print("\n\n")

	shortestPath := ShortestPathDFS(graph, 1, 6)
	fmt.Printf("Path : %v\n", shortestPath)
}