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

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

var basicElements = []string{"air", "earth", "fire", "water"}

func main() {
	baseURL := "https://little-alchemy.fandom.com"
	listURL := baseURL + "/wiki/Elements_(Little_Alchemy_2)"

	fmt.Println("Starting Little Alchemy 2 recipe scraper...")

	elementsList, err := getElementsList(listURL)
	if err != nil {
		log.Fatalf("Failed to get elements list: %v", err)
	}

	// Add basic elements if not already present
	for _, basic := range basicElements {
		found := false
		for _, elem := range elementsList {
			if strings.EqualFold(elem, basic) {
				found = true
				break
			}
		}
		if !found {
			elementsList = append(elementsList, basic)
		}
	}

	fmt.Printf("Found %d elements to scrape\n", len(elementsList))
	if len(elementsList) > 0 {
		fmt.Println("Sample elements:", elementsList[:min(5, len(elementsList))])
	}

	var elements []Element
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 3) // Limit concurrent requests
	var mutex sync.Mutex

	for _, elemName := range elementsList {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			url := baseURL + "/wiki/" + strings.ReplaceAll(name, " ", "_")
			fmt.Printf("Scraping: %s\n", name)

			recipes, err := scrapeRecipes(url)
			if err != nil {
				log.Printf("Error scraping %s: %v", name, err)
				return
			}

			mutex.Lock()
			elements = append(elements, Element{
				Name:    name,
				Recipes: recipes,
			})
			mutex.Unlock()

			time.Sleep(300 * time.Millisecond) // Be nice to the server
		}(elemName)
	}

	wg.Wait()

	outputFile := "elements.json"
	err = saveStructuredJSON(elements, outputFile)
	if err != nil {
		log.Fatalf("Failed to save %s: %v", outputFile, err)
	}

	fmt.Printf("Scraping complete! Data saved to %s\n", outputFile)
}

func getElementsList(url string) ([]string, error) {
	elements := []string{}
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Try multiple selectors to find elements
	selectors := []string{
		"div#mw-content-text > div > ul > li > a",
		"div.mw-parser-output > ul > li > a",
		"div#mw-content-text table td a",
		"table.article-table a",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			name := strings.TrimSpace(s.Text())
			if name != "" && !strings.Contains(name, "Category:") {
				elements = append(elements, name)
			}
		})
		if len(elements) > 0 {
			break
		}
	}

	return elements, nil
}

func scrapeRecipes(url string) ([][]string, error) {
	recipes := [][]string{}
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	elementName := getElementNameFromURL(url)
	fmt.Printf("Looking for recipes for: %s\n", elementName)

	// Check for recipe tables
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		isRecipeTable := false
		table.Find("th").Each(func(j int, th *goquery.Selection) {
			if strings.Contains(strings.ToLower(th.Text()), "recipe") ||
				strings.Contains(strings.ToLower(th.Text()), "ingredient") ||
				strings.Contains(strings.ToLower(th.Text()), "combination") {
				isRecipeTable = true
			}
		})

		if isRecipeTable {
			table.Find("tr").Each(func(j int, tr *goquery.Selection) {
				var ingredients []string
				tr.Find("td").Each(func(k int, td *goquery.Selection) {
					text := strings.TrimSpace(td.Text())
					if text != "" && len(text) < 30 {
						ingredients = append(ingredients, text)
					}
				})
				if len(ingredients) >= 2 {
					recipes = append(recipes, []string{ingredients[0], ingredients[1]})
				}
			})
		}
	})

	// Check for recipe sections
	recipeSection := false
	doc.Find("h2, h3, p, li").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		lower := strings.ToLower(text)

		// Detect recipe sections
		if strings.Contains(lower, "recipe") ||
			strings.Contains(lower, "how to make") ||
			strings.Contains(lower, "combination") {
			recipeSection = true
		}

		if recipeSection {
			// Format 1: "X + Y"
			if strings.Contains(text, "+") {
				parts := strings.Split(text, "+")
				if len(parts) == 2 {
					ing1 := strings.TrimSpace(parts[0])
					ing2 := strings.TrimSpace(parts[1])
					if ing1 != "" && ing2 != "" && len(ing1) < 30 && len(ing2) < 30 {
						recipes = append(recipes, []string{ing1, ing2})
					}
				}
			}

			// Format 2: "Combine X and Y"
			if strings.Contains(lower, "combine") && strings.Contains(lower, "and") {
				parts := strings.Split(strings.TrimPrefix(lower, "combine"), "and")
				if len(parts) == 2 {
					ing1 := strings.TrimSpace(parts[0])
					ing2 := strings.TrimSpace(parts[1])
					if ing1 != "" && ing2 != "" && len(ing1) < 30 && len(ing2) < 30 {
						recipes = append(recipes, []string{ing1, ing2})
					}
				}
			}
		}

		// Reset section flag if we hit another header
		if s.Is("h2") || s.Is("h3") {
			if !strings.Contains(lower, "recipe") &&
				!strings.Contains(lower, "how to make") &&
				!strings.Contains(lower, "combination") {
				recipeSection = false
			}
		}
	})

	fmt.Printf("Found %d recipes for %s\n", len(recipes), elementName)
	return recipes, nil
}

func getElementNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return strings.ReplaceAll(parts[len(parts)-1], "_", " ")
	}
	return ""
}

func saveStructuredJSON(elements []Element, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(elements)
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
