package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func Scraping() {
	baseURL := "https://little-alchemy.fandom.com"
	listURL := baseURL + "/wiki/Elements_(Little_Alchemy_2)"

	fmt.Println("Starting Little Alchemy 2 recipe scraper...")

	// 1. Get list of elements from wiki
	elementsList, err := getElementsList(listURL)
	if err != nil {
		log.Fatalf("Failed to get elements list: %v", err)
	}

	// 2. Ensure basic elements are included & unique
	seen := map[string]bool{}
	for _, n := range elementsList {
		seen[strings.ToLower(n)] = true
	}
	for _, b := range basicElements {
		if !seen[b] {
			elementsList = append(elementsList, b)
			seen[b] = true
		}
	}

	fmt.Printf("Found %d elements to scrape\n", len(elementsList))
	if len(elementsList) > 0 {
		fmt.Println("Sample:", elementsList[:min(5, len(elementsList))])
	}

	// 3. Scrape recipes concurrently
	var elements []Element
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)
	var mu sync.Mutex

	for _, name := range elementsList {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			url := baseURL + "/wiki/" + strings.ReplaceAll(name, " ", "_")
			fmt.Printf("Scraping: %s\n", name)
			recs, err := scrapeRecipesLA2Only(url, name)
			if err != nil {
				log.Printf("  error on %s: %v\n", name, err)
				return
			}

			// Normalize element names in recipes
			normalizedRecs := normalizeRecipes(recs)

			mu.Lock()
			elements = append(elements, Element{
				Name:    name,
				Recipes: normalizedRecs,
				Tier:    -1, // will be filled later
			})
			mu.Unlock()

			time.Sleep(300 * time.Millisecond)
		}(name)
	}
	wg.Wait()

	// Debug: print elements and their recipes
	elementsWithNoRecipes := 0
	for _, e := range elements {
		if len(e.Recipes) == 0 && !contains(basicElements, strings.ToLower(e.Name)) {
			fmt.Printf("WARNING: %s has no recipes\n", e.Name)
			elementsWithNoRecipes++
		}
	}
	fmt.Printf("Elements with no recipes: %d\n", elementsWithNoRecipes)

	// Debug: basic elements check
	fmt.Println("Basic elements status:")
	for _, b := range basicElements {
		found := false
		for _, e := range elements {
			if strings.ToLower(e.Name) == strings.ToLower(b) {
				found = true
				break
			}
		}
		fmt.Printf("%s: %v\n", b, found)
	}

	// 4. Calculate Tiers with improved algorithm
	calcTiersFix(elements)

	// 5. Save JSON with Tier field
	outFile := "elements.json"
	if err := saveJSON(elements, outFile); err != nil {
		log.Fatalf("Failed saving %s: %v", outFile, err)
	}
	fmt.Printf("Done! Data with tiers in %s\n", outFile)

	// 6. Add tier analysis
	analyzeTiers(elements)
}

// Function to normalize recipes
func normalizeRecipes(recipes [][]string) [][]string {
	normalized := make([][]string, 0, len(recipes))
	for _, recipe := range recipes {
		if len(recipe) == 2 {
			// Clean and normalize element names
			a := strings.TrimSpace(recipe[0])
			b := strings.TrimSpace(recipe[1])

			// Remove special characters and extra spaces
			a = cleanElementName(a)
			b = cleanElementName(b)

			if a != "" && b != "" {
				normalized = append(normalized, []string{a, b})
			}
		}
	}
	return normalized
}

// Function to clean element name
func cleanElementName(name string) string {
	// Remove special characters and text in parentheses
	name = strings.Split(name, "(")[0]
	name = strings.TrimSpace(name)
	return name
}

func getElementsList(url string) ([]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	selectors := []string{
		"div#mw-content-text > div > ul > li > a",
		"div.mw-parser-output > ul > li > a",
		"div#mw-content-text table td a",
		"table.article-table a",
	}
	var elems []string
	seen := map[string]bool{}

	for _, sel := range selectors {
		doc.Find(sel).Each(func(_ int, s *goquery.Selection) {
			name := strings.TrimSpace(s.Text())
			lc := strings.ToLower(name)
			if name != "" && !strings.Contains(name, "Category:") && !seen[lc] {
				elems = append(elems, name)
				seen[lc] = true
			}
		})
		if len(elems) > 0 {
			break
		}
	}
	return elems, nil
}

// New function that only scrapes Little Alchemy 2 recipes
func scrapeRecipesLA2Only(url string, targetElement string) ([][]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	recipes := make([][]string, 0)
	seen := make(map[string]bool)

	// Track if we're in Little Alchemy 2 section
	// Track if we're in Little Alchemy 2 section
	inLA2Section := false
	inUsedInSection := false

	// Identify sections
	doc.Find("h1, h2, h3, h4").Each(func(_ int, h *goquery.Selection) {
		headerText := strings.ToLower(strings.TrimSpace(h.Text()))

		if strings.Contains(headerText, "little alchemy 2") {
			inLA2Section = true
			inUsedInSection = false
			fmt.Printf("Found Little Alchemy 2 section for: %s\n", targetElement)
		} else if strings.Contains(headerText, "little alchemy") && !strings.Contains(headerText, "2") {
			inLA2Section = false
			fmt.Printf("Found Little Alchemy 1 section for: %s (ignoring)\n", targetElement)
		} else if strings.Contains(headerText, "used in") {
			inUsedInSection = true
			fmt.Printf("Found 'Used in' section for: %s (ignoring)\n", targetElement)
		} else {
			// Reset flag if new unrelated header
			inLA2Section = false
			inUsedInSection = false
		}
	})

	// Process tables or lists, but only if inLA2Section == true
	doc.Find("div.mw-parser-output > *").Each(func(_ int, s *goquery.Selection) {
		if inLA2Section && !inUsedInSection {
			if s.Is("table") {
				s.Find("tr").Each(func(_ int, row *goquery.Selection) {
					cols := row.Find("td")
					if cols.Length() >= 3 {
						result := strings.TrimSpace(cols.Eq(2).Text())
						if strings.EqualFold(result, targetElement) {
							a := strings.TrimSpace(cols.Eq(0).Text())
							b := strings.TrimSpace(cols.Eq(1).Text())
							addRecipe(&recipes, seen, a, b)
						}
					}
				})
			}

			if s.Is("ul") {
				s.Find("li").Each(func(_ int, li *goquery.Selection) {
					text := li.Text()
					if strings.Contains(text, "+") || strings.Contains(text, "→") || strings.Contains(text, "=") {
						parseRecipeFromText(&recipes, seen, text, targetElement)
					}
				})
			}
		}
	})

	// 2. Process dedicated Little Alchemy 2 lists
	doc.Find("div.mw-parser-output > *").Each(func(i int, s *goquery.Selection) {
		// Check if this is under Little Alchemy 2 section
		var inLA2Context bool

		// Look backward for the closest heading
		prev := s.Prev()
		for prev.Length() > 0 {
			if prev.Is("h1, h2, h3, h4") {
				headerText := strings.ToLower(strings.TrimSpace(prev.Text()))
				inLA2Context = strings.Contains(headerText, "little alchemy 2")
				break
			}
			prev = prev.Prev()
		}

		if inLA2Context && s.Is("ul") && !inUsedInSection {
			s.Find("li").Each(func(_ int, li *goquery.Selection) {
				text := li.Text()
				if strings.Contains(text, "→") || strings.Contains(text, "=") || strings.Contains(text, "+") {
					parseRecipeFromText(&recipes, seen, text, targetElement)
				}
			})
		}
	})

	// 3. Check for "Little Alchemy 2" text in paragraphs for context
	if len(recipes) == 0 {
		inLA2Context := false

		doc.Find("div.mw-parser-output p").Each(func(_ int, p *goquery.Selection) {
			text := strings.ToLower(p.Text())

			// Check if paragraph mentions Little Alchemy 2
			if strings.Contains(text, "little alchemy 2") {
				inLA2Context = true
			}

			// If we're in LA2 context and paragraph has recipe indicators
			if inLA2Context && !inUsedInSection &&
				(strings.Contains(text, "recipe") || strings.Contains(text, "combine") ||
					strings.Contains(text, "make") || strings.Contains(text, "create")) {
				parseRecipeFromText(&recipes, seen, text, targetElement)
			}
		})
	}

	// 4. If we didn't find specific LA2 recipes but the page is about an LA2 element,
	// look for generic recipes that might apply to both games
	if len(recipes) == 0 {
		// Check page title or content indicators to confirm it's about LA2
		isLA2Page := false

		doc.Find("title, h1.page-header__title").Each(func(_ int, title *goquery.Selection) {
			if strings.Contains(strings.ToLower(title.Text()), "little alchemy 2") {
				isLA2Page = true
			}
		})

		if isLA2Page {
			// Get generic recipes as fallback
			fallbackRecipes := fallbackScrape(doc, targetElement)
			for _, recipe := range fallbackRecipes {
				addRecipe(&recipes, seen, recipe[0], recipe[1])
			}
		}
	}

	// Debug info
	fmt.Printf("Found %d Little Alchemy 2 recipes for %s\n", len(recipes), targetElement)

	return recipes, nil
}

func addRecipe(recipes *[][]string, seen map[string]bool, a, b string) {
	// Normalize and sort
	a = cleanElementName(a)
	b = cleanElementName(b)
	if a == "" || b == "" {
		return
	}

	// Sort for consistency
	if a > b {
		a, b = b, a
	}

	key := strings.ToLower(a + "|" + b)
	if !seen[key] {
		*recipes = append(*recipes, []string{a, b})
		seen[key] = true
	}
}

func parseRecipeFromText(recipes *[][]string, seen map[string]bool, text, targetElement string) {
	text = strings.TrimSpace(text)

	var parts []string
	if strings.Contains(text, "→") {
		parts = strings.Split(text, "→")
	} else if strings.Contains(text, "=") {
		parts = strings.Split(text, "=")
	} else if strings.Contains(text, "+") {
		// add this: if in list "Little Alchemy 2" has A + B (without →), assume it's a valid recipe
		parts = []string{text, targetElement}
	} else {
		return
	}

	if len(parts) < 2 {
		return
	}

	result := strings.TrimSpace(parts[1])
	if !strings.EqualFold(result, targetElement) {
		return
	}

	ingredients := strings.Split(parts[0], "+")
	if len(ingredients) == 2 {
		a := strings.TrimSpace(ingredients[0])
		b := strings.TrimSpace(ingredients[1])
		addRecipe(recipes, seen, a, b)
	}
}

func fallbackScrape(doc *goquery.Document, targetElement string) [][]string {
	recipes := make([][]string, 0)
	seen := make(map[string]bool)

	// Generic tables - only use if page context confirms LA2
	doc.Find("table.wikitable, table.article-table").Each(func(_ int, table *goquery.Selection) {
		// Skip tables in "Used in" sections
		inUsedSection := false
		prev := table.Prev()
		for prev.Length() > 0 && !prev.Is("h1, h2, h3, h4") {
			prev = prev.Prev()
		}
		if prev.Length() > 0 {
			sectionHeader := strings.ToLower(strings.TrimSpace(prev.Text()))
			if strings.Contains(sectionHeader, "used in") {
				inUsedSection = true
			} else if strings.Contains(sectionHeader, "little alchemy") &&
				!strings.Contains(sectionHeader, "little alchemy 2") {
				// Skip LA1 tables
				inUsedSection = true
			}
		}

		if !inUsedSection {
			table.Find("tr").Each(func(_ int, row *goquery.Selection) {
				cols := row.Find("td")
				if cols.Length() >= 3 {
					result := strings.TrimSpace(cols.Eq(2).Text())
					if strings.EqualFold(result, targetElement) {
						a := strings.TrimSpace(cols.Eq(0).Text())
						b := strings.TrimSpace(cols.Eq(1).Text())
						key := strings.ToLower(a + "|" + b)
						if a != "" && b != "" && !seen[key] {
							recipes = append(recipes, []string{a, b})
							seen[key] = true
						}
					}
				}
			})
		}
	})

	return recipes
}

// ================ MAIN IMPROVEMENTS ================
// New algorithm to calculate tiers
func calcTiersFix(elements []Element) {
	fmt.Println("\n=== Using optimized tier algorithm ===")

	// Step 1: Create required mappings
	elementMap := make(map[string]*Element)       // element name -> Element pointer
	elementRecipes := make(map[string][][]string) // element name -> its recipes

	// Convert all element names to lowercase for consistency
	for i := range elements {
		lowerName := strings.ToLower(elements[i].Name)
		elementMap[lowerName] = &elements[i]

		// Create copy of recipes with normalized names
		var normalizedRecipes [][]string
		for _, recipe := range elements[i].Recipes {
			if len(recipe) == 2 {
				a := strings.ToLower(recipe[0])
				b := strings.ToLower(recipe[1])
				normalizedRecipes = append(normalizedRecipes, []string{a, b})
			}
		}
		elementRecipes[lowerName] = normalizedRecipes
	}

	// Step 2: Set tier 0 for basic elements
	for _, basic := range basicElements {
		if elem, exists := elementMap[basic]; exists {
			elem.Tier = 0
		} else {
			fmt.Printf("ERROR: Basic element %s missing from map\n", basic)
		}
	}

	// Step 3: Create dependency graph to determine calculation order
	dependencies := make(map[string]map[string]bool) // element -> {dependencies}
	for name, recipes := range elementRecipes {
		deps := make(map[string]bool)
		for _, recipe := range recipes {
			if len(recipe) == 2 {
				deps[recipe[0]] = true
				deps[recipe[1]] = true
			}
		}
		dependencies[name] = deps
	}

	// Step 4: Iterate to calculate tiers - using dynamic programming approach
	// maxIterations to avoid infinite loop if there are cyclic dependencies
	changed := true
	iteration := 0
	maxIterations := 100 // Maximum iterations to avoid infinite loop

	for changed && iteration < maxIterations {
		changed = false
		iteration++

		for elemName, elem := range elementMap {
			// Skip elements that already have a tier or don't have recipes
			if elem.Tier != -1 || len(elementRecipes[elemName]) == 0 {
				continue
			}

			// For each recipe, check if all dependencies have tiers
			for _, recipe := range elementRecipes[elemName] {
				if len(recipe) != 2 {
					continue
				}

				// Find maximum tier from recipe ingredients
				maxIngredientTier := -1
				allIngredientsHaveTier := true

				for _, ingredient := range recipe {
					if ingElem, exists := elementMap[ingredient]; exists && ingElem.Tier != -1 {
						if ingElem.Tier > maxIngredientTier {
							maxIngredientTier = ingElem.Tier
						}
					} else {
						allIngredientsHaveTier = false
						break
					}
				}

				// If all ingredients have a tier, calculate this element's tier
				if allIngredientsHaveTier {
					elem.Tier = maxIngredientTier + 1
					changed = true
					break // Go to next element
				}
			}
		}

		fmt.Printf("Iteration %d: Updated %d elements\n", iteration, countUpdatedElements(elements))
	}

	// Step 5: Set tier for elements without recipes (except basic elements)
	// Elements without recipes might be additional basic elements or final elements
	missingRecipesCount := 0
	for _, elem := range elements {
		if elem.Tier == -1 && len(elem.Recipes) == 0 && !isBasicElement(elem.Name) {
			// Elements without recipes that aren't basic elements, set tier to 999 as marker
			elem.Tier = 999
			missingRecipesCount++
		}
	}

	// Step 6: Set tier for other elements still at -1 to 998 to mark a problem
	unresolvableCount := 0
	for _, elem := range elements {
		if elem.Tier == -1 {
			elem.Tier = 998 // Can't determine tier
			unresolvableCount++
		}
	}

	// Report
	fmt.Printf("\nTier calculation completed in %d iterations\n", iteration)
	fmt.Printf("Elements with missing recipes: %d\n", missingRecipesCount)
	fmt.Printf("Elements with unresolvable tiers: %d\n", unresolvableCount)
}

// Helper to count elements that have been updated (tier != -1)
func countUpdatedElements(elements []Element) int {
	count := 0
	for _, e := range elements {
		if e.Tier != -1 {
			count++
		}
	}
	return count
}

// Function to analyze tier distribution
func analyzeTiers(elements []Element) {
	// Counting tiers
	tierCounts := make(map[int]int)
	for _, elem := range elements {
		tierCounts[elem.Tier]++
	}

	// Reporting
	fmt.Println("\n--- Tier Distribution Analysis ---")
	fmt.Printf("Total elements: %d\n", len(elements))

	// Sort and print tier counts
	fmt.Println("Tier distribution:")

	// First print normal tiers (0-20)
	totalNormalTiers := 0
	for t := 0; t <= 20; t++ {
		if count, ok := tierCounts[t]; ok && count > 0 {
			fmt.Printf("  Tier %d: %d elements\n", t, count)
			totalNormalTiers += count
		}
	}

	// Then print special tiers
	if count, ok := tierCounts[-1]; ok && count > 0 {
		fmt.Printf("  Tier -1 (unprocessed): %d elements\n", count)
	}
	if count, ok := tierCounts[998]; ok && count > 0 {
		fmt.Printf("  Tier 998 (unresolvable dependencies): %d elements\n", count)
	}
	if count, ok := tierCounts[999]; ok && count > 0 {
		fmt.Printf("  Tier 999 (no recipes): %d elements\n", count)
	}

	fmt.Printf("\nElements with normal tiers (0-20): %d (%.1f%%)\n",
		totalNormalTiers, float64(totalNormalTiers)/float64(len(elements))*100)

	// Sample each category
	categories := []struct {
		name string
		tier int
		max  int
	}{
		{"Normal tier 1", 1, 5},
		{"Normal tier 2", 2, 5},
		{"Normal tier 3", 3, 5},
		{"Unresolvable dependencies (998)", 998, 5},
		{"No recipes (999)", 999, 5},
	}

	for _, cat := range categories {
		if tierCounts[cat.tier] > 0 {
			fmt.Printf("\nSample of %s elements:\n", cat.name)
			count := 0
			for _, elem := range elements {
				if elem.Tier == cat.tier {
					recipeStr := "no recipes"
					if len(elem.Recipes) > 0 {
						ingredients := elem.Recipes[0]
						if len(ingredients) >= 2 {
							recipeStr = fmt.Sprintf("%s + %s", ingredients[0], ingredients[1])
						}
					}
					fmt.Printf("  - %s (%s)\n", elem.Name, recipeStr)
					count++
					if count >= cat.max {
						break
					}
				}
			}
		}
	}

	// Specifically for tier 1 elements, show all
	fmt.Println("\nAll Tier 1 elements:")
	for _, elem := range elements {
		if elem.Tier == 1 {
			recipeStr := "no recipes"
			if len(elem.Recipes) > 0 {
				ingredients := elem.Recipes[0]
				if len(ingredients) >= 2 {
					recipeStr = fmt.Sprintf("%s + %s", ingredients[0], ingredients[1])
				}
			}
			fmt.Printf("  - %s (%s)\n", elem.Name, recipeStr)
		}
	}
}

func saveJSON(elements []Element, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(elements)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Function to check if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
