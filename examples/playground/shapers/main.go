package main

import (
	"math/rand"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/controls/val"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/value/template"

	"github.com/almerlucke/muse/modules/waveshaper"
	"github.com/almerlucke/muse/value"
)

func newShapeSwitcher() *waveshaping.Switch {
	return waveshaping.NewSwitch(0,
		waveshaping.NewSeries(
			waveshaping.NewSineTable(512),
			waveshaping.NewChebyshev(map[int]float64{1: 1.0, 2: 0.8, 3: 0.7, 4: 0.6, 5: 0.5, 6: 0.4}),
		),
		waveshaping.NewSoftSyncTriangle(1.25),
		waveshaping.NewSuperSaw(1.5, 0.25, 0.88),
	)
}

func switchControlFunction(index int, value any, shaper waveshaping.Shaper) {
	sw := shaper.(*waveshaping.Switch)
	shapeCtrlMap := value.(map[string]any)

	sw.Index = shapeCtrlMap["index"].(int)

	serial := sw.Shapers[0].(*waveshaping.Series)
	cheby := serial.Shapers[1].(*waveshaping.Chebyshev)
	cheby.SetHarmonics(shapeCtrlMap["harmonics"].(map[int]float64))

	tri := sw.Shapers[1].(*waveshaping.SoftSyncTriangle)
	tri.SetA1(shapeCtrlMap["tri"].(float64))

	superSaw := sw.Shapers[2].(*waveshaping.SuperSaw)
	superSawMap := shapeCtrlMap["superSaw"].(map[string]any)
	superSaw.SetA1(superSawMap["a1"].(float64))
	superSaw.SetM1(superSawMap["m1"].(float64))
	superSaw.SetM2(superSawMap["m2"].(float64))
}

func switchControlTemplate() template.Template {
	return template.Template{
		"index": value.NewSequence[any]([]any{0, 1, 2, 0, 1, 2, 0, 0, 2, 2, 1, 1}),
		"tri":   value.NewSequence[any]([]any{1.25, 1.75, 2.0, 2.2, 0.75, 1.0, 1.1}),
		"superSaw": template.Template{
			"a1": value.NewSequence[any]([]any{1.5, 1.75, 2.0, 2.2, 0.75, 1.0}),
			"m1": value.NewSequence[any]([]any{0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.4}),
			"m2": value.NewSequence[any]([]any{0.88, 0.7, 0.5, 0.95, 0.6}),
		},
		"harmonics": value.NewFunction[any](func() any {
			harmonics := map[int]float64{}

			for i := 0; i < 20; i++ {
				maxAmp := (1.0-float64(i)/20.0)*0.5 + 0.5
				harmonics[rand.Intn(20)+1] = rand.Float64()*maxAmp*0.9 + 0.1
			}

			return harmonics
		}),
	}
}

func main() {
	root := muse.New(1)

	timer1 := timer.NewControlTimer(250.0).CtrlAddTo(root)
	timer2 := timer.NewControlTimer(1500.0).CtrlAddTo(root)

	freqGen := val.New[float64](value.NewSequence[float64]([]float64{40.0, 60.0, 80.0, 90.0, 95.0}))
	shapersControlGen := banger.NewControlTemplateGenerator(switchControlTemplate())

	freqGen.CtrlIn(timer1)
	shapersControlGen.CtrlIn(timer2)

	phase := phasor.New(200.0, 0.0).AddTo(root)
	shaper := waveshaper.New(newShapeSwitcher(), 1, switchControlFunction, nil).AddTo(root)

	phase.CtrlIn(freqGen)
	shaper.In(phase)
	shaper.CtrlIn(shapersControlGen)
	root.In(shaper)

	//env.SynthesizeToFile("/Users/almerlucke/Desktop/notaliased.aiff", 10.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	root.RenderAudio()
}
