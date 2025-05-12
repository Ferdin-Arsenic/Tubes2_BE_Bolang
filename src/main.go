package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	// "sync"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var elementMap map[string]Element

type RequestData struct {
	Algorithm  string `json:"algorithm"`
	Target     string `json:"target"`
	MaxRecipes string `json:"maxRecipes"`
	LiveUpdate bool   `json:"liveUpdate"`
	Delay      int   `json:"delay"`
}

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
		"status":  "Processing",
		"message": "Loading elements data",
	})

	data, err := ioutil.ReadFile("data/elements.json")
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

	elementMap = make(map[string]Element)
	for _, e := range elements {
		elementMap[strings.ToLower(e.Name)] = e
	}

	var recipePlans []TreeNode

	maxRecipeInput, err := strconv.Atoi(reqData.MaxRecipes)
	if err != nil || maxRecipeInput <= 0 {
		conn.WriteJSON(map[string]interface{}{
			"error": "Invalid MaxRecipes value",
		})
		return
	}

	var nodesVisited int
	startTime := time.Now()
	fmt.Printf("Delay: %d\n", reqData.Delay)

	if reqData.Algorithm == "BFS" {

		conn.WriteJSON(map[string]interface{}{
			"status":  "Starting BFS",
			"message": "Initializing search algorithm",
		})
		recipePlans = bfsMultiple(elementMap, strings.ToLower(reqData.Target), maxRecipeInput)

	} else if reqData.Algorithm == "DFS" {

		conn.WriteJSON(map[string]interface{}{
			"status":  "Starting DFS",
			"message": "Initializing search algorithm",
		})

		if reqData.LiveUpdate {
			recipePlans, nodesVisited = dfsMultipleLive(strings.ToLower(reqData.Target), maxRecipeInput, reqData.Delay, conn)
		} else {
			recipePlans, nodesVisited = dfsMultiple(strings.ToLower(reqData.Target), maxRecipeInput)
		}
	}
	// } else if reqData.Algorithm == "BID" {

	// 	conn.WriteJSON(map[string]interface{}{
	// 		"status":  "Starting Bidirectional",
	// 		"message": "Initializing search algorithm",
	// 	})
	// 	recipePlans = bidirectionalMultiple(elementMap, strings.ToLower(reqData.Target), maxRecipeInput)
	// }

	elapsed := time.Since(startTime)
	fmt.Printf("Ditemukan %d resep via %s.\n", len(recipePlans), reqData.Algorithm)
	
	if len(recipePlans) == 0 {
		conn.WriteJSON(map[string]interface{}{
			"status":  "Completed",
			"message": "No recipe plans found",
		})
		return
	}

	fmt.Println("Waktu eksekusi: ", elapsed)

	conn.WriteJSON(map[string]interface{}{
		"status":   "Completed",
		"message":  fmt.Sprintf("Found %d recipe plans", len(recipePlans)),
		"duration": elapsed.String(),
		"treeData": recipePlans,
		"nodes": nodesVisited,
	})
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

func mainCli() {
	data, err := ioutil.ReadFile("data/elements.json")
	if err != nil {
		log.Fatalf("Failed to read elements.json: %v", err)
	}

	var elements []Element
	if err := json.Unmarshal(data, &elements); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	elementMap := make(map[string]Element)
	for _, e := range elements {
		elementMap[strings.ToLower(e.Name)] = e
	}

	var target string
	fmt.Print("Masukkan target element: ")
	fmt.Scanln(&target)
	target = strings.ToLower(target)

	fmt.Print("Pilih mode (1 = shortest/single, 2 = multiple): ")
	var mode int
	fmt.Scanln(&mode)

	fmt.Print("Pilih algoritma (1 = bfs, 2 = dfs, 3 = bidirectional): ")
	var algo int
	fmt.Scanln(&algo)

	fmt.Print("Apakah ingin tree detail sampai ke elemen dasar? (y/n): ")
	var detailInput string
	fmt.Scanln(&detailInput)

	startTime := time.Now()
	var algoName string
	switch algo {
	case 1:
		algoName = "bfs"
	case 2:
		algoName = "dfs"
	case 3:
		algoName = "bidirectional"
	default:
		algoName = "unknown"
	}

	if mode == 1 {
		var foundRecipePlans []TreeNode
		var recipePlan map[string][]string

		if algo == 1 {
			fmt.Println("BFS untuk mode shortest belum diimplementasikan dengan struktur resep baru.")
			return
		} else {
			// foundRecipePlans, nodesVisited = dfsMultiple(elementMap, target, 1)
		}

		if recipePlan == nil {
			fmt.Println("Tidak ditemukan resep untuk", target)
			return
		}

		fmt.Println("Resep ditemukan (via DFS):")
		writeJSON(foundRecipePlans, target+"_single_dfs.json")
		fmt.Println("Tree saved to", target+"_single_dfs.json")

	} else if mode == 2 {
		var recipePlans []TreeNode
		var maxRecipeInput int
		fmt.Print("Masukkan maksimal recipe: ")
		fmt.Scanln(&maxRecipeInput)

		fmt.Printf("Mencari resep untuk %s dengan algoritma %s...\n", target, algoName)
		fmt.Print("Pilih sumber (1 = explicit dari file, 2 = pencarian traversal): ")
		var source int
		fmt.Scanln(&source)

		fmt.Printf("Ditemukan %d resep via %s.\n", len(recipePlans), algoName)
		if len(recipePlans) == 0 {
			return
		}

		// for i := range len(recipePlans) {
		// 	recipePrinter(recipePlans[i])
		// }
		fmt.Println("Waktu eksekusi: ", time.Since(startTime))

		if len(recipePlans) > 0 {
			writeJSON(recipePlans, target+"_multiple_dfs.json")
			fmt.Println("Semua tree tersimpan di", target+"_multiple_dfs.json")
		} else {
			fmt.Println("Tidak ada tree yang dihasilkan.")
		}
	} else {
		fmt.Println("Mode tidak dikenali.")
	}
}

func getExplicitRecipes(target string, elementMap map[string]Element, max int) []map[string][]string {
	var result []map[string][]string
	recipes := elementMap[strings.ToLower(target)].Recipes

	for i := 0; i < len(recipes) && i < max; i++ {
		recipe := recipes[i]
		if len(recipe) != 2 {
			continue
		}
		plan := map[string][]string{
			target: {
				strings.ToLower(recipe[0]),
				strings.ToLower(recipe[1]),
			},
		}
		result = append(result, plan)
	}
	return result
}

func copyMap(original map[string][]string) map[string][]string {
	copied := make(map[string][]string)
	for k, v := range original {
		copied[k] = append([]string{}, v...)
	}
	return copied
}

func recipePrinter(recipe map[string][]string) {
	// Enhanced printing with better formatting
	fmt.Println("Printing recipe:")

	for element, ingredients := range recipe {
		fmt.Printf("  To make %s, combine:\n", element)

		if len(ingredients) == 2 {
			fmt.Printf("    - %s\n    - %s\n", ingredients[0], ingredients[1])
		} else {
			fmt.Println("    (Invalid recipe format)")
		}
	}
	fmt.Println() // Empty line after recipe
}
