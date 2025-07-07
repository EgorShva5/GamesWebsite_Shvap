package main

import (
	"fmt"
	"slices"
)

func main() {
	a := []int{1, 2, 3}
	a = slices.Insert(a, 0, 20)
	fmt.Println(a)
}
