package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"log"
	"net/http"
	"strconv"
	"github.com/gorilla/websocket"
	"time"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

type RequestData struct {
    Algorithm  string `json:"algorithm"`
    Target     string `json:"target"`
    MaxRecipes string `json:"maxRecipes"`
}
// func climain(){
// 	data, err := ioutil.ReadFile("elements.json")
// 	if err != nil {log.Fatalf("Failed to read elements.json: %v", err)}

// 	var elements []Element
// 	if err := json.Unmarshal(data, &elements); err != nil {
// 		log.Fatalf("Failed to parse JSON: %v", err)
// 	}

// 	elementMap := make(map[string]Element)
// 	for _, e := range elements {
// 		elementMap[strings.ToLower(e.Name)] = e
// 	}

// 	// Input
// 	var target string
// 	fmt.Print("Masukkan target element: ")
// 	fmt.Scanln(&target)
// 	target = strings.ToLower(target)

// 	fmt.Print("Pilih mode (1 = shortest, 2 = multiple): ")
// 	var mode int
// 	fmt.Scanln(&mode)

// 	fmt.Print("Pilih algoritma (1 = bfs, 2 = dfs): ")
// 	var algo int
// 	fmt.Scanln(&algo)


// 	var path []string
// 	if mode == 1 {
// 		// ===== Shortest recipe mode =====
// 		if algo == 1 {
// 			path = bfsShortest(elementMap, target)
// 		} else {
// 			path = dfsMultiple(elementMap, basicElements,target, 1)[0]
// 		}
// 		if len(path) == 0 {
// 			fmt.Println("Tidak ditemukan jalur ke", target)
// 			return
// 		}

// 		fmt.Println("Shortest path ditemukan:", path)

// 		visited := make(map[string]bool)
// 		tree := buildFullTree(target, elementMap, visited)
// 		tree.Highlight = true

// 		writeJSON([]TreeNode{tree}, target+"_shortest.json")
// 		fmt.Println("Tree saved to", target+"_shortest.json")
	
// 	} else if mode == 2 {
// 		var paths [][]string
// 		var maxRecipe int
// 		fmt.Print("Masukkan maksimal recipe: ")
// 		fmt.Scanln(&maxRecipe)

// 		if algo == 1 {
// 			paths = bfsMultiple(elementMap, 
// 				target, maxRecipe)
		
// 		} else {
// 			paths = dfsMultiple(elementMap, basicElements,target, maxRecipe)
// 		}
// 		fmt.Println("Ditemukan", len(paths), "recipe")
// 		fmt.Println("Path yang ditemukan:")
// 		for _, path := range paths {
// 			fmt.Println(path)
// 		}


// 		var wg sync.WaitGroup
// 		treeChan := make(chan TreeNode, len(paths))

// 		for _, path := range paths {
// 			wg.Add(1)
// 			go func(p []string) {
// 				defer wg.Done()
// 				visited := make(map[string]bool)
// 				tree := buildFullTree(p[len(p)-1], elementMap, visited)
// 				tree.Highlight = true
// 				treeChan <- tree
// 			}(path)
// 		}

// 		wg.Wait()
// 		close(treeChan)

// 		unique := make(map[string]bool)
// 		var allTrees []TreeNode

// 		for t := range treeChan {
// 			jsonBytes, _ := json.Marshal(t)
// 			key := string(jsonBytes)
// 			if !unique[key] {
// 				allTrees = append(allTrees, t)
// 				unique[key] = true
// 			}
// 		}

// 		writeJSON(allTrees, target+"_multiple.json")
// 		fmt.Println("Semua tree tersimpan di", target+"_multiple.json")

// 	} else {
// 		fmt.Println("Mode tidak dikenali.")
// 	}
// }

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")

	var reqData RequestData
	err = conn.ReadJSON(&reqData)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	log.Printf("Received request - Element: %s, Algorithm: %s, MaxRecipes: %s", 
		reqData.Target, reqData.Algorithm, reqData.MaxRecipes)

	conn.WriteJSON(map[string]interface{}{
		"status": "Processing",
		"message": "Loading elements data",
	})

	data, err := ioutil.ReadFile("elements.json")
	if err != nil {
		log.Fatalf("Failed to read elements.json: %v", err)
		conn.WriteJSON(map[string]interface{}{
			"error": "Failed to read elements data",
		})
		return
	}

	var elements []Element
	if err := json.Unmarshal(data, &elements); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
		conn.WriteJSON(map[string]interface{}{
			"error": "Failed to parse elements data",
		})
		return
	}

	elementMap := make(map[string]Element)
	for _, e := range elements {
		elementMap[strings.ToLower(e.Name)] = e
	}

	if reqData.Algorithm == "BFS" {

		count, err := strconv.Atoi(reqData.MaxRecipes)
		if err != nil {
			conn.WriteJSON(map[string]interface{}{
				"error": "Invalid MaxRecipes value",
			})
			return
		}

		start := time.Now()
		paths := BfsMultiple(elementMap, strings.ToLower(reqData.Target), count, conn)
		elapsed := time.Since(start)
		
		

		
		err = conn.WriteJSON(map[string]interface{}{
			"resultInfo": map[string]string{
				"time": elapsed.String(),
				"node": fmt.Sprintf("%d", len(paths)),
			},
			"status": "Completed",
		})
		if err != nil {
			log.Printf("Error sending final result: %v", err)
		}

	} else if reqData.Algorithm == "DFS" {
		log.Println("DFS algorithm selected")
		count, err := strconv.Atoi(reqData.MaxRecipes)
		if err != nil {
			conn.WriteJSON(map[string]interface{}{
				"error": "Invalid MaxRecipes value",
			})
			return
		}

		start := time.Now()
		paths := DfsMultiple(elementMap, basicElements,strings.ToLower(reqData.Target), count, conn)
		elapsed := time.Since(start)
		
		
		// err = conn.WriteJSON(map[string]interface{}{
		// 	"treeData": treeData,
		// })
		
		err = conn.WriteJSON(map[string]interface{}{
			"resultInfo": map[string]string{
				"time": elapsed.String(),
				"node": fmt.Sprintf("%d", len(paths)),
			},
			"status": "Completed",
		})
		if err != nil {
			log.Printf("Error sending final result: %v", err)
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "../public/tree.json")
	})

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}