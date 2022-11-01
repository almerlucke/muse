package main

import "log"

func main() {
	a := []int{1, 2, 3, 4}
	removeIndex := 0

	a = append(a[:removeIndex], a[removeIndex+1:]...)

	log.Printf("a: %v", a)
}
