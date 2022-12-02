package main

import "log"

type Euclidean struct {
	pattern       []bool
	steps         int
	events        int
	rotation      int
	BarDurationMS float64
}

func (euclid *Euclidean) recalculate() {
	//Each iteration is a process of pairing strings X and Y and the remainder from the pairings
	//X will hold the "dominant" pair (the pair that there are more of)
	x := "1"
	x_amount := euclid.events

	y := "0"
	y_amount := euclid.steps - euclid.events

	for true {
		x_temp := x_amount
		y_temp := y_amount
		y_copy := y

		//Check which is the dominant pair
		if x_temp >= y_temp {
			//Set the new number of pairs for X and Y
			x_amount = y_temp
			y_amount = x_temp - y_temp

			//The previous dominant pair becomes the new non dominant pair
			y = x
		} else {
			x_amount = x_temp
			y_amount = y_temp - x_temp
		}

		//Create the new dominant pair by combining the previous pairs
		x += y_copy

		if x_amount > 1 && y_amount > 1 {
			continue
		} else {
			break
		}
	}

	//By this point, we have strings X and Y formed through a series of pairings of the initial strings "1" and "0"
	//X is the final dominant pair and Y is the second to last dominant pair

	rhythm := ""

	for i := 1; i <= x_amount; i++ {
		rhythm += x
	}

	for i := 1; i <= y_amount; i++ {
		rhythm += y
	}

	pattern := make([]bool, euclid.steps)

	for i, c := range rhythm {
		pattern[i] = c == '1'
	}

	euclid.pattern = pattern
}

func main() {
	euclid := &Euclidean{steps: 8, events: 5, rotation: 0}

	euclid.recalculate()

	for _, tick := range euclid.pattern {
		if tick {
			log.Printf("1")
		} else {
			log.Printf("0")
		}
	}
}
