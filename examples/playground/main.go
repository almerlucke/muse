package main

import (
	"math/rand"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/controls/val"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/functor"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"

	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/waveshaper"
	"github.com/almerlucke/muse/value"
)

func main() {
	env := muse.NewEnvironment(1, 44100, 512)

	harmonicsTimer := env.AddControl(timer.NewControlTimer(1500.0, env.Config, ""))
	freqTimer := env.AddControl(timer.NewControlTimer(375.0, env.Config, ""))

	freqGen := val.NewVal[float64](value.NewSequence[float64]([]float64{40.0, 60.0, 80.0, 120.0}), "")

	harmonicsGen := val.NewVal[map[int]float64](value.NewFunction[map[int]float64](func() map[int]float64 {
		harmonics := map[int]float64{}

		for i := 0; i < 20; i++ {
			maxAmp := (1.0-float64(i)/20.0)*0.5 + 0.5
			harmonics[rand.Intn(20)+1] = rand.Float64()*maxAmp*0.9 + 0.1
		}

		return harmonics
	}), "")

	harmonicsTimer.CtrlConnect(0, harmonicsGen, 0)
	freqTimer.CtrlConnect(0, freqGen, 0)

	adsrEnv := adsr.NewADSR(adsrc.StepsRatio(1.0, 0.1, 0.4, 0.1, 0.3, 0.5), adsrc.Ratio, adsrc.Automatic, 1.0, env.Config, "")
	adsrEnv.SetDuration(375.0)

	env.AddModule(adsrEnv)

	freqTimer.CtrlConnect(0, adsrEnv, 0)

	sine := env.AddModule(blosc.NewOsc(200.0, 0.0, env.Config, ""))

	shaper := env.AddModule(
		waveshaper.NewWaveShaper(
			waveshaping.NewChebyshev(
				map[int]float64{1: 1.0, 2: 0.8, 3: 0.7, 4: 0.6, 5: 0.5, 6: 0.4}), 1, func(index int, value any, shaper waveshaping.Shaper) {
				shaper.(*waveshaping.Chebyshev).SetHarmonics(value.(map[int]float64))
			}, nil, env.Config, "",
		),
	)

	freqGen.CtrlConnect(0, sine, 0)
	harmonicsGen.CtrlConnect(0, shaper, 0)

	mult := env.AddModule(functor.NewMult(2, env.Config))

	sine.Connect(0, shaper, 0)
	adsrEnv.Connect(0, mult, 0)
	shaper.Connect(0, mult, 1)
	mult.Connect(0, env, 0)

	env.QuickPlayAudio()
}
