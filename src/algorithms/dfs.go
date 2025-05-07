package main

type Graph map[int][]int

func MultiplePathDFS(graph Graph, startNode, endNode int, maxPaths int) [][]int {
	foundPaths := [][]int{}
	visited := make(map[int]bool)
	currentPath := []int{}
	PathDFS(graph, startNode, endNode, visited, currentPath, &foundPaths, maxPaths)
	return foundPaths 
}


func ShortestPathDFS(graph Graph, startNode, endNode int) []int {
	paths := MultiplePathDFS(graph, startNode, endNode, 100)
	minLen := len(paths[0])
	shortestPath := paths[0]

	for _, path := range paths {
		if len(path) < minLen {
			minLen = len(path)
			shortestPath = path
		}
	}
	return shortestPath
}

func PathDFS(graph Graph, currentNode, endNode int, visited map[int]bool, currentPath []int, foundPaths *[][]int, maxPaths int) bool {
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
		stopSearch := PathDFS(graph, neighbor, endNode, visited, currentPath, foundPaths, maxPaths)
		if stopSearch {
			return true
		}
	}
	return false
}