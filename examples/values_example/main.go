package main

import (
	"log"
	"math/rand"
)

// "github.com/almerlucke/muse/value"

type Mode int

type MirrorMode int

const (
	Up Mode = iota
	Converge
	Alternate
	Random
)

const (
	None MirrorMode = iota
	Exclusive
	Inclusive
)

func main() {
	mode := Alternate
	mirrorMode := Exclusive
	reverse := false
	sequence := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}

	n := len(sequence)
	pattern := []int{}

	switch mode {
	case Random:
		for i := 0; i < n; i++ {
			pattern = append(pattern, i)
		}
		rand.Shuffle(len(pattern), func(i, j int) { pattern[i], pattern[j] = pattern[j], pattern[i] })
	case Up:
		for i := 0; i < n; i++ {
			pattern = append(pattern, i)
		}
	case Converge:
		for i := 0; i < n/2; i++ {
			pattern = append(pattern, i, n-1-i)
		}
		if n%2 == 1 {
			pattern = append(pattern, n/2)
		}
	case Alternate:
		for i := 1; i < n; i++ {
			pattern = append(pattern, 0, i)
		}
	}

	n = len(pattern)

	if mirrorMode == Exclusive {
		for i := n - 2; i >= 1; i-- {
			pattern = append(pattern, pattern[i])
		}
	} else if mirrorMode == Inclusive {
		for i := n - 1; i >= 0; i-- {
			pattern = append(pattern, pattern[i])
		}
	}

	n = len(pattern)

	if reverse {
		for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
			pattern[i], pattern[j] = pattern[j], pattern[i]
		}
	}

	log.Printf("pattern: %v", pattern)
}
