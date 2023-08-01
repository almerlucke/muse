package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/oversampler"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/waveshaper"
	"github.com/dh1tw/gosamplerate"
)

func superSawParamHandler(index int, val any, shaper waveshaping.Shaper) {
	switch index {
	case 0:
		shaper.(*waveshaping.SuperSaw).SetA1(val.(float64))
	case 1:
		shaper.(*waveshaping.SuperSaw).SetM1(val.(float64))
	case 2:
		shaper.(*waveshaping.SuperSaw).SetM2(val.(float64))
	}
}

func main() {
	root := muse.New(1)

	config := muse.CurrentConfiguration()

	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 4 * config.SampleRate,
		BufferSize: config.BufferSize,
	})

	p := muse.NewPatch(0, 1)
	superSawShaper := waveshaping.NewSuperSaw(1.5, 0.25, 0.88)
	ph := phasor.New(100.0, 0.0).AddTo(p)
	sh := waveshaper.New(superSawShaper, 3, superSawParamHandler, nil).In(ph).AddTo(p)
	p.In(sh)

	muse.PopConfiguration()

	lfoA1 := lfo.NewBasicControlLFO(0.1, 0.5, 2.2).CtrlAddTo(root)
	lfoM1 := lfo.NewBasicControlLFO(0.05, 0.1, 0.6).CtrlAddTo(root)
	lfoM2 := lfo.NewBasicControlLFO(0.025, 0.5, 0.95).CtrlAddTo(root)
	sh.CtrlIn(lfoA1, lfoM1, lfoM2)

	osa, _ := oversampler.New(p, gosamplerate.SRC_SINC_BEST_QUALITY)
	osa.AddTo(root)
	root.In(osa)

	// root.RenderToSoundFile("/Users/almerlucke/Desktop/supersaw.aifc", 20.0, 44100.0, false)
	root.RenderAudio()
}
