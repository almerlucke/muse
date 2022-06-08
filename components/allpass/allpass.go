package allpass

import "github.com/almerlucke/muse/components/delay"

type Allpass struct {
	*delay.Delay
	Feedback float64
}

func NewAllpass(length int, feedback float64) *Allpass {
	return &Allpass{
		Delay:    delay.NewDelay(length),
		Feedback: feedback,
	}
}

func (allpass *Allpass) Process(xn float64, location float64) float64 {
	vm := allpass.Delay.Read(location)
	vn := xn - allpass.Feedback*vm
	yn := vn*allpass.Feedback + vm

	allpass.Delay.Write(vn)

	return yn
}
