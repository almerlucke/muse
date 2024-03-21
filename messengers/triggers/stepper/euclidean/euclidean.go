package euclidean

type Euclidean struct {
	rhythm         string
	currentStep    int
	numSteps       int
	numEvents      int
	rotation       int
	stepDurationMS float64
}

func NewEuclidian(numSteps int, numEvents int, rotation int, stepDurationMS float64) *Euclidean {
	rotation = rotation % numSteps
	if numEvents > numSteps {
		numEvents = numSteps
	}

	euclid := &Euclidean{
		currentStep:    rotation,
		numSteps:       numSteps,
		numEvents:      numEvents,
		rotation:       rotation,
		stepDurationMS: stepDurationMS,
	}

	euclid.recalculate()

	return euclid
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

func (euclid *Euclidean) NextStep() float64 {
	step := euclid.rhythm[euclid.currentStep] == '1'
	euclid.currentStep = (euclid.currentStep + 1) % euclid.numSteps

	if step {
		return euclid.stepDurationMS
	}

	return -euclid.stepDurationMS
}

func (euclid *Euclidean) SetState(state map[string]any) {

}

func (euclid *Euclidean) GetState() map[string]any {
	return map[string]any{}
}

// muse.Stater
