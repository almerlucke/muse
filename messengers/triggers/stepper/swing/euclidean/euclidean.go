package euclidean

import "github.com/almerlucke/muse/messengers/triggers/stepper/swing"

type StepConfig struct {
	Shuffle     float64
	ShuffleRand float64
	SkipChance  float64
	Multiply    float64
	BurstChance float64
	NumBurst    int
}

type Euclidean struct {
	rhythm    string
	stepIndex int
	stepCnt   int
	numSteps  int
	numEvents int
	rotation  int
	config    StepConfig
}

func New(numSteps int, numEvents int, rotation int, config *StepConfig) *Euclidean {
	rotation = rotation % numSteps
	if numEvents > numSteps {
		numEvents = numSteps
	}

	euclid := &Euclidean{
		stepIndex: rotation,
		numSteps:  numSteps,
		numEvents: numEvents,
		rotation:  rotation,
	}

	if config != nil {
		euclid.config = *config
	}

	if euclid.config.Multiply <= 0.0 {
		euclid.config.Multiply = 1.0
	}

	euclid.recalculate()

	return euclid
}

func (euclid *Euclidean) Set(numSteps int, numEvents int, rotation int) {
	rotation = rotation % numSteps
	if numEvents > numSteps {
		numEvents = numSteps
	}

	euclid.numSteps = numSteps
	euclid.numEvents = numEvents
	euclid.rotation = rotation

	euclid.recalculate()
	euclid.Reset()
}

func (euclid *Euclidean) recalculate() {
	//Each iteration is a process of pairing strings X and Y and the remainder from the pairings
	//X will hold the "dominant" pair (the pair that there are more of)
	x := "1"
	xAmount := euclid.numEvents

	y := "0"
	yAmount := euclid.numSteps - euclid.numEvents

	for {
		xTemp := xAmount
		yTemp := yAmount
		yCopy := y

		//Check which is the dominant pair
		if xTemp >= yTemp {
			//Set the new number of pairs for X and Y
			xAmount = yTemp
			yAmount = xTemp - yTemp

			//The previous dominant pair becomes the new non dominant pair
			y = x
		} else {
			xAmount = xTemp
			yAmount = yTemp - xTemp
		}

		//Create the new dominant pair by combining the previous pairs
		x += yCopy

		if xAmount <= 1 || yAmount <= 1 {
			break
		}
	}

	//By this point, we have strings X and Y formed through a series of pairings of the initial strings "1" and "0"
	//X is the final dominant pair and Y is the second to last dominant pair

	rhythm := ""

	for i := 1; i <= xAmount; i++ {
		rhythm += x
	}

	for i := 1; i <= yAmount; i++ {
		rhythm += y
	}

	euclid.rhythm = rhythm
}

func (euclid *Euclidean) Generate() *swing.Step {
	step := euclid.rhythm[euclid.stepIndex] == '1'
	euclid.stepIndex = (euclid.stepIndex + 1) % euclid.numSteps
	euclid.stepCnt++

	if step {
		return &swing.Step{
			Shuffle:     euclid.config.Shuffle,
			ShuffleRand: euclid.config.ShuffleRand,
			SkipChance:  euclid.config.SkipChance,
			Multiply:    euclid.config.Multiply,
			BurstChance: euclid.config.BurstChance,
			NumBurst:    euclid.config.NumBurst,
		}
	}

	return &swing.Step{
		Skip: true,
	}
}

func (euclid *Euclidean) Done() bool {
	return euclid.stepCnt >= euclid.numSteps
}

func (euclid *Euclidean) Reset() {
	euclid.stepCnt = 0
	euclid.stepIndex = euclid.rotation
}

func (euclid *Euclidean) Continuous() bool {
	return false
}
