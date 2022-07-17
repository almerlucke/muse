package sliceprovider

type SliceProvider struct {
	steps     []float64
	stepIndex int
}

func New(steps []float64) *SliceProvider {
	return &SliceProvider{
		steps: steps,
	}
}

func (sp *SliceProvider) NextStep() float64 {
	step := sp.steps[sp.stepIndex]

	sp.stepIndex++

	if sp.stepIndex >= len(sp.steps) {
		sp.stepIndex = 0
	}

	return step
}
