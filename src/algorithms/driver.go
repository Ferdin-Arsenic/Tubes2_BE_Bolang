package main

import (
	"fmt"
	"time"
)

func main() {
	// Example Integer Graph
	graph := Graph{}

	edgeList := [][2]int {
		{0, 2}, {0, 4}, {0, 5}, {1, 5}, {2, 3}, {2, 4}, {4, 8}, {5, 6}, {6, 8}, {7, 0}, {7, 4}, {7, 6},
		{0, 1}, {0, 2}, {0, 5}, {1, 3}, {1, 4}, {1, 6}, {2, 4}, {2, 7}, {3, 8}, {4, 5}, {4, 9}, {5, 10}, 
		{6, 9}, {6, 11}, {7, 10}, {7, 12}, {8, 13}, {9, 10}, {9, 14}, {10, 15}, {11, 14}, {11, 16}, {12, 15},
		{12, 17}, {13, 18}, {14, 15}, {14, 18}, {15, 19}, {16, 19}, {17, 19}, {18, 19}, {0, 17}, {3, 16}, {8, 12}, 
		{13, 17},
	}

	for _, edge := range edgeList {
		start, end := edge[0], edge[1]
		graph[start] = append(graph[start], end)
		graph[end] = append(graph[end], start)
	}
	start := time.Now()
	paths := multiplePathDFS(graph, 1, 6, 50)
	for i, path := range paths {
		fmt.Printf("Path %d: %v\n", i+1, path)
		// fmt.Printf("Path : %v\n", paths[6])
	}
	elapsed := time.Since(start)
	fmt.Printf("Execution time: %s\n", elapsed)
	fmt.Print("\n\n")
		
	start = time.Now()
	randomPaths := dfsPathsRandomized(graph, 1, 6, 50)
	for i, path := range randomPaths {
		fmt.Printf("Path %d: %v\n", i+1, path)
	}
	elapsed = time.Since(start)
	fmt.Printf("Execution time: %s\n", elapsed)
	fmt.Print("\n\n")

	// start = time.Now()
	// shortestPath := shortestPathDFS(graph, 1, 6, 10000, false)
	// fmt.Printf("Path : %v\n", shortestPath)
	// elapsed = time.Since(start)
	// fmt.Printf("Execution time: %s\n", elapsed)
}