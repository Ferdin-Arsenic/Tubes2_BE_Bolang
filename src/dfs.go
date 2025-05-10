package main

import "strings"

type DFSData struct {
	elementMap      map[string]Element
    initialTarget   string
    maxRecipes      int
    allFoundRecipes *[]map[string][]string
}

func dfsMultiple(elementMap map[string]Element, target string, maxRecipes int) []map[string][]string {

	allFoundRecipes := make([]map[string][]string, 0)

	 if isBasicElement(target) {
		emptyRecipe := make(map[string][]string)
		allFoundRecipes = append(allFoundRecipes, emptyRecipe)
        return allFoundRecipes
    }	

	// Map untuk membangun langkah-langkah resep untuk satu resep
	currentRecipeSteps := make(map[string][]string)

	dfsData := DFSData{
		elementMap:      elementMap,
		initialTarget:   target,
		maxRecipes:      maxRecipes,
		allFoundRecipes: &allFoundRecipes,
	}

	// Param target pertama untuk rekursi pertama, target kedua untuk setiap rekursi
	dfsData.dfsRecursive(target, currentRecipeSteps)

	return allFoundRecipes
}

// Fungsi pembantu rekursif untuk DFS backward
func (d *DFSData) dfsRecursive(elementToMakeCurrently string, currentRecipeSteps map[string][]string) bool {
	if len(*d.allFoundRecipes) >= d.maxRecipes {
		return false 
	}

	if isBasicElement(elementToMakeCurrently) {
		return true
	}

	elemDetails, exists := d.elementMap[elementToMakeCurrently]
	if !exists || len(elemDetails.Recipes) == 0 {
		return false // tidak ada resep
	}

	productTier := elemDetails.Tier
	var madeSuccessfully bool = false // Berhasil dibuat dgn setidaknya satu cara

	for _, recipePair := range elemDetails.Recipes {
		if d.maxRecipes > 0 && len(*d.allFoundRecipes) >= d.maxRecipes && elementToMakeCurrently != d.initialTarget {
			break
		}


		if len(recipePair) != 2 {
			continue
		}
		parent1Name := strings.ToLower(recipePair[0])
		parent2Name := strings.ToLower(recipePair[1])

		elemParent1, p1Exists := d.elementMap[parent1Name]
		elemParent2, p2Exists := d.elementMap[parent2Name]

		if !p1Exists || !p2Exists {
			continue
		}
		if elemParent1.Tier >= productTier || elemParent2.Tier >= productTier {
			continue
		}

		currentRecipeSteps[elementToMakeCurrently] = []string{parent1Name, parent2Name}

		parent1OK := d.dfsRecursive(parent1Name, currentRecipeSteps)
		if !parent1OK {
			delete(currentRecipeSteps, elementToMakeCurrently) // Backtrack
			continue                                           // Coba resep lain
		}
        if len(*d.allFoundRecipes) >= d.maxRecipes && elementToMakeCurrently != d.initialTarget {
             delete(currentRecipeSteps, elementToMakeCurrently)
             continue
        }


		parent2OK := d.dfsRecursive(parent2Name, currentRecipeSteps)
		if !parent2OK {
			delete(currentRecipeSteps, elementToMakeCurrently) // Backtrack
			continue                                           // Coba resep lain
		}

		// Jika sampai sini, kedua parent OK, jadi element berhasil dibuat dengan resep ini
		madeSuccessfully = true

		if elementToMakeCurrently == d.initialTarget {
			if d.maxRecipes <= 0 || len(*d.allFoundRecipes) < d.maxRecipes {
				finalRecipe := make(map[string][]string)
				for key, val := range currentRecipeSteps {
					parentsCopy := make([]string, len(val))
					copy(parentsCopy, val)
					finalRecipe[key] = parentsCopy
				}
				*d.allFoundRecipes = append(*d.allFoundRecipes, finalRecipe)
			}
		}
		
        if elementToMakeCurrently == d.initialTarget && len(*d.allFoundRecipes) >= d.maxRecipes {
            break
        }
	}
	return madeSuccessfully
}