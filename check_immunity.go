package main

import (
	"fmt"
	"goldbox-rpg/pkg/game"
)

func main() {
	fmt.Println("ImmunityNone:", int(game.ImmunityNone))
	fmt.Println("ImmunityPartial:", int(game.ImmunityPartial))
	fmt.Println("ImmunityComplete:", int(game.ImmunityComplete))
	fmt.Println("ImmunityReflect:", int(game.ImmunityReflect))
}
