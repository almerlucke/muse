package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/ops"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/fmsynth"
	"github.com/almerlucke/muse/ui/controls"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

func strtof(s string) any {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func fmodeToString(mode ops.FrequencyMode) string {
	switch mode {
	case ops.FixedFrequency:
		return "fixed"
	default:
		return "track"
	}
}

func stringToFMode(s string) ops.FrequencyMode {
	switch s {
	case "fixed":
		return ops.FixedFrequency
	default:
		return ops.TrackFrequency
	}
}

func operatorControls(index int, setting fmsynth.OperatorSetting) *controls.Group {
	opGroup := controls.NewGroup(fmt.Sprintf("op.%d.group", index), fmt.Sprintf("Operator %d", index+1))
	opGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.level", index), "Level", 0.0, 1.0, 0.01, setting.Level))
	opGroup.AddControl(controls.NewEntry(fmt.Sprintf("op.%d.frequency", index), "Frequency", fmt.Sprintf("%f", setting.Frequency), strtof))
	opGroup.AddControl(controls.NewEntry(fmt.Sprintf("op.%d.ratio", index), "Ratio", fmt.Sprintf("%f", setting.FrequencyRatio), strtof))
	opGroup.AddControl(controls.NewRadio(fmt.Sprintf("op.%d.fmode", index), "Frequency Mode", []string{"fixed", "track"}, fmodeToString(setting.FrequencyMode)))
	opLevelGroup := controls.NewGroup("op.levels", "Env Levels")
	opLevelGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envLevel.1", index), "Env Level 1", 0.0, 1.0, 0.01, setting.LevelEnvLevels[0]))
	opLevelGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envLevel.2", index), "Env Level 2", 0.0, 1.0, 0.01, setting.LevelEnvLevels[1]))
	opLevelGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envLevel.3", index), "Env Level 3", 0.0, 1.0, 0.01, setting.LevelEnvLevels[2]))
	opLevelGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envLevel.4", index), "Env Level 4", 0.0, 1.0, 0.01, setting.LevelEnvLevels[3]))
	opGroup.AddChild(opLevelGroup)
	opRateGroup := controls.NewGroup("op1.rates", "Env Rates")
	opRateGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envRate.1", index), "Env Rate 1", 0.0, 1.0, 0.01, setting.LevelEnvRates[0]))
	opRateGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envRate.2", index), "Env Rate 2", 0.0, 1.0, 0.01, setting.LevelEnvRates[1]))
	opRateGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envRate.3", index), "Env Rate 3", 0.0, 1.0, 0.01, setting.LevelEnvRates[2]))
	opRateGroup.AddControl(controls.NewSlider(fmt.Sprintf("op.%d.envRate.4", index), "Env Rate 4", 0.0, 1.0, 0.01, setting.LevelEnvRates[3]))
	opGroup.AddChild(opRateGroup)
	return opGroup
}

func main() {
	root := muse.New(1)

	fm := fmsynth.New(18, waveshaping.NewSineTable(2048)).Named("fm").(*fmsynth.FMSynth)

	fm.OperatorSettings[1].Level = 0.5
	fm.OperatorSettings[5].Level = 0.5
	fm.PitchEnvLevels = [4]float64{0.49, 0.51, 0.495, 0.5}
	fm.PitchEnvRates = [4]float64{0.95, 0.95, 0.95, 0.95}
	fm.ReleaseMode = ops.EnvelopeDurationRelease
	fm.ApplySettingsChange()
	fm.AddTo(root)
	root.In(fm)

	opsGroup := controls.NewGroup("ops.group", "FM Synth")
	for i := 0; i < 6; i++ {
		opsGroup.AddChild(operatorControls(i, fm.OperatorSettings[i]))
	}

	opsGroup.AddListenerDeep(controls.NewChangeCallback(func(c controls.Control, old any, new any, setter any) {
		components := strings.Split(c.Identifier(), ".")
		if components[0] == "op" {
			opIndex, _ := strconv.Atoi(components[1])
			switch components[2] {
			case "level":
				fm.OperatorSettings[opIndex].Level = new.(float64)
			case "frequency":
				fm.OperatorSettings[opIndex].Frequency = new.(float64)
			case "ratio":
				fm.OperatorSettings[opIndex].FrequencyRatio = new.(float64)
			case "fmode":
				fm.OperatorSettings[opIndex].FrequencyMode = stringToFMode(new.(string))
			case "envLevel":
				levelIndex, _ := strconv.Atoi(components[3])
				fm.OperatorSettings[opIndex].LevelEnvLevels[levelIndex-1] = new.(float64)
			case "envRate":
				rateIndex, _ := strconv.Atoi(components[3])
				fm.OperatorSettings[opIndex].LevelEnvRates[rateIndex-1] = new.(float64)
			}
		}
		fm.ApplySettingsChange()
	}))

	root.AddMessenger(banger.NewTemplateGenerator([]string{"fm"}, template.Template{
		"noteOn":   value.NewSequence([]any{36, 36, 48, 41, 51, 51, 49, 47, 32, 33}),
		"duration": value.NewSequence([]any{500.0, 300.0, 250.0, 150.0, 300.0, 125.0, 125.0, 500.0, 375.0}),
		"level":    1.0,
	}).MsgrNamed("notes"))

	root.AddMessenger(timer.NewTimer(250.0, []string{"notes"}))

	err := root.InitializeAudio()
	if err != nil {
		log.Fatalf("error initializing audio: %v", err)
	}

	defer root.TerminateAudio()

	a := app.New()

	a.Settings().SetTheme(&theme.Theme{})

	w := a.NewWindow("Muse")

	w.Resize(fyne.Size{
		Width:  700,
		Height: 400,
	})

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
					// env.SynthesizeToFile("/Users/almerlucke/Desktop/waterFlow.aiff", 240.0, env.Config.SampleRate, sndfile.SF_FORMAT_AIFF)
					root.StartAudio()
				}),
				widget.NewButton("Stop", func() {
					root.StopAudio()
				}),
				// widget.NewButton("Notes Off", func() {
				// 	poly.(*polyphony.Polyphony).AllNotesOff()
				// }),
			),
			container.NewHScroll(opsGroup.UI()),
		),
	)

	w.ShowAndRun()
}
