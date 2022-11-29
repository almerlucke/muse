package sampler

type ZeroCrossingDirection int

const (
	ZeroCrossingUp ZeroCrossingDirection = iota
	ZeroCrossingDown
	ZeroCrossingAny
)

type ZeroCrossing struct {
	Position  int
	Direction ZeroCrossingDirection
}

func (z *ZeroCrossing) MatchesDirection(direction ZeroCrossingDirection) bool {
	if direction == ZeroCrossingAny {
		return true
	}

	return z.Direction == direction
}

// Single channel buffer information
type BufferInfo struct {
	Buffer        []float64
	ZeroCrossings []*ZeroCrossing
}

func NewBufferInfo(buffer []float64) *BufferInfo {
	zeroCrossings := []*ZeroCrossing{}

	prevSample := 0.0

	for index, sample := range buffer {
		if prevSample <= 0.0 && sample > 0.0 {
			zeroCrossings = append(zeroCrossings, &ZeroCrossing{Position: index, Direction: ZeroCrossingUp})
		} else if prevSample >= 0.0 && sample < 0.0 {
			zeroCrossings = append(zeroCrossings, &ZeroCrossing{Position: index, Direction: ZeroCrossingDown})
		}

		prevSample = sample
	}

	return &BufferInfo{Buffer: buffer, ZeroCrossings: zeroCrossings}
}

func (info *BufferInfo) ClosestCrossingOfType(location int, direction ZeroCrossingDirection) *ZeroCrossing {
	var closest *ZeroCrossing
	distance := 0

	for _, zero := range info.ZeroCrossings {
		if zero.MatchesDirection(direction) {
			if closest == nil {
				closest = zero
				distance = zero.Position - location
				if distance < 0 {
					distance = -distance
				}
			} else {
				dist := zero.Position - location
				if dist < 0 {
					dist = -dist
				}
				if dist < distance {
					closest = zero
					distance = dist
				}
			}
		}
	}

	return closest
}
