package main

import "fmt"

func calculateLevel(exp int) int {
	levels := []int{0, 2000, 4000, 8000, 16000, 32000, 64000}
	for level, requirement := range levels {
		if exp < requirement {
			return level
		}
	}
	return len(levels)
}

func main() {
	fmt.Printf("calculateLevel(0) = %d\n", calculateLevel(0))
	fmt.Printf("calculateLevel(1000) = %d\n", calculateLevel(1000))
	fmt.Printf("calculateLevel(1999) = %d\n", calculateLevel(1999))
	fmt.Printf("calculateLevel(2000) = %d\n", calculateLevel(2000))
	fmt.Printf("calculateLevel(4000) = %d\n", calculateLevel(4000))
}
