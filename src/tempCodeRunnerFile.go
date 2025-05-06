unc BFS(graph map[string][]string, start, target string) ([]string, map[string]string) {
	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{start}
	visited[start] = true

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node == target {
			break
		}

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = node
				queue = append(queue, neighbor)
			}
		}
	}

	// Bangun jalur mundur dari target ke start
	var path []string
	curr := target
	for curr != "" {
		path = append([]string{curr}, path...)
		prev, ok := parent[curr]
		if !ok {
			break
		}
		curr = prev
	}
	if len(path) == 0 || path[0] != start {
		return nil, nil
	}
	return path, parent
}