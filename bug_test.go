package main

import (
	"fmt"
)

// demonstrateBug demonstrates the equipment parsing bug
func demonstrateBug() {
	properties := []string{"strength+2", "strength-2", "dexterity+10", "dexterity-10"}

	for _, property := range properties {
		fmt.Printf("Testing property: %s\n", property)

		var stat string
		var modifier int
		var sign int

		if len(property) > 1 {
			// Current buggy logic
			if property[len(property)-2] == '+' {
				stat = property[:len(property)-2]
				sign = 1
				fmt.Sscanf(property[len(property)-1:], "%d", &modifier)
			} else if property[len(property)-2] == '-' {
				stat = property[:len(property)-2]
				sign = -1
				fmt.Sscanf(property[len(property)-1:], "%d", &modifier)
			}
		}

		fmt.Printf("  Result: stat=%s, modifier=%d, sign=%d, total=%d\n", stat, modifier, sign, sign*modifier)
		fmt.Println()
	}
}

func main() {
	demonstrateBug()
}
