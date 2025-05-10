package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func Scraping() {
	baseURL := "https://little-alchemy.fandom.com"
	listURL := baseURL + "/wiki/Elements_(Little_Alchemy_2)"

	fmt.Println("Starting Little Alchemy 2 recipe scraper...")

	// 1. Ambil daftar elemen dari wiki
	elementsList, err := getElementsList(listURL)
	if err != nil {
		log.Fatalf("Failed to get elements list: %v", err)
	}

	// 2. Pastikan elemen dasar ada & unik
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

	// 3. Scrape recipes secara concurrent
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
			recs, err := scrapeRecipes(url, name)
			if err != nil {
				log.Printf("  error on %s: %v\n", name, err)
				return
			}

			// Normalisasi nama elemen dalam resep
			normalizedRecs := normalizeRecipes(recs)

			mu.Lock()
			elements = append(elements, Element{
				Name:    name,
				Recipes: normalizedRecs,
				Tier:    -1, // nanti diisi
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

	// 4. Hitung Tier dengan algoritma yang diperbaiki
	calcTiersFix(elements)

	// 5. Simpan JSON dengan field Tier
	outFile := "elements.json"
	if err := saveJSON(elements, outFile); err != nil {
		log.Fatalf("Failed saving %s: %v", outFile, err)
	}
	fmt.Printf("Done! Data with tiers in %s\n", outFile)

	// 6. Tambahkan analisis tier
	analyzeTiers(elements)
}

// Fungsi tambahan untuk normalisasi resep
func normalizeRecipes(recipes [][]string) [][]string {
	normalized := make([][]string, 0, len(recipes))
	for _, recipe := range recipes {
		if len(recipe) == 2 {
			// Bersihkan dan normalisasi nama elemen
			a := strings.TrimSpace(recipe[0])
			b := strings.TrimSpace(recipe[1])

			// Hilangkan karakter khusus dan extra spaces
			a = cleanElementName(a)
			b = cleanElementName(b)

			if a != "" && b != "" {
				normalized = append(normalized, []string{a, b})
			}
		}
	}
	return normalized
}

// Fungsi untuk membersihkan nama elemen
func cleanElementName(name string) string {
	// Remove any text in parentheses
	if idx := strings.Index(name, "("); idx != -1 {
		name = name[:idx]
	}

	// Remove any text after special characters that might appear in descriptions
	for _, char := range []string{"→", "=", ":", "/", "-", ",", ";"} {
		if idx := strings.Index(name, char); idx != -1 {
			name = name[:idx]
		}
	}

	// Trim leading/trailing whitespace
	name = strings.TrimSpace(name)

	// Remove any non-essential characters
	name = strings.TrimRight(name, ".,;:•·-")

	// Remove special characters used in the wiki
	name = strings.ReplaceAll(name, "•", "")
	name = strings.ReplaceAll(name, "·", "")

	// Remove common patterns in the wiki text that aren't part of element names
	patterns := []string{
		"Recipe:", "Recipes:", "Created from:", "Makes:",
		"Created by:", "Created with:", "Combines with:",
		"How to make", "Made from", "Created using",
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(pattern)) {
			name = name[len(pattern):]
			name = strings.TrimSpace(name)
		}
	}

	// Second round of whitespace trimming
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

func scrapeRecipes(url string, targetElement string) ([][]string, error) {
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

	// First, check if we're on the correct page for the element
	pageTitle := doc.Find("h1.page-header__title").Text()
	pageTitle = strings.TrimSpace(pageTitle)

	fmt.Printf("Scraping page for '%s', page title: '%s'\n", targetElement, pageTitle)

	// ==================== APPROACH 1: Find recipes in bullet points ====================
	// Look for the bullet point format common on element pages
	doc.Find("ul li").Each(func(_ int, li *goquery.Selection) {
		liText := li.Text()
		if strings.Contains(liText, "+") {
			parts := strings.Split(liText, "+")
			if len(parts) == 2 {
				a := strings.TrimSpace(parts[0])
				b := strings.TrimSpace(parts[1])

				// Clean the element names
				a = cleanElementName(a)
				b = cleanElementName(b)

				if a != "" && b != "" {
					key := strings.ToLower(a + "|" + b)
					if !seen[key] {
						recipes = append(recipes, []string{a, b})
						seen[key] = true
					}
				}
			}
		}
	})

	// ==================== APPROACH 2: Find recipes in tables ====================
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			// First check if this is a recipe table
			cells := row.Find("td")

			// Check if this looks like a recipe row (ingredients and result)
			if cells.Length() >= 3 {
				a := strings.TrimSpace(cells.Eq(0).Text())
				b := strings.TrimSpace(cells.Eq(1).Text())
				result := strings.TrimSpace(cells.Eq(2).Text())

				// If we have a table where the result matches our target element
				isTargetResult := strings.EqualFold(cleanElementName(result), cleanElementName(targetElement))

				// Or if we're on the element's page and this looks like a recipe row
				onElementPage := strings.EqualFold(cleanElementName(pageTitle), cleanElementName(targetElement))

				if (isTargetResult || onElementPage) && a != "" && b != "" {
					a = cleanElementName(a)
					b = cleanElementName(b)

					key := strings.ToLower(a + "|" + b)
					if !seen[key] {
						recipes = append(recipes, []string{a, b})
						seen[key] = true
					}
				}
			}
		})
	})

	// ==================== APPROACH 3: Visual recipe format ====================
	// This approach is generalized to work with any element, not just Brick
	doc.Find("div.mw-parser-output").Each(func(_ int, div *goquery.Selection) {
		div.Find("p, div").Each(func(_ int, elem *goquery.Selection) {
			elemHTML, _ := elem.Html()

			// If we find a pattern that looks like a recipe block (contains + sign)
			if strings.Contains(elemHTML, "+") {
				elemText := elem.Text()
				lines := strings.Split(elemText, "\n")

				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(line, "+") {
						parts := strings.Split(line, "+")
						if len(parts) == 2 {
							a := strings.TrimSpace(parts[0])
							b := strings.TrimSpace(parts[1])

							a = cleanElementName(a)
							b = cleanElementName(b)

							if a != "" && b != "" {
								key := strings.ToLower(a + "|" + b)
								if !seen[key] {
									recipes = append(recipes, []string{a, b})
									seen[key] = true
								}
							}
						}
					}
				}
			}
		})
	})

	// ==================== APPROACH 4: Look for specific list class ====================
	// This is a common format on the wiki for recipe lists
	doc.Find("ul.wds-list, ul.wikia-gallery").Each(func(_ int, ul *goquery.Selection) {
		ul.Find("li").Each(func(_ int, li *goquery.Selection) {
			elemText := li.Text()
			if strings.Contains(elemText, "+") {
				parts := strings.Split(elemText, "+")
				if len(parts) == 2 {
					a := strings.TrimSpace(parts[0])
					b := strings.TrimSpace(parts[1])

					a = cleanElementName(a)
					b = cleanElementName(b)

					if a != "" && b != "" {
						key := strings.ToLower(a + "|" + b)
						if !seen[key] {
							recipes = append(recipes, []string{a, b})
							seen[key] = true
						}
					}
				}
			}
		})
	})

	// ==================== APPROACH 5: Look for list items with images ====================
	// Many recipe displays use images for the elements
	doc.Find("li").Each(func(_ int, li *goquery.Selection) {
		// Check if this list item might be a recipe (contains images and + sign)
		if li.Find("img").Length() > 0 && strings.Contains(li.Text(), "+") {
			elemText := li.Text()
			parts := strings.Split(elemText, "+")
			if len(parts) == 2 {
				a := strings.TrimSpace(parts[0])
				b := strings.TrimSpace(parts[1])

				a = cleanElementName(a)
				b = cleanElementName(b)

				if a != "" && b != "" {
					key := strings.ToLower(a + "|" + b)
					if !seen[key] {
						recipes = append(recipes, []string{a, b})
						seen[key] = true
					}
				}
			}
		}
	})

	// ==================== APPROACH 6: Additional recipe list classes ====================
	// Looking for more specific classes used in recipe lists
	doc.Find("div.recipe-list, div.recipes, section.recipes").Each(func(_ int, div *goquery.Selection) {
		div.Find("li, div").Each(func(_ int, item *goquery.Selection) {
			elemText := item.Text()
			if strings.Contains(elemText, "+") {
				parts := strings.Split(elemText, "+")
				if len(parts) == 2 {
					a := strings.TrimSpace(parts[0])
					b := strings.TrimSpace(parts[1])

					a = cleanElementName(a)
					b = cleanElementName(b)

					if a != "" && b != "" {
						key := strings.ToLower(a + "|" + b)
						if !seen[key] {
							recipes = append(recipes, []string{a, b})
							seen[key] = true
						}
					}
				}
			}
		})
	})

	// ==================== APPROACH 7: Fallback if no recipes found ====================
	// More aggressive search if we haven't found any recipes yet
	if len(recipes) == 0 {
		// Look for any element that might contain a recipe pattern
		doc.Find("*").Each(func(_ int, elem *goquery.Selection) {
			elemText := elem.Text()
			if strings.Contains(elemText, "+") {
				lines := strings.Split(elemText, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(line, "+") {
						parts := strings.Split(line, "+")
						if len(parts) == 2 {
							a := strings.TrimSpace(parts[0])
							b := strings.TrimSpace(parts[1])

							a = cleanElementName(a)
							b = cleanElementName(b)

							if a != "" && b != "" {
								key := strings.ToLower(a + "|" + b)
								if !seen[key] {
									recipes = append(recipes, []string{a, b})
									seen[key] = true
								}
							}
						}
					}
				}
			}
		})
	}

	// Debug output
	fmt.Printf("Found %d recipes for %s\n", len(recipes), targetElement)
	for i, recipe := range recipes {
		if i < 10 { // Limit debug output
			fmt.Printf("  Recipe %d: %s + %s\n", i+1, recipe[0], recipe[1])
		}
	}

	// Sort recipes for consistent output
	sort.Slice(recipes, func(i, j int) bool {
		if recipes[i][0] == recipes[j][0] {
			return recipes[i][1] < recipes[j][1]
		}
		return recipes[i][0] < recipes[j][0]
	})

	// If this is likely a main element page and we have no recipes,
	// add a warning to help with debugging
	if len(recipes) == 0 && strings.EqualFold(cleanElementName(pageTitle), cleanElementName(targetElement)) {
		fmt.Printf("WARNING: No recipes found for %s on its own page!\n", targetElement)
	}

	return recipes, nil
}

func fallbackScrape(doc *goquery.Document) [][]string {
	recipes := make([][]string, 0)
	seen := make(map[string]bool)

	// Tabel umum
	doc.Find("table.wikitable, table.article-table").Each(func(_ int, table *goquery.Selection) {
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			cols := row.Find("td")
			if cols.Length() >= 3 {
				a := strings.TrimSpace(cols.Eq(0).Text())
				b := strings.TrimSpace(cols.Eq(1).Text())
				if a != "" && b != "" {
					key := strings.ToLower(a + "|" + b)
					if !seen[key] {
						recipes = append(recipes, []string{a, b})
						seen[key] = true
					}
				}
			}
		})
	})

	// List kombinasi
	doc.Find("div.mw-parser-output ul li").Each(func(_ int, li *goquery.Selection) {
		text := li.Text()
		if strings.Contains(text, "→") {
			text = strings.Split(text, "→")[0]
		}
		if strings.Contains(text, "=") {
			text = strings.Split(text, "=")[0]
		}
		if strings.Contains(text, "+") {
			parts := strings.Split(text, "+")
			if len(parts) == 2 {
				a := strings.TrimSpace(parts[0])
				b := strings.TrimSpace(parts[1])
				key := strings.ToLower(a + "|" + b)
				if a != "" && b != "" && !seen[key] {
					recipes = append(recipes, []string{a, b})
					seen[key] = true
				}
			}
		}
	})

	return recipes
}

// ================ PERBAIKAN UTAMA ================
// Algoritma baru untuk menghitung tier
func calcTiersFix(elements []Element) {
	fmt.Println("\n=== Menggunakan algoritma tier yang dioptimalkan ===")

	// Step 1: Buat beberapa mapping yang diperlukan
	elementMap := make(map[string]*Element)       // nama elemen -> pointer ke Element
	elementRecipes := make(map[string][][]string) // nama elemen -> resep-resepnya

	// Konversi semua nama elemen ke lowercase untuk konsistensi
	for i := range elements {
		lowerName := strings.ToLower(elements[i].Name)
		elementMap[lowerName] = &elements[i]

		// Buat salinan resep dengan nama yang dinormalisasi
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

	// Step 2: Set tier 0 untuk elemen dasar
	for _, basic := range basicElements {
		if elem, exists := elementMap[basic]; exists {
			elem.Tier = 0
		} else {
			fmt.Printf("ERROR: Basic element %s missing from map\n", basic)
		}
	}

	// Step 3: Buat dependency graph untuk menentukan urutan penghitungan
	dependencies := make(map[string]map[string]bool) // elemen -> {dependencies}
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

	// Step 4: Iterasi untuk menghitung tier - menggunakan pendekatan dynamic programming
	// maxIterations untuk menghindari infinite loop jika ada cyclic dependency
	changed := true
	iteration := 0
	maxIterations := 100 // Jumlah maksimum iterasi untuk menghindari infinite loop

	for changed && iteration < maxIterations {
		changed = false
		iteration++

		for elemName, elem := range elementMap {
			// Skip elemen yang sudah memiliki tier atau tidak memiliki resep
			if elem.Tier != -1 || len(elementRecipes[elemName]) == 0 {
				continue
			}

			// Untuk setiap resep, cek apakah semua dependencies sudah memiliki tier
			for _, recipe := range elementRecipes[elemName] {
				if len(recipe) != 2 {
					continue
				}

				// Cari tier maksimum dari bahan resep
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

				// Jika semua ingredients sudah punya tier, hitung tier elemen ini
				if allIngredientsHaveTier {
					elem.Tier = maxIngredientTier + 1
					changed = true
					break // Lanjut ke elemen berikutnya
				}
			}
		}

		fmt.Printf("Iteration %d: Updated %d elements\n", iteration, countUpdatedElements(elements))
	}

	// Step 5: Set tier untuk elemen tanpa resep (kecuali elemen dasar)
	// Elemen tanpa resep bisa jadi merupakan elemen dasar tambahan atau elemen final
	missingRecipesCount := 0
	for _, elem := range elements {
		if elem.Tier == -1 && len(elem.Recipes) == 0 && !isBasicElement(elem.Name) {
			// Elemen tanpa resep yang bukan elemen dasar, set tier ke 999 sebagai penanda
			elem.Tier = 999
			missingRecipesCount++
		}
	}

	// Step 6: Set tier elemen lain yang masih -1 ke 998 untuk menandai bahwa ada masalah
	unresolvableCount := 0
	for _, elem := range elements {
		if elem.Tier == -1 {
			elem.Tier = 998 // Tidak bisa menentukan tier
			unresolvableCount++
		}
	}

	// Report
	fmt.Printf("\nTier calculation completed in %d iterations\n", iteration)
	fmt.Printf("Elements with missing recipes: %d\n", missingRecipesCount)
	fmt.Printf("Elements with unresolvable tiers: %d\n", unresolvableCount)
}

// Helper untuk menghitung jumlah elemen yang sudah diupdate (tier != -1)
func countUpdatedElements(elements []Element) int {
	count := 0
	for _, e := range elements {
		if e.Tier != -1 {
			count++
		}
	}
	return count
}

// Fungsi untuk menganalisis distribusi tier
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

	// Pertama cetak tier normal (0-20)
	totalNormalTiers := 0
	for t := 0; t <= 20; t++ {
		if count, ok := tierCounts[t]; ok && count > 0 {
			fmt.Printf("  Tier %d: %d elements\n", t, count)
			totalNormalTiers += count
		}
	}

	// Kemudian cetak tier khusus
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

	// Sample tiap kategori
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

	// Khusus untuk elemen tier 1, tampilkan semua
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

// Fungsi untuk mengecek apakah slice string mengandung string tertentu
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
