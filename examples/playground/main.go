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
	env := muse.NewEnvironment(1, 44100, 1024)

	fm := fmsynth.NewFMSynth(18, waveshaping.NewSineTable(2048), env.Config, "fm")
	fm.OperatorSettings[1].Level = 0.5
	fm.OperatorSettings[5].Level = 0.5
	fm.PitchEnvLevels = [4]float64{0.49, 0.51, 0.495, 0.5}
	fm.PitchEnvRates = [4]float64{0.95, 0.95, 0.95, 0.95}
	fm.ReleaseMode = ops.EnvelopeDurationRelease
	fm.ApplySettingsChange()

	env.AddModule(fm)

	env.AddMessenger(banger.NewTemplateGenerator([]string{"fm"}, template.Template{
		"noteOn":   value.NewSequence([]any{36, 36, 48, 41, 51, 51, 49, 47, 32, 33}),
		"duration": value.NewSequence([]any{500.0, 300.0, 250.0, 150.0, 300.0, 125.0, 125.0, 500.0, 375.0}),
		"level":    value.NewSequence([]any{1.0}),
	}, "notes"))

	env.AddMessenger(timer.NewTimer(250.0, []string{"notes"}, env.Config, ""))

	fm.Connect(0, env, 0)

	env.QuickPlayAudio()
}
