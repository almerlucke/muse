package lfo

import (
	"github.com/almerlucke/genny/float/phasor"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/lookup"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"github.com/almerlucke/muse"
)

type LFO struct {
	*muse.BaseModule
	phasor *phasor.Phasor
	table  lookup.Table
	lin    *linear.Linear
	ser    *series.Series
}

func New(fc float64, minVal float64, maxVal float64) *LFO {
	ph := phasor.New(fc, muse.CurrentConfiguration().SampleRate, 0.0)
	tab := lookup.NewNormalizedSineTable(512)
	lin := linear.New(maxVal-minVal, minVal)
	ser := series.New(tab, lin)

	l := &LFO{
		BaseModule: muse.NewBaseModule(0, 1),
		phasor:     ph,
		table:      tab,
		lin:        lin,
		ser:        ser,
	}

	l.SetSelf(l)

	return l
}

func (l *LFO) Synthesize() bool {
	if !l.BaseModule.Synthesize() {
		return false
	}

	for i := 0; i < l.Config.BufferSize; i++ {
		vs := l.phasor.Generate()
		l.Outputs[0].Buffer[i] = l.ser.Shape(vs)
	}

	return true
}
