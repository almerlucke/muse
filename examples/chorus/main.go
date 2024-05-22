package main

import (
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/lookup"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/osc"
	"github.com/almerlucke/muse/utils/notes"
)

func makeLFO(speed float64, targets []string, shaper shape.Shaper, param string, templ template.Template) *lfo.LFO {
	ts := make([]*lfo.Target, len(targets))
	for i, target := range targets {
		ts[i] = lfo.NewTarget(target, shaper, param, templ)
	}

	return lfo.NewLFO(speed, ts)
}

func main() {
	root := muse.New(2)

	root.AddMessenger(banger.NewTemplateBang([]string{"osc"}, template.Template{
		"frequency": sequence.NewLoop(notes.A3.Freq(), notes.C3.Freq(), notes.E3.Freq()),
	}).MsgrNamed("sequencer"))

	root.AddMessenger(banger.NewTemplateBang([]string{"osc2"}, template.Template{
		"frequency": sequence.NewLoop(notes.A2.Freq(), notes.C2.Freq(), notes.E2.Freq()),
	}).MsgrNamed("sequencer2"))

	root.AddMessenger(stepper.NewStepper(
		sequence.NewLoop([]float64{250, -125, 250, 250, -125, 125, -125, 250}...),
		[]string{"sequencer", "sequencer2"},
	))

	sineTable := lookup.NewNormalizedSineTable(128)
	oscScale := linear.New(0.7, 0.1)
	mixScale := linear.New(1.0, 0.0)

	root.AddMessenger(makeLFO(0.24, []string{"osc"}, series.New(sineTable, oscScale), "pw", template.Template{
		"pulseWidth": template.NewParameter("pw", nil),
	}))
	root.AddMessenger(makeLFO(0.13, []string{"osc2"}, series.New(sineTable, oscScale), "pw", template.Template{
		"pulseWidth": template.NewParameter("pw", nil),
	}))
	root.AddMessenger(makeLFO(0.16, []string{"filter"}, series.New(sineTable, linear.New(16500.0, 1300.0)), "freq", template.Template{
		"frequency": template.NewParameter("freq", nil),
	}))

	root.AddMessenger(makeLFO(0.15, []string{"osc"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix1": template.NewParameter("mix", nil),
	}))
	root.AddMessenger(makeLFO(0.17, []string{"osc"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix2": template.NewParameter("mix", nil),
	}))
	root.AddMessenger(makeLFO(0.19, []string{"osc"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix3": template.NewParameter("mix", nil),
	}))
	root.AddMessenger(makeLFO(0.21, []string{"osc"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix4": template.NewParameter("mix", nil),
	}))

	root.AddMessenger(makeLFO(0.25, []string{"osc2"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix1": template.NewParameter("mix", nil),
	}))
	root.AddMessenger(makeLFO(0.37, []string{"osc2"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix2": template.NewParameter("mix", nil),
	}))
	root.AddMessenger(makeLFO(0.09, []string{"osc2"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix3": template.NewParameter("mix", nil),
	}))
	root.AddMessenger(makeLFO(0.11, []string{"osc2"}, series.New(sineTable, mixScale), "mix", template.Template{
		"mix4": template.NewParameter("mix", nil),
	}))

	osc1 := root.AddModule(osc.NewX(100.0, 0.0, 0.2, [4]float64{0.1, 0.1, 0.4, 0.1}).Named("osc"))
	osc2 := root.AddModule(osc.NewX(100.0, 0.5, 0.2, [4]float64{0.1, 0.1, 0.4, 0.1}).Named("osc2"))
	filter := root.AddModule(moog.New(8300.0, 0.63, 0.7))
	ch := root.AddModule(chorus.New(true, 56, 14, 0.2, 3.6, 0.6, lookup.NewSineTable(512.0)))

	osc1.Connect(4, filter, 0)
	osc2.Connect(4, filter, 0)
	filter.Connect(0, ch, 0)
	ch.Connect(0, root, 0)
	ch.Connect(0, root, 1)

	_ = root.RenderAudio()
}
