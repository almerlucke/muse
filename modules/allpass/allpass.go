package allpass

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/allpassc"
)

type Allpass struct {
	*muse.BaseModule
	allpass      *allpassc.Allpass
	readLocation float64
}

func NewAllpass(length float64, location float64, feedback float64, config *muse.Configuration, identifier string) *Allpass {
	return &Allpass{
		BaseModule:   muse.NewBaseModule(3, 1, config, identifier),
		allpass:      allpassc.NewAllpass(int(config.SampleRate*length*0.001), feedback),
		readLocation: config.SampleRate * location * 0.001,
	}
}

func (a *Allpass) Synthesize() bool {
	if !a.BaseModule.Synthesize() {
		return false
	}

	in := a.Inputs[0].Buffer
	out := a.Outputs[0].Buffer

	for i := 0; i < a.Config.BufferSize; i++ {
		readLocation := a.readLocation
		if a.Inputs[1].IsConnected() {
			readLocation = a.Inputs[1].Buffer[i] * a.Config.SampleRate * 0.001
		}

		if a.Inputs[2].IsConnected() {
			a.allpass.Feedback = a.Inputs[2].Buffer[i]
		}

		out[i] = a.allpass.Process(in[i], readLocation)
	}

	return true
}
