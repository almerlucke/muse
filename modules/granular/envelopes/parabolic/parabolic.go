package parabolic

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/granular"
)

type Envelope struct {
	currentAmp float64
	slope      float64
	curve      float64
}

func (e *Envelope) New(_ any) granular.Envelope {
	return &Envelope{}
}

func (e *Envelope) Activate(amplitude float64, durationSamples int64, _ granular.Parameter, _ *muse.Configuration) {
	e.currentAmp = 0
	rdur := 1.0 / float64(durationSamples)
	rdur2 := rdur * rdur
	e.slope = 4.0 * amplitude * (rdur - rdur2)
	e.curve = -8.0 * amplitude * rdur2
}

func (e *Envelope) Synthesize(buf []float64, bufSize int) {
	for i := 0; i < bufSize; i++ {
		buf[i] = e.currentAmp
		e.currentAmp = e.currentAmp + e.slope
		e.slope = e.slope + e.curve
	}
}
