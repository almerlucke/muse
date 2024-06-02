package main

import (
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/emulations/supersaw"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/freeverb"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/vartri"
	"github.com/almerlucke/muse/modules/waveshaper"
	"github.com/almerlucke/muse/utils"
)

func msgMapper(msg any, shaper shape.Shaper) {
	// params, ok := msg.(map[string]any)
	// if ok {
	// 	s, ok := params["sync"]
	// 	if ok {
	// 	}
	// }
}

func paramMapper(param int, value any, shaper shape.Shaper) {
	if param == 0 {
		shaper.(*supersaw.SuperSaw).SetM1(value.(float64))
	}
}

func main() {
	root := muse.New(2)

	seq := sequence.NewLoop(utils.ReadJSONNull[[][]*muse.Message]("examples/shaping/sequence1.json")...)
	root.AddMessenger(banger.NewGenBang(seq).MsgrNamed("sequencer1"))

	root.AddMessenger(stepper.NewStepper(
		sequence.NewLoop(250.0, -125.0, 250.0, 250.0, -125.0, 125.0, -125.0, 250.0),
		[]string{"sequencer1", "adsr1"},
	))

	setting := adsrc.NewSetting(1.0, 20.0, 0.4, 20.0, 20.0, 100.0)

	paramVarTri1 := vartri.New(0.25, 0.0, 0.5).AddTo(root)
	paramVarTri2 := vartri.New(0.325, 0.0, 0.5).AddTo(root)
	superSawParam := root.AddModule(functor.NewScale(0.82, 0.15))
	adsrEnv1 := root.AddModule(adsr.New(setting, adsrc.Automatic, 1.0).Named("adsr1"))
	mult1 := root.AddModule(functor.NewMult(2))
	filterParam := root.AddModule(functor.NewScale(2200.0, 40.0))
	osc1 := root.AddModule(phasor.New(140.0, 0.0).Named("osc1"))
	shaper1 := root.AddModule(waveshaper.New(supersaw.New(1.5, 0.25, 0.88), 1, paramMapper, nil))
	allp := root.AddModule(allpass.New(375.0, 375.0, 0.3))
	allpassAmp := root.AddModule(functor.NewAmp(0.5))
	// filter := env.AddModule(butterworth.NewButterworth(300.0, 0.4, env.Config, "filter"))
	// filter := env.AddModule(rbj.NewRBJFilter(rbjc.Lowpass, 300.0, 10.0, env.Config, "filter"))
	filter := root.AddModule(moog.New(300.0, 0.45, 1.0))
	reverb := freeverb.New().AddTo(root)

	reverb.(*freeverb.FreeVerb).SetDamp(0.1)
	reverb.(*freeverb.FreeVerb).SetDry(0.7)
	reverb.(*freeverb.FreeVerb).SetWet(0.1)
	reverb.(*freeverb.FreeVerb).SetRoomSize(0.8)
	reverb.(*freeverb.FreeVerb).SetWidth(0.8)

	paramVarTri1.Connect(0, superSawParam, 0)
	paramVarTri2.Connect(0, filterParam, 0)
	osc1.Connect(0, shaper1, 0)
	superSawParam.Connect(0, shaper1, 1)
	shaper1.Connect(0, mult1, 0)
	adsrEnv1.Connect(0, mult1, 1)
	mult1.Connect(0, filter, 0)
	filterParam.Connect(0, filter, 1)
	filter.Connect(0, allp, 0)
	allp.Connect(0, allpassAmp, 0)
	filter.Connect(0, reverb, 0)
	allpassAmp.Connect(0, reverb, 1)
	reverb.Connect(0, root, 0)
	reverb.Connect(1, root, 1)

	_ = root.RenderAudio()

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/shaper.aiff", 24.0, 44100.0, true, sndfile.SF_FORMAT_AIFF)
}
