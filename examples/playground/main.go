package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/ops"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/fmsynth"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

func main() {
	root := muse.New(1)

	fm := fmsynth.New(18, waveshaping.NewSineTable(2048)).Named("fm").(*fmsynth.FMSynth)
	fm.OperatorSettings[1].Level = 0.5
	fm.OperatorSettings[5].Level = 0.5
	fm.PitchEnvLevels = [4]float64{0.49, 0.51, 0.495, 0.5}
	fm.PitchEnvRates = [4]float64{0.95, 0.95, 0.95, 0.95}
	fm.ReleaseMode = ops.EnvelopeDurationRelease
	fm.ApplySettingsChange()
	fm.Add(root)

	banger.NewTemplateGenerator([]string{"fm"}, template.Template{
		"noteOn":   value.NewSequence([]any{36, 36, 48, 41, 51, 51, 49, 47, 32, 33}),
		"duration": value.NewSequence([]any{500.0, 300.0, 250.0, 150.0, 300.0, 125.0, 125.0, 500.0, 375.0}),
		"level":    1.0,
	}).MsgrNamed("notes").MsgrAdd(root)

	timer.NewTimer(250.0, []string{"notes"}).MsgrAdd(root)

	root.In(fm)

	root.RenderLive()
}
