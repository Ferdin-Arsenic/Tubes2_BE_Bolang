package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var recipeLimitReached atomic.Bool

// Tipe-tipe data yang sudah ada sebelumnya
type RecipeStep struct {
	Element     string
	Ingredients []string
}

// Struktur untuk item dalam queue pencarian BFS
type bfsQueueItem struct {
	path  []RecipeStep
	open  map[string]bool // elemen yang masih memerlukan resep
	depth int
}

// Konstanta untuk pembatasan pencarian
const (
	bfsMaxDepth     = 500
	bfsMaxQueueSize = 100000
)

// Fungsi validasi resep
func isValidRecipe(a, b string, targetTier int, elementMap map[string]Element) bool {
	elemA, okA := elementMap[a]
	elemB, okB := elementMap[b]

	if !okA || !okB {
		return false
	}

	// Jika tier belum dikenali (-1), masih bisa dianggap valid
	if elemA.Tier < 0 || elemB.Tier < 0 {
		return true
	}
	return elemA.Tier < targetTier && elemB.Tier < targetTier
}

// Fungsi untuk membuat string key dari path
func pathToStringKey(steps []RecipeStep) string {
	keys := make([]string, 0, len(steps))
	for _, step := range steps {
		a := strings.ToLower(step.Ingredients[0])
		b := strings.ToLower(step.Ingredients[1])
		if a > b {
			a, b = b, a
		}
		keys = append(keys, fmt.Sprintf("%s:%s+%s", step.Element, a, b))
	}
	sort.Strings(keys)
	return strings.Join(keys, "|")
}

// Variabel global untuk melacak path yang sudah dilihat
var globalSeenPathKeys = make(map[string]bool)

// Fungsi untuk memeriksa duplikasi struktural
func isStructuralDuplicate(steps []RecipeStep, elementMap map[string]Element) bool {
	var normalized []string
	for _, step := range steps {
		a := strings.ToLower(step.Ingredients[0])
		b := strings.ToLower(step.Ingredients[1])
		e := strings.ToLower(step.Element)

		aTier := elementMap[a].Tier
		bTier := elementMap[b].Tier
		eTier := elementMap[e].Tier

		if a > b {
			a, b = b, a
			aTier, bTier = bTier, aTier
		}

		normalized = append(normalized,
			fmt.Sprintf("%s(%d):%s(%d)+%s(%d)", e, eTier, a, aTier, b, bTier))
	}

	sort.Strings(normalized)
	key := strings.Join(normalized, "|")

	return globalSeenPathKeys[key]
}

// Fungsi untuk membuat pohon dari langkah-langkah
func buildTreeFromSteps(root string, steps []RecipeStep, elementMap map[string]Element) TreeNode {
	recipeMap := make(map[string][]string)
	for _, step := range steps {
		recipeMap[step.Element] = step.Ingredients
	}
	tree := buildBFSRecipeTree(root, recipeMap, elementMap, make(map[string]bool), make(map[string]TreeNode))
	sortTreeChildren(&tree)
	return tree
}

// Fungsi untuk membuat pohon resep rekursif
func buildBFSRecipeTree(
	currentElement string,
	recipeMap map[string][]string,
	elementMap map[string]Element,
	visited map[string]bool,
	treeCache map[string]TreeNode,
) TreeNode {
	// Cek cache
	if cached, exists := treeCache[currentElement]; exists {
		return cached
	}

	// Elemen dasar
	if isBasicElement(currentElement) {
		node := TreeNode{Name: capitalize(currentElement)}
		treeCache[currentElement] = node
		return node
	}

	// Hindari siklus
	if visited[currentElement] {
		return TreeNode{Name: capitalize(currentElement)}
	}
	visited[currentElement] = true
	defer delete(visited, currentElement)

	// Cari bahan untuk elemen ini
	ingredients, exists := recipeMap[currentElement]
	if !exists || len(ingredients) != 2 {
		node := TreeNode{Name: capitalize(currentElement)}
		treeCache[currentElement] = node
		return node
	}

	// Buat anak-anak pohon
	childNodes := []TreeNode{
		buildBFSRecipeTree(strings.ToLower(ingredients[0]), recipeMap, elementMap, visited, treeCache),
		buildBFSRecipeTree(strings.ToLower(ingredients[1]), recipeMap, elementMap, visited, treeCache),
	}

	// Buat node saat ini
	node := TreeNode{
		Name:     capitalize(currentElement),
		Children: childNodes,
	}

	treeCache[currentElement] = node
	return node
}

// Fungsi untuk mengurutkan anak-anak pohon
func sortTreeChildren(node *TreeNode) {
	if len(node.Children) > 0 {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})
		for i := range node.Children {
			sortTreeChildren(&node.Children[i])
		}
	}
}

// Fungsi untuk membuat pohon menjadi string kanonik
func canonicalizeTree(node TreeNode) string {
	if len(node.Children) == 0 {
		return node.Name
	}
	var childrenStr []string
	for _, child := range node.Children {
		childrenStr = append(childrenStr, canonicalizeTree(child))
	}
	sort.Strings(childrenStr)
	return fmt.Sprintf("%s(%s)", node.Name, strings.Join(childrenStr, ","))
}

// Fungsi untuk membuat queue items awal
// Fungsi untuk membuat queue items awal
func createInitialBFSQueueItems(target string, elementMap map[string]Element) []bfsQueueItem {
	queue := []bfsQueueItem{}
	targetTier := elementMap[target].Tier

	for _, recipe := range elementMap[target].Recipes {
		if len(recipe) != 2 {
			continue
		}
		a := strings.ToLower(recipe[0])
		b := strings.ToLower(recipe[1])

		// Validasi resep
		if !isValidRecipe(a, b, targetTier, elementMap) {
			continue
		}

		// Buat langkah resep
		step := RecipeStep{Element: target, Ingredients: []string{a, b}}

		// Tentukan elemen yang masih perlu resep
		open := map[string]bool{}
		if !isBasicElement(a) {
			open[a] = true
		}
		if !isBasicElement(b) {
			open[b] = true
		}

		queue = append(queue, bfsQueueItem{
			path:  []RecipeStep{step},
			open:  open,
			depth: 1,
		})
	}
	return queue
}

// Fungsi untuk mengekspansi elemen terbuka
func expandOpenBFSElement(openElem string, curr bfsQueueItem, elementMap map[string]Element) []bfsQueueItem {
	newItems := []bfsQueueItem{}
	elemTier := elementMap[openElem].Tier

	for _, recipe := range elementMap[openElem].Recipes {
		if len(recipe) != 2 {
			continue
		}
		a := strings.ToLower(recipe[0])
		b := strings.ToLower(recipe[1])

		// Validasi resep
		if !isValidRecipe(a, b, elemTier, elementMap) {
			continue
		}

		// Buat langkah baru
		newStep := RecipeStep{Element: openElem, Ingredients: []string{a, b}}

		// Salin dan perbarui map elemen terbuka
		newOpen := copyOpenMap(curr.open)
		delete(newOpen, openElem)

		// Tambahkan elemen non-dasar ke dalam elemen terbuka
		if !isBasicElement(a) {
			newOpen[a] = true
		}
		if !isBasicElement(b) {
			newOpen[b] = true
		}

		// Buat path baru
		newPath := append([]RecipeStep{}, curr.path...)
		newPath = append(newPath, newStep)

		newItems = append(newItems, bfsQueueItem{
			path:  newPath,
			open:  newOpen,
			depth: curr.depth + 1,
		})
	}

	return newItems
}

// Fungsi pembantu untuk menyalin map
func copyOpenMap(orig map[string]bool) map[string]bool {
	newMap := make(map[string]bool, len(orig))
	for k, v := range orig {
		newMap[k] = v
	}
	return newMap
}

// Fungsi utama BFS dengan multithreading

func bfsMultiple(elementMap map[string]Element, target string, maxRecipes int) ([]TreeNode, int) {
	target = strings.ToLower(target)
	recipeLimitReached.Store(false)

	// Quick exits
	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}, 0
	}
	elem, exists := elementMap[target]
	if !exists || len(elem.Recipes) == 0 {
		return nil, 0
	}

	// Inisialisasi queue input
	initialQueue := createInitialBFSQueueItems(target, elementMap)
	queueChan := make(chan bfsQueueItem, bfsMaxQueueSize)
	resultChan := make(chan TreeNode, maxRecipes)

	var (
		nodeCounter      int64
		seenPathKeys     = make(map[string]bool)
		seenPathKeysLock = &sync.RWMutex{}
		treesLock        = &sync.Mutex{}
	)

	isDuplicatePath := func(path []RecipeStep) bool {
		key := pathToStringKey(path)
		seenPathKeysLock.Lock()
		defer seenPathKeysLock.Unlock()
		if seenPathKeys[key] || isStructuralDuplicate(path, elementMap) {
			return true
		}
		seenPathKeys[key] = true
		return false
	}

	// 1) Track outstanding tasks
	var taskCount int64 = int64(len(initialQueue))

	// 2) Seed initial tasks
	for _, item := range initialQueue {
		queueChan <- item
	}

	// 3) Monitor: tutup queueChan hanya jika taskCount==0
	go func() {
		for {
			if atomic.LoadInt64(&taskCount) == 0 {
				close(queueChan)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// 4) Worker pool
	var wg sync.WaitGroup
	numWorkers := min(8, runtime.NumCPU())
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for curr := range queueChan {
				// kita “mengambil” satu tugas
				atomic.AddInt64(&taskCount, -1)

				if curr.depth > bfsMaxDepth {
					continue
				}

				// jika sudah complete path-nya
				if len(curr.open) == 0 {
					if isDuplicatePath(curr.path) {
						continue
					}
					tree := buildTreeFromSteps(target, curr.path, elementMap)
					select {
					case resultChan <- tree:
						atomic.AddInt64(&nodeCounter, 1)
					default:
						// penuh → skip
					}
					continue
				}

				// expand satu open element per iterasi
				for openElem := range curr.open {
					newItems := expandOpenBFSElement(openElem, curr, elementMap)
					for _, item := range newItems {
						// hitung sub-tugas baru sebelum kirim
						atomic.AddInt64(&taskCount, 1)
						queueChan <- item
					}
					break
				}
			}
		}()
	}

	// 5) Setelah semua worker selesai, tutup resultChan
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 6) Kumpulkan hingga maxRecipes
	var (
		resultTrees    []TreeNode
		fingerprintSet = make(map[string]bool)
	)
	for tree := range resultChan {
		treesLock.Lock()
		if len(resultTrees) >= maxRecipes {
			treesLock.Unlock()
			break
		}
		fp := canonicalizeTree(tree)
		if !fingerprintSet[fp] {
			fingerprintSet[fp] = true
			resultTrees = append(resultTrees, tree)
		}
		treesLock.Unlock()

		if len(resultTrees) >= maxRecipes {
			recipeLimitReached.Store(true)
			break
		}
	}

	// 7) Sort untuk output deterministik
	sort.Slice(resultTrees, func(i, j int) bool {
		return canonicalizeTree(resultTrees[i]) < canonicalizeTree(resultTrees[j])
	})

	fmt.Println("Total node visited:", nodeCounter)
	return resultTrees, int(nodeCounter)
}

// Fungsi live update untuk BFS
func bfsMultipleLive(elementMap map[string]Element, target string, maxRecipes int, delay int, conn *websocket.Conn) []TreeNode {
	// Normalisasi target
	target = strings.ToLower(target)

	// Jalankan pencarian BFS
	trees, nodesVisited := bfsMultiple(elementMap, target, maxRecipes)

	// Kirim update live untuk setiap pohon
	for _, tree := range trees {
		conn.WriteJSON(map[string]interface{}{
			"status":       "Progress",
			"message":      "Finding trees",
			"treeData":     []TreeNode{tree},
			"nodesVisited": nodesVisited,
		})
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	return trees
}

// Fungsi utilitas tambahan untuk memeriksa apakah elemen ada dalam map
func containsElement(m map[string]bool, elem string) bool {
	_, exists := m[elem]
	return exists
}

// Tambahan fungsi-fungsi lain yang mungkin diperlukan
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Jika membutuhkan fungsi tambahan untuk memproses elemen atau map
func mergeElementMaps(maps ...map[string]Element) map[string]Element {
	result := make(map[string]Element)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
