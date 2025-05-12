package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"log"
	"net/http"
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
	MaxRecipes int `json:"maxRecipes"`
	LiveUpdate bool   `json:"liveUpdate"`
	Delay      int   `json:"delay"`
}

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
	Tier    int        `json:"tier"`
}

type TreeNode struct {
	Name      string     `json:"name"`
	Children  []TreeNode `json:"children,omitempty"`
	Highlight bool       `json:"highlight,omitempty"`
}

var basicElements = []string{"air", "earth", "fire", "water"}

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

	var nodesVisited int
	startTime := time.Now()
	fmt.Printf("Delay: %d\n", reqData.Delay)

	if reqData.Algorithm == "BFS" {

		conn.WriteJSON(map[string]interface{}{
			"status":  "Starting BFS",
			"message": "Initializing search algorithm",
		})

		if reqData.LiveUpdate {
			recipePlans = bfsMultipleLive(elementMap, strings.ToLower(reqData.Target), reqData.MaxRecipes, reqData.Delay, conn)
			log.Println("BFS Live Update")
		} else {
			recipePlans = bfsMultiple(elementMap, strings.ToLower(reqData.Target), reqData.MaxRecipes)
		}

	} else if reqData.Algorithm == "DFS" {

		conn.WriteJSON(map[string]interface{}{
			"status":  "Starting DFS",
			"message": "Initializing search algorithm",
		})

		if reqData.LiveUpdate {
			recipePlans, nodesVisited = dfsMultipleLive(elementMap, strings.ToLower(reqData.Target), reqData.MaxRecipes, reqData.Delay, conn)
		} else {
			recipePlans, nodesVisited = dfsMultiple(elementMap, strings.ToLower(reqData.Target), reqData.MaxRecipes)
		}
	} else if reqData.Algorithm == "BID" {

		conn.WriteJSON(map[string]interface{}{
			"status":  "Starting Bidirectional Search",
			"message": "Initializing search algorithm",
		})

		if reqData.LiveUpdate {
			recipePlans = bidirectionalSearchLive(elementMap, strings.ToLower(reqData.Target), reqData.MaxRecipes, reqData.Delay, conn)
		} else {
			recipePlans = bidirectionalSearch(elementMap, strings.ToLower(reqData.Target), reqData.MaxRecipes)
		}
	}

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

func isBasicElement(name string) bool {
	for _, b := range basicElements {
		if strings.ToLower(name) == strings.ToLower(b) {
			return true
		}
	}
	return false
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}