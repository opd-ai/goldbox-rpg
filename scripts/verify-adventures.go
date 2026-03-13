//go:build ignore

// Adventure verification script for GoldBox RPG Engine
// Validates all adventure YAML files and reports status

package main

import (
	"fmt"
	"os"
	"sort"

	"goldbox-rpg/pkg/game"
)

func main() {
	mgr := game.NewAdventureManager("data/adventures")
	if err := mgr.LoadAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading adventures: %v\n", err)
		os.Exit(1)
	}

	summaries := mgr.List()
	count := len(summaries)

	if count == 0 {
		fmt.Println("⚠️  No adventures found in data/adventures/")
		os.Exit(1)
	}

	// Sort by level range for display
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].MinLevel < summaries[j].MinLevel
	})

	fmt.Printf("✅ Adventure Verification Report\n")
	fmt.Printf("================================\n")
	fmt.Printf("Total Adventures: %d/10\n\n", count)

	totalHours := 0
	totalMaps := 0
	totalQuests := 0

	for i, s := range summaries {
		fmt.Printf("%2d. %s (%s)\n", i+1, s.Title, s.Slug)
		fmt.Printf("    Levels: %d-%d | Est. Hours: %s | Maps: %d | Quests: %d\n",
			s.MinLevel, s.MaxLevel, s.EstHours, s.MapCount, s.QuestCount)
		totalMaps += s.MapCount
		totalQuests += s.QuestCount
		// Parse hours for rough total (take first number)
		var h int
		fmt.Sscanf(s.EstHours, "%d", &h)
		totalHours += h
	}

	fmt.Printf("\n================================\n")
	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Maps: %d\n", totalMaps)
	fmt.Printf("  Total Quests: %d\n", totalQuests)
	fmt.Printf("  Est. Total Playtime: %d+ hours\n", totalHours)

	if count >= 10 {
		fmt.Printf("\n✅ All 10 adventures validated successfully!\n")
	} else {
		fmt.Printf("\n⚠️  Only %d/10 adventures present\n", count)
	}
}
