package main

import (
	"math/rand"
	"time"
)

type Graph map[int][]int

func multiplePathDFS(graph Graph, startNode, endNode int, maxPaths int) [][]int {
	foundPaths := [][]int{}
	visited := make(map[int]bool)
	currentPath := []int{}
	dfs(graph, startNode, endNode, visited, currentPath, &foundPaths, maxPaths)
	return foundPaths 
}

// Structured dfs
func dfs(graph Graph, currentNode, endNode int, visited map[int]bool, currentPath []int, foundPaths *[][]int, maxPaths int) bool {
	if len(*foundPaths) >= maxPaths {
		return true
	}

	if visited[currentNode] {
		return false
	}

	visited[currentNode] = true
	defer func() { delete(visited, currentNode)} ()

	currentPath = append(currentPath, currentNode)
	if currentNode == endNode {
		pathCopy := make([]int, len(currentPath))
		copy(pathCopy, currentPath)
		*foundPaths = append(*foundPaths, pathCopy)

		return len(*foundPaths) >= maxPaths
	}

	neighbors, exists := graph[currentNode]
	if !exists {
		return false
	}

	for _, neighbor := range neighbors {
		stopSearch := dfs(graph, neighbor, endNode, visited, currentPath, foundPaths, maxPaths)
		if stopSearch {
			return true
		}
	}
	return false
}

// Multiple Paths DFS with randomization
func dfsPathsRandomized(graph Graph, startNode, endNode, maxPaths int) [][]int {
	rand.Seed(time.Now().UnixNano())

	foundPaths := [][]int{}
	visited := make(map[int]bool)
	currentPath := []int{}

	dfsRandomized(graph, startNode, endNode, visited, currentPath, &foundPaths, maxPaths)

	return foundPaths
}

func dfsRandomized(graph Graph, currentNode, endNode int, visited map[int]bool, currentPath []int, foundPaths *[][]int, maxPaths int) bool {
	if len(*foundPaths) >= maxPaths {return true}

	if visited[currentNode] {return false}

	visited[currentNode] = true
	defer func() { delete(visited, currentNode) }()

	currentPath = append(currentPath, currentNode)
	if currentNode == endNode {
		pathCopy := make([]int, len(currentPath))
		copy(pathCopy, currentPath)
		*foundPaths = append(*foundPaths, pathCopy)

		return len(*foundPaths) >= maxPaths
	}

	neighbors, exists := graph[currentNode]
	if !exists {
		return false
	}

	// Create a copy of neighbors to shuffle
	shuffledNeighbors := make([]int, len(neighbors))
	copy(shuffledNeighbors, neighbors)

	// Randomize neighbors order
	rand.Shuffle(len(shuffledNeighbors), func(i, j int) {
		shuffledNeighbors[i], shuffledNeighbors[j] = shuffledNeighbors[j], shuffledNeighbors[i]
	})

	for _, neighbor := range shuffledNeighbors {
		stopSearch := dfsRandomized(graph, neighbor, endNode, visited, currentPath, foundPaths, maxPaths)
		if stopSearch {
			return true
		}
	}

	return false
}

// Uses randomized slices for BFS-like shortest pathfinding
func shortestPathDFS(graph Graph, startNode, endNode int, precision int, randomize bool) []int {
	foundPaths := [][]int{}
	if (randomize){
		 foundPaths = dfsPathsRandomized(graph, startNode, endNode, precision)
	} else {
		foundPaths = multiplePathDFS(graph, startNode, endNode, precision)
	}

	minLen := len(foundPaths[0])
	shortestPath := foundPaths[0]
	for _, path := range foundPaths {
		if len(path) < minLen {
			minLen = len(path)
			shortestPath = path
		}
	}
	return shortestPath
}