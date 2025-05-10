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
	
	// Map untuk mendeteksi cycle
	pathCycleDetector := make(map[string]bool)

	// Map untuk membangun langkah-langkah resep untuk satu resep
	currentRecipeSteps := make(map[string][]string)

	dfsData := DFSData{
		elementMap:      elementMap,
		initialTarget:   target,
		maxRecipes:      maxRecipes,
		allFoundRecipes: &allFoundRecipes,
	}

	// Param target pertama untuk rekursi pertama, target kedua untuk setiap rekursi
	dfsData.dfsRecursive(target, currentRecipeSteps, pathCycleDetector)

	return allFoundRecipes
}

// Fungsi pembantu rekursif untuk DFS backward
func (d *DFSData) dfsRecursive(elementToMakeCurrently string, currentRecipeSteps map[string][]string, pathCycleDetector map[string]bool,) {
	if len(*d.allFoundRecipes) >= d.maxRecipes {
		return
	}

	// Base case: Jika elemen saat ini adalah elemen dasar, cabang ini berhasil.
	if isBasicElement(elementToMakeCurrently) {
		return
	}

	if pathCycleDetector[elementToMakeCurrently] {
		return // Siklus terdeteksi, jalur ini tidak valid
	}
	pathCycleDetector[elementToMakeCurrently] = true
	defer delete(pathCycleDetector, elementToMakeCurrently)

	elemDetails, exists := d.elementMap[elementToMakeCurrently]
	if !exists || len(elemDetails.Recipes) == 0 {
		// Tidak ada resep untuk , jalur ini salah.
		return
	}

	productTier := elemDetails.Tier

	// Coba setiap resep yang tersedia untuk membuat elementToMakeCurrently
	for _, recipePair := range elemDetails.Recipes { // recipePair adalah [P1, P2]
		if d.maxRecipes > 0 && len(*d.allFoundRecipes) >= d.maxRecipes {
			return
		}

		if len(recipePair) != 2 {
			continue
		}
		parent1Name := strings.ToLower(recipePair[0])
		parent2Name := strings.ToLower(recipePair[1])

		// Validasi Tier
		elemParent1, p1Exists := d.elementMap[parent1Name]
		elemParent2, p2Exists := d.elementMap[parent2Name]

		if !p1Exists || !p2Exists { // Ngecek bahan ada atau tidak
			continue
		}
		if elemParent1.Tier > productTier || elemParent2.Tier > productTier {
			continue
		}

		// Resep Valid, catat
		currentRecipeSteps[elementToMakeCurrently] = []string{parent1Name, parent2Name}

		d.dfsRecursive(parent1Name, currentRecipeSteps, pathCycleDetector)
		
		// Cek apakah parent1 berhasil dibuat (jika tidak dasar, harus ada di currentRecipeSteps)
		if !isBasicElement(parent1Name) && currentRecipeSteps[parent1Name] == nil {
			delete(currentRecipeSteps, elementToMakeCurrently)
			continue
		}
        if len(*d.allFoundRecipes) >= d.maxRecipes {
            continue
        }


		// Rekursi untuk parent2
		d.dfsRecursive(parent2Name, currentRecipeSteps, pathCycleDetector)

		if !isBasicElement(parent2Name) && currentRecipeSteps[parent2Name] == nil {
			delete(currentRecipeSteps, elementToMakeCurrently)
			continue
		}
        if len(*d.allFoundRecipes) >= d.maxRecipes {
            continue
        }

		if elementToMakeCurrently == d.initialTarget {
			if len(*d.allFoundRecipes) >= d.maxRecipes {
            	continue
        	}
			// Copy currentRecipeSteps agar tidak termodifikasi
			finalRecipe := make(map[string][]string)
			for key, val := range currentRecipeSteps {
				parentsCopy := make([]string, len(val))
				copy(parentsCopy, val)
				finalRecipe[key] = parentsCopy
			}
			*d.allFoundRecipes = append(*d.allFoundRecipes, finalRecipe)
		}
		
		delete(currentRecipeSteps, elementToMakeCurrently)

	}
}