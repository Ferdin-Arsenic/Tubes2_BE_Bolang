package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Struktur data yang sudah ada di files lain tidak perlu didefinisikan ulang
// Kami menggunakan type RecipeStep yang baru untuk pencarian BFS multi-thread

type RecipeStep struct {
	Element     string
	Ingredients []string
}

type BuildQueueItem struct {
	Path  []RecipeStep
	Open  map[string]bool
	Depth int
}

// Struktur data thread-safe untuk multithreading BFS
type SafeQueue struct {
	queue []BuildQueueItem
	mutex sync.Mutex
}

type SafeResults struct {
	trees        []TreeNode
	fingerprints map[string]bool
	mutex        sync.Mutex
}

type SafePathKeys struct {
	keys  map[string]bool
	mutex sync.RWMutex
}

type Counter struct {
	count int
	mutex sync.Mutex
}

// Konstanta untuk pencarian
const (
	bfsMaxDepth     = 500
	bfsMaxQueueSize = 100000
	numWorkers      = 16 // Jumlah worker thread
	batchSize       = 50 // Ukuran batch per worker
)

// Metode untuk Counter
func (c *Counter) Increment() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.count++
}

func (c *Counter) Get() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.count
}

// Metode untuk SafeQueue
func newSafeQueue() *SafeQueue {
	return &SafeQueue{
		queue: []BuildQueueItem{},
	}
}

func (sq *SafeQueue) Push(items ...BuildQueueItem) {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()
	sq.queue = append(sq.queue, items...)
}

func (sq *SafeQueue) Pop(count int) []BuildQueueItem {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	if len(sq.queue) == 0 {
		return nil
	}

	if count > len(sq.queue) {
		count = len(sq.queue)
	}

	items := sq.queue[:count]
	sq.queue = sq.queue[count:]
	return items
}

func (sq *SafeQueue) Length() int {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()
	return len(sq.queue)
}

func (sq *SafeQueue) PruneLarge() {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	if len(sq.queue) > bfsMaxQueueSize {
		sort.Slice(sq.queue, func(i, j int) bool {
			return sq.queue[i].Depth < sq.queue[j].Depth
		})
		sq.queue = sq.queue[:bfsMaxQueueSize]
	}
}

func (sq *SafeQueue) PruneLargeWithPriority() {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	if len(sq.queue) > bfsMaxQueueSize {
		sort.Slice(sq.queue, func(i, j int) bool {
			return sq.queue[i].Depth+len(sq.queue[i].Open) < sq.queue[j].Depth+len(sq.queue[j].Open)
		})
		sq.queue = sq.queue[:bfsMaxQueueSize]
	}
}

// Metode untuk SafeResults
func newSafeResults(maxSize int) *SafeResults {
	return &SafeResults{
		trees:        make([]TreeNode, 0, maxSize),
		fingerprints: make(map[string]bool),
	}
}

func (sr *SafeResults) Add(tree TreeNode, fingerprint string) bool {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if sr.fingerprints[fingerprint] {
		return false
	}

	if len(sr.trees) >= cap(sr.trees) {
		return false
	}

	sr.fingerprints[fingerprint] = true
	sr.trees = append(sr.trees, tree)
	return true
}

func (sr *SafeResults) IsFull() bool {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	return len(sr.trees) >= cap(sr.trees)
}

func (sr *SafeResults) GetTrees() []TreeNode {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	return sr.trees
}

func (sr *SafeResults) Count() int {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	return len(sr.trees)
}

// Metode untuk SafePathKeys
func newSafePathKeys() *SafePathKeys {
	return &SafePathKeys{
		keys: make(map[string]bool),
	}
}

func (spk *SafePathKeys) Check(key string) bool {
	spk.mutex.RLock()
	defer spk.mutex.RUnlock()
	return spk.keys[key]
}

func (spk *SafePathKeys) Add(key string) {
	spk.mutex.Lock()
	defer spk.mutex.Unlock()
	spk.keys[key] = true
}

// Fungsi utama BFS multithreading
func bfsMultiple(elementMap map[string]Element, target string, maxRecipes int) ([]TreeNode, int) {
	target = strings.ToLower(target)

	counter := &Counter{}

	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}, counter.Get()
	}

	if elem, ok := elementMap[target]; !ok || len(elem.Recipes) == 0 {
		return []TreeNode{}, counter.Get()
	}

	trees := bfsBuildRecipeTreesParallel(target, elementMap, maxRecipes, counter)
	fmt.Printf("Total nodes visited: %d\n", counter.Get())
	return trees, counter.Get()
}

func bfsBuildRecipeTreesParallel(target string, elementMap map[string]Element, maxRecipes int, counter *Counter) []TreeNode {
	queue := newSafeQueue()
	results := newSafeResults(maxRecipes)
	pathKeys := newSafePathKeys()

	// Inisialisasi queue dengan recipe awal
	queue.Push(createInitialQueueItems(target, elementMap)...)

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Membuat worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(queue, results, pathKeys, elementMap, target, counter, &wg, done)
	}

	// Goroutine untuk memonitor kondisi selesai
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if queue.Length() == 0 || results.IsFull() {
				close(done) // Signal all workers to finish
				break
			}
		}
	}()

	wg.Wait()
	return results.GetTrees()
}

func worker(queue *SafeQueue, results *SafeResults, pathKeys *SafePathKeys,
	elementMap map[string]Element, target string, counter *Counter,
	wg *sync.WaitGroup, done chan struct{}) {
	defer wg.Done()

	for {
		select {
		case <-done:
			return
		default:
			items := queue.Pop(batchSize)
			if items == nil || len(items) == 0 {
				runtime.Gosched()
				continue
			}

			for _, curr := range items {
				counter.Increment()

				if curr.Depth > bfsMaxDepth {
					continue
				}

				if len(curr.Open) == 0 {
					key := pathToStringKey(curr.Path)
					if pathKeys.Check(key) || isStructuralDuplicate(curr.Path, elementMap, pathKeys) {
						continue
					}
					pathKeys.Add(key)

					fp := canonicalizeSteps(curr.Path, elementMap)
					if !results.Add(TreeNode{}, fp) {
						continue
					}

					tree := buildTreeFromSteps(target, curr.Path, elementMap)
					results.mutex.Lock()
					results.trees[len(results.trees)-1] = tree
					results.mutex.Unlock()

					if results.IsFull() {
						return
					}
					continue
				}

				// Expand first open element
				for openElem := range curr.Open {
					queue.Push(expandOpenElement(openElem, curr, elementMap)...)
					break
				}

				queue.PruneLargeWithPriority()
			}
		}
	}
}

func isValidRecipe(a, b string, targetTier int, elementMap map[string]Element) bool {
	elemA, okA := elementMap[a]
	elemB, okB := elementMap[b]

	if !okA || !okB {
		return false
	}
	// ❗ Jika tier belum dikenali (-1), masih bisa dianggap valid
	if elemA.Tier < 0 || elemB.Tier < 0 {
		return true
	}
	return elemA.Tier < targetTier && elemB.Tier < targetTier
}

func createInitialQueueItems(target string, elementMap map[string]Element) []BuildQueueItem {
	queue := []BuildQueueItem{}
	targetTier := elementMap[target].Tier

	for _, recipe := range elementMap[target].Recipes {
		if len(recipe) != 2 {
			continue
		}
		a := strings.ToLower(recipe[0])
		b := strings.ToLower(recipe[1])
		if !isValidRecipe(a, b, targetTier, elementMap) {
			continue
		}

		step := RecipeStep{Element: target, Ingredients: []string{a, b}}
		open := map[string]bool{}
		if !isBasicElement(a) {
			open[a] = true
		}
		if !isBasicElement(b) {
			open[b] = true
		}

		queue = append(queue, BuildQueueItem{
			Path:  []RecipeStep{step},
			Open:  open,
			Depth: 1,
		})
	}
	return queue
}

func expandOpenElement(openElem string, curr BuildQueueItem, elementMap map[string]Element) []BuildQueueItem {
	newItems := []BuildQueueItem{}
	elemTier := elementMap[openElem].Tier

	for _, recipe := range elementMap[openElem].Recipes {
		if len(recipe) != 2 {
			continue
		}
		a := strings.ToLower(recipe[0])
		b := strings.ToLower(recipe[1])
		if !isValidRecipe(a, b, elemTier, elementMap) {
			continue
		}

		newStep := RecipeStep{Element: openElem, Ingredients: []string{a, b}}
		newOpen := copyOpenMap(curr.Open)
		delete(newOpen, openElem)
		if !isBasicElement(a) {
			newOpen[a] = true
		}
		if !isBasicElement(b) {
			newOpen[b] = true
		}

		newPath := append([]RecipeStep{}, curr.Path...)
		newPath = append(newPath, newStep)

		newItems = append(newItems, BuildQueueItem{
			Path:  newPath,
			Open:  newOpen,
			Depth: curr.Depth + 1,
		})
	}

	return newItems
}

func copyOpenMap(orig map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range orig {
		newMap[k] = v
	}
	return newMap
}

func buildTreeFromSteps(root string, steps []RecipeStep, elementMap map[string]Element) TreeNode {
	recipeMap := make(map[string][]string)
	for _, step := range steps {
		recipeMap[step.Element] = step.Ingredients
	}

	// Menggunakan fungsi buildRecipeTree yang sudah ada di treebuilder.go
	visitedMap := make(map[string]bool)
	memoizeMap := make(map[string]TreeNode)
	tree := buildRecipeTree(root, recipeMap, elementMap, visitedMap, memoizeMap)
	sortTreeChildren(&tree)
	return tree
}

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

func isStructuralDuplicate(steps []RecipeStep, elementMap map[string]Element, pathKeys *SafePathKeys) bool {
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

	if pathKeys.Check(key) {
		return true
	}
	pathKeys.Add(key)
	return false
}

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

// Fungsi live update untuk WebSocket
func bfsMultipleLive(elementMap map[string]Element, target string, maxRecipes int, delay int, conn *websocket.Conn) []TreeNode {
	target = strings.ToLower(target)
	counter := &Counter{}
	pathKeys := newSafePathKeys()
	results := newSafeResults(maxRecipes)

	if isBasicElement(target) {
		return []TreeNode{{Name: capitalize(target)}}
	}

	if elem, ok := elementMap[target]; !ok || len(elem.Recipes) == 0 {
		return []TreeNode{}
	}

	queue := newSafeQueue()
	queue.Push(createInitialQueueItems(target, elementMap)...)

	conn.WriteJSON(map[string]interface{}{
		"status":       "Starting",
		"message":      "Initializing single-threaded BFS...",
		"treeData":     []TreeNode{},
		"nodesVisited": 0,
	})

	previewSent := make(map[string]bool)

	for queue.Length() > 0 && !results.IsFull() {
		items := queue.Pop(1)
		if len(items) == 0 {
			break
		}
		curr := items[0]
		counter.Increment()

		if len(curr.Path) > 0 {
			lastStep := curr.Path[len(curr.Path)-1]
			previewKey := pathToStringKey([]RecipeStep{lastStep})
			if !previewSent[previewKey] {
				tree := buildTreeFromSteps(lastStep.Element, []RecipeStep{lastStep}, elementMap)
				conn.WriteJSON(map[string]interface{}{
					"status":       "Preview",
					"message":      "Discovered: " + capitalize(lastStep.Element),
					"treeData":     []TreeNode{tree},
					"nodesVisited": counter.Get(),
				})
				time.Sleep(time.Duration(delay) * time.Millisecond)
				previewSent[previewKey] = true
			}
		}

		if curr.Depth > bfsMaxDepth {
			continue
		}

		if len(curr.Open) == 0 {
			key := pathToStringKey(curr.Path)
			if pathKeys.Check(key) || isStructuralDuplicate(curr.Path, elementMap, pathKeys) {
				continue
			}
			pathKeys.Add(key)

			fp := canonicalizeSteps(curr.Path, elementMap)
			if !results.Add(TreeNode{}, fp) {
				continue
			}

			tree := buildTreeFromSteps(target, curr.Path, elementMap)
			results.trees[len(results.trees)-1] = tree

			conn.WriteJSON(map[string]interface{}{
				"status":       "Final",
				"message":      "Final tree found!",
				"treeData":     []TreeNode{tree},
				"nodesVisited": counter.Get(),
			})
			time.Sleep(time.Duration(delay) * time.Millisecond)
			continue
		}

		for openElem := range curr.Open {
			queue.Push(expandOpenElement(openElem, curr, elementMap)...)
			break
		}

		queue.PruneLargeWithPriority()
	}

	conn.WriteJSON(map[string]interface{}{
		"status":       "Completed",
		"message":      fmt.Sprintf("Found %d recipes, explored %d nodes", results.Count(), counter.Get()),
		"nodesVisited": counter.Get(),
	})

	return results.GetTrees()
}

func canonicalizeSteps(steps []RecipeStep, elementMap map[string]Element) string {
	var normalized []string
	for _, step := range steps {
		a := strings.ToLower(step.Ingredients[0])
		b := strings.ToLower(step.Ingredients[1])
		e := strings.ToLower(step.Element)
		if a > b {
			a, b = b, a
		}
		normalized = append(normalized, fmt.Sprintf("%s:%s+%s", e, a, b))
	}
	sort.Strings(normalized)
	return strings.Join(normalized, "|")
}
