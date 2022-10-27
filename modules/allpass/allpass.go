package allpass

import (
	"log"

	"github.com/almerlucke/muse"
	allpassc "github.com/almerlucke/muse/components/allpass"
)

type Allpass struct {
	*muse.BaseModule
	allpass        *allpassc.Allpass
	readLocation   float64
	readLocationMS float64
}

func NewAllpass(length float64, location float64, feedback float64, config *muse.Configuration, identifier string) *Allpass {
	return &Allpass{
		BaseModule:     muse.NewBaseModule(3, 1, config, identifier),
		allpass:        allpassc.NewAllpass(int(config.SampleRate*length*0.001), feedback),
		readLocation:   config.SampleRate * location * 0.001,
		readLocationMS: location,
	}
}

func (a *Allpass) Feedback() float64 {
	return a.allpass.Feedback
}

func (a *Allpass) SetFeedback(fb float64) {
	a.allpass.Feedback = fb
}

func (a *Allpass) ReadLocation() float64 {
	return a.readLocationMS
}

func (a *Allpass) SetReadLocation(readLocation float64) {
	a.readLocationMS = readLocation
	a.readLocation = readLocation * a.Config.SampleRate * 0.001
}

func (a *Allpass) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Read Location
		a.SetReadLocation(value.(float64))
	}
}

func (a *Allpass) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if readLocMS, ok := content["location"]; ok {
		a.SetReadLocation(readLocMS.(float64))
	}

	return nil
}

func (a *Allpass) Synthesize() bool {
	if !a.BaseModule.Synthesize() {
		return false
	}

	in := a.Inputs[0].Buffer
	out := a.Outputs[0].Buffer

	for i := 0; i < a.Config.BufferSize; i++ {
		if a.Inputs[1].IsConnected() {
			log.Printf("input 1 is connected")
			a.SetReadLocation(a.Inputs[1].Buffer[i])
		}

		if a.Inputs[2].IsConnected() {
			log.Printf("input 2 is connected")
			a.SetFeedback(a.Inputs[2].Buffer[i])
		}

		out[i] = a.allpass.Process(in[i], a.readLocation)
	}

	return true
}
