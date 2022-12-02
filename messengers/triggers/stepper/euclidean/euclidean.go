package euclidean

import "math"

/*
func bresenhamEuclidean(onsets: Int, pulses: Int) -> [Int] {
    let slope = Double(onsets) / Double(pulses)
    var result = [Int]()
    var previous: Int? = nil
    for i in 0..<pulses {
        let current = Int(floor(Double(i) * slope))
        result.append(current != previous ? 1 : 0)
        previous = current
    }
    return result
}
*/

type Euclidean struct {
	pattern       []bool
	steps         int
	events        int
	rotation      int
	BarDurationMS float64
}

func (euclid *Euclidean) recalculatePattern() {
	slope := float64(euclid.events) / float64(euclid.steps)
	pattern := make([]bool, euclid.steps)
	previous := -1

	for i := 0; i < euclid.steps; i++ {
		current := int(math.Floor(float64(i) * slope))
		pattern[i] = current != previous
		previous = current
	}
}
