package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/oversampler"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/vartri"
	"github.com/almerlucke/muse/modules/waveshaper"
	"github.com/almerlucke/muse/modules/xfade"
	"github.com/dh1tw/gosamplerate"
)

func oscSyncHandler(index int, val any, shaper waveshaping.Shaper) {
	interpol := shaper.(*waveshaping.Interpolator)
	superSaw := interpol.Shapers[0].(*waveshaping.SuperSaw)
	hardSync := interpol.Shapers[1].(*waveshaping.HardSync)
	softSync := interpol.Shapers[2].(*waveshaping.SoftSyncTriangle)
	triMod := interpol.Shapers[3].(*waveshaping.JP8000triMod)

	switch index {
	case 0:
		superSaw.SetA1(val.(float64))
		hardSync.SetA1(val.(float64))
		softSync.SetA1(val.(float64))
	case 1:
		superSaw.SetM1(val.(float64))
	case 2:
		triMod.SetMod(val.(float64))
	case 3:
		interpol.Index = val.(float64)
	}
}

func main() {
	root := muse.New(2)

	config := muse.CurrentConfiguration()

	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 4 * config.SampleRate,
		BufferSize: config.BufferSize,
	})

	p := muse.NewPatch(0, 2)

	interpol := waveshaping.NewInterpolator(
		waveshaping.NewSuperSaw(1.5, 0.25, 0.88),
		waveshaping.NewHardSync(1.4),
		waveshaping.NewSoftSyncTriangle(1.25),
		waveshaping.NewJP8000triMod(0.7),
	)

	/*
		- xf between var tri and phasor
			- source
		- var tri duty width
			- var tri duty width
		- interpolation index -> super saw, hard sync, soft sync
			- shape
		- super saw A1, hard sync A1, soft sync A
			- sync
		- super saw M1
			- ripple
		- tri mod
			- modulation
	*/

	vt1 := vartri.New(150.0, 0.0, 0.25).AddTo(p)
	vt2 := vartri.New(152.0, 0.0, 0.25).AddTo(p)
	ph1 := phasor.New(151.0, 0.0).AddTo(p)
	ph2 := phasor.New(153.0, 0.0).AddTo(p)
	xf1 := xfade.New(0.0).AddTo(p).In(vt1, ph1)
	xf2 := xfade.New(0.0).AddTo(p).In(vt2, ph2)

	sh1 := waveshaper.New(interpol, 0, oscSyncHandler, nil).AddTo(p).In(xf1)
	sh2 := waveshaper.New(interpol, 0, nil, nil).AddTo(p).In(xf2)

	// flt1 := korg35.New(8500.0, 1.9, 1.1).AddTo(p).In(sh1)
	// flt2 := korg35.New(8500.0, 1.9, 1.1).AddTo(p).In(sh2)

	p.In(sh1, sh2)

	muse.PopConfiguration()

	lfoOscSync := lfo.NewBasicControlLFO(0.086, 1.01, 1.9).CtrlAddTo(root)
	lfoSuperSawM1 := lfo.NewBasicControlLFO(0.0532, 0.25, 0.75).CtrlAddTo(root)
	lfoTriMod := lfo.NewBasicControlLFO(0.084, 0.7, 1.0).CtrlAddTo(root)
	lfoInterpolator := lfo.NewBasicControlLFO(0.078, 0.0, 1.0).CtrlAddTo(root)
	lfoVarTriW1 := lfo.NewBasicControlLFO(0.063, 0.1, 0.9).CtrlAddTo(root)
	lfoVarTriW2 := lfo.NewBasicControlLFO(0.067, 0.1, 0.9).CtrlAddTo(root)
	// lfoFlt1 := lfo.NewBasicControlLFO(0.073, 100.0, 12000.0).CtrlAddTo(root)
	// lfoFlt2 := lfo.NewBasicControlLFO(0.077, 100.0, 12000.0).CtrlAddTo(root)
	lfoXf1 := lfo.NewBasicControlLFO(0.143, 0.0, 1.0).CtrlAddTo(root)
	lfoXf2 := lfo.NewBasicControlLFO(0.157, 0.0, 1.0).CtrlAddTo(root)

	vt1.CtrlIn(lfoVarTriW1, 0, 2)
	vt2.CtrlIn(lfoVarTriW2, 0, 2)
	xf1.CtrlIn(lfoXf1)
	xf2.CtrlIn(lfoXf2)
	// flt1.CtrlIn(lfoFlt1)
	// flt2.CtrlIn(lfoFlt2)
	sh1.CtrlIn(lfoOscSync, lfoSuperSawM1, lfoTriMod, lfoInterpolator)

	osa, _ := oversampler.New(p, gosamplerate.SRC_SINC_BEST_QUALITY)
	osa.AddTo(root)

	root.In(osa, osa, 1)

	// root.RenderToSoundFile("/Users/almerlucke/Desktop/wave_interpol.aifc", 40.0, root.Config.SampleRate, true)
	root.RenderAudio()
}
