package main

import (
	"fmt"
	"github.com/almerlucke/muse/modules/effects/chorus"
	"github.com/almerlucke/muse/modules/filters/moog"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/synths/classic"
	"github.com/almerlucke/muse/ui/controls"
	"github.com/almerlucke/muse/utils/notes"

	museMidi "github.com/almerlucke/muse/midi"

	"gitlab.com/gomidi/midi"
	. "gitlab.com/gomidi/midi/midimessage/channel" // (Channel Messages)
	"gitlab.com/gomidi/midi/reader"
)

type ClassicSynth struct {
	*muse.BasePatch
	controls         *controls.Group
	ampEnvSetting    *adsr.Setting
	filterEnvSetting *adsr.Setting
	Synth            *classic.Synth
	chorus1          *chorus.Chorus
	chorus2          *chorus.Chorus
}

func NewClassicSynth(bpm float64) *ClassicSynth {
	synth := &ClassicSynth{
		BasePatch: muse.NewPatch(0, 2),
		controls:  controls.NewGroup("group.main", "Classic Synth"),
	}

	synth.SetIdentifier("synth")
	synth.SetSelf(synth)

	// Add self as receiver
	synth.AddMessageReceiver(synth, "synth")

	ampEnvSetting := adsr.NewSetting(1.0, 5.0, 0.3, 5.0, 0.0, 2000.0)
	filterEnvSetting := adsr.NewSetting(1.0, 3.0, 0.3, 5.0, 0.0, 2000.0)

	synth.ampEnvSetting = ampEnvSetting
	synth.filterEnvSetting = filterEnvSetting
	synth.Synth = classic.New(40, ampEnvSetting, ampEnvSetting, &moog.Factory{}, moog.DefaultConfig()).Named("poly").AddTo(synth).(*classic.Synth)
	synth.chorus1 = chorus.New(0.24, 0.83, 0.76, 0.2, 1.0, 0.75, nil).AddTo(synth).In(synth.Synth, synth.Synth, 1).(*chorus.Chorus)
	synth.In(synth.chorus1, synth.chorus1, 1)

	return synth
}

func (cs *ClassicSynth) SetupControls(w fyne.Window) {
	filterGroup := cs.controls.AddChild(controls.NewGroup("group.filter", "Filter"))
	filterGroup.AddControl(controls.NewSlider("voice.filterFcMin", "Filter Frequency Min", 50.0, 8000.0, 1.0, 50.0))
	filterGroup.AddControl(controls.NewSlider("voice.filterFcMax", "Filter Frequency Max", 50.0, 8000.0, 1.0, 8000.0))
	filterGroup.AddControl(controls.NewSlider("voice.filterResonance", "Resonance", 0.0, 1.0, 0.01, 0.7))

	mixerGroup := cs.controls.AddChild(controls.NewGroup("group.mixer", "Mixer"))
	mixerGroup.AddControl(controls.NewSlider("voice.osc1Mix", "Osc1 Mix", 0.0, 1.0, 0.01, 0.6))
	mixerGroup.AddControl(controls.NewSlider("voice.osc2Mix", "Osc2 Mix", 0.0, 1.0, 0.01, 0.35))
	mixerGroup.AddControl(controls.NewSlider("voice.noiseMix", "Noise Mix", 0.0, 1.0, 0.01, 0.05))

	mixerGroup.AddControl(controls.NewRadio("test.radioButton", "Radio Button", []string{"selection1", "selection2", "selection3"}, "selection1"))

	mixerGroup.AddControl(controls.NewEntry("test.entry", "This is an entry", "3.4", func(v string) any {
		fv, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return fv
		}

		return 0.0
	}))

	mixerGroup.AddControl(controls.NewFilePicker("test.filePicker", "File Picker", w))

	osc1Group := cs.controls.AddChild(controls.NewGroup("group.osc1", "Oscillator 1"))
	osc1Group.AddControl(controls.NewSlider("voice.osc1PulseWidth", "Pulse Width", 0.0, 1.0, 0.01, 0.5))
	osc1Group.AddControl(controls.NewSlider("voice.osc1SineMix", "Sine Mix", 0.0, 1.0, 0.01, 0.0))
	osc1Group.AddControl(controls.NewSlider("voice.osc1SawMix", "Saw Mix", 0.0, 1.0, 0.01, 0.0))
	osc1Group.AddControl(controls.NewSlider("voice.osc1PulseMix", "Pulse Mix", 0.0, 1.0, 0.01, 1.0))
	osc1Group.AddControl(controls.NewSlider("voice.osc1TriMix", "Tri Mix", 0.0, 1.0, 0.01, 0.0))

	osc2Group := cs.controls.AddChild(controls.NewGroup("group.osc2", "Oscillator 2"))
	osc2Group.AddControl(controls.NewSlider("voice.osc2PulseWidth", "Pulse Width", 0.0, 1.0, 0.01, 0.5))
	osc2Group.AddControl(controls.NewSlider("voice.osc2SineMix", "Sine Mix", 0.0, 1.0, 0.01, 0.0))
	osc2Group.AddControl(controls.NewSlider("voice.osc2SawMix", "Saw Mix", 0.0, 1.0, 0.01, 0.0))
	osc2Group.AddControl(controls.NewSlider("voice.osc2PulseMix", "Pulse Mix", 0.0, 1.0, 0.01, 1.0))
	osc2Group.AddControl(controls.NewSlider("voice.osc2TriMix", "Tri Mix", 0.0, 1.0, 0.01, 0.0))
	osc2Group.AddControl(controls.NewSlider("voice.osc2Tuning", "Tuning", 0.125, 8.0, 0.01, 2.0))

	panGroup := cs.controls.AddChild(controls.NewGroup("group.pan", "Pan"))
	panGroup.AddControl(controls.NewSlider("voice.pan", "Pan", 0.0, 1.0, 0.01, 0.5))

	ampEnvGroup := cs.controls.AddChild(controls.NewGroup("group.ampEnv", "Amplitude ADSR"))
	ampEnvGroup.AddControl(controls.NewSlider("adsr.amplitude.attackLevel", "Attack Level", 0.0, 1.0, 0.01, 1.0))
	ampEnvGroup.AddControl(controls.NewSlider("adsr.amplitude.attackDuration", "Attack Duration (ms)", 2.0, 1000.0, 1.0, 25.0))
	ampEnvGroup.AddControl(controls.NewSlider("adsr.amplitude.decayLevel", "Decay Level", 0.0, 1.0, 0.01, 0.3))
	ampEnvGroup.AddControl(controls.NewSlider("adsr.amplitude.decayDuration", "Decay Duration (ms)", 2.0, 1000.0, 1.0, 80.0))
	ampEnvGroup.AddControl(controls.NewSlider("adsr.amplitude.releaseDuration", "Release Duration (ms)", 5.0, 4000.0, 1.0, 2000.0))

	filterEnvGroup := cs.controls.AddChild(controls.NewGroup("group.filterEnv", "Filter ADSR"))
	filterEnvGroup.AddControl(controls.NewSlider("adsr.filter.attackLevel", "Attack Level", 0.0, 1.0, 0.01, 0.9))
	filterEnvGroup.AddControl(controls.NewSlider("adsr.filter.attackDuration", "Attack Duration (ms)", 2.0, 1000.0, 1.0, 25.0))
	filterEnvGroup.AddControl(controls.NewSlider("adsr.filter.decayLevel", "Decay Level", 0.0, 1.0, 0.01, 0.5))
	filterEnvGroup.AddControl(controls.NewSlider("adsr.filter.decayDuration", "Decay Duration (ms)", 2.0, 1000.0, 1.0, 80.0))
	filterEnvGroup.AddControl(controls.NewSlider("adsr.filter.releaseDuration", "Release Duration (ms)", 5.0, 4000.0, 1.0, 2000.0))

	cs.controls.AddListenerDeep(cs)
}

func (cs *ClassicSynth) ControlChanged(ctrl controls.Control, oldValue any, newValue any, setter any) {
	id := ctrl.Identifier()
	components := strings.Split(id, ".")
	route := components[0]

	if route == "voice" {
		// If voice control send through to polyphony module (which will pass message to voices)
		cs.Synth.ReceiveMessage(map[string]any{
			"command":     "voice",
			components[1]: newValue,
		})
	} else if route == "adsr" {
		// If adsr set steps
		var setting *adsr.Setting
		if components[1] == "filter" {
			setting = cs.filterEnvSetting
		} else if components[1] == "amplitude" {
			setting = cs.ampEnvSetting
		}

		if setting != nil {
			switch components[2] {
			case "attackLevel":
				setting.AttackLevel = newValue.(float64)
			case "attackDuration":
				setting.AttackDuration = newValue.(float64)
			case "decayLevel":
				setting.DecayLevel = newValue.(float64)
			case "decayDuration":
				setting.DecayDuration = newValue.(float64)
			case "releaseDuration":
				setting.ReleaseDuration = newValue.(float64)
			}
		}
	} else {
		log.Printf("id %v %v", id, newValue)
	}
}

func (cs *ClassicSynth) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	for k, v := range content {
		ctrl := cs.controls.ControlById(k)
		if ctrl != nil {
			if ctrl.Type() == controls.SliderType {
				// Change control from message will take a lot of cpy because it updates fyne UI elements, which is really inefficient
				// ctrl.(*control.SliderControl).Set(v.(float64), nil)
				cs.ControlChanged(ctrl, 0, v.(float64), nil)
			}
		}
	}

	return nil
}

func main() {
	root := muse.New(2)

	bpm := 100.0
	synth := NewClassicSynth(bpm)

	synth.AddTo(root)
	root.In(synth, synth, 1)

	// synth.AddMessenger(banger.NewTemplateGenerator([]string{"poly"}, template.Template{
	// 	"command":   "trigger",
	// 	"duration":  value.NewSequence([]any{375.0, 750.0, 1000.0, 250.0, 250.0, 375.0, 750.0}),
	// 	"amplitude": value.NewConst[any](1.0),
	// 	"message": template.Template{
	// 		// "osc1SineMix":  value.NewSequence([]any{1.0, 0.7, 0.5, 0.3, 0.1, 0.0, 0.1, 0.2, 0.3, 0.5, 0.7}),
	// 		// "osc1SawMix":   value.NewSequence([]any{0.0, 0.1, 0.3, 0.5, 0.7, 1.0, 0.7, 0.5, 0.3, 0.2, 0.1}),
	// 		// "osc1PulseMix": value.NewSequence([]any{0.2, 0.5, 0.7, 1.0, 0.7, 0.5, 0.2, 0.1, 0.0}),
	// 		// "osc1TriMix":   value.NewSequence([]any{0.7, 1.0, 1.0, 0.7, 0.3, 0.1, 0.0, 0.0, 0.0}),
	// 		// "osc2SineMix":  value.NewSequence([]any{0.7, 1.0, 1.0, 0.7, 0.3, 0.1, 0.0, 0.0, 0.0}),
	// 		// "osc2SawMix":   value.NewSequence([]any{0.2, 0.5, 0.7, 1.0, 0.7, 0.5, 0.2, 0.1, 0.0}),
	// 		// "osc2PulseMix": value.NewSequence([]any{0.0, 0.1, 0.3, 0.5, 0.7, 1.0, 0.7, 0.5, 0.3, 0.2, 0.1}),
	// 		// "osc2TriMix":   value.NewSequence([]any{1.0, 0.7, 0.5, 0.3, 0.1, 0.0, 0.1, 0.2, 0.3, 0.5, 0.7}),
	// 		"frequency": noteSequence(notes.O3),
	// 	},
	// }, "control"))

	// synth.AddMessenger(stepper.NewStepper(
	// 	swing.New(value.NewConst(bpm), value.NewConst(2.0),
	// 		value.NewSequence([]*swing.Step{{}, {Skip: true}}),
	// 	),
	// 	[]string{"control"}, "",
	// ))

	// synth.AddMessenger(lfo.NewBasicLFO(0.14, 0.7, 0.15, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.osc1PulseWidth": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.103, 0.7, 0.15, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.osc2PulseWidth": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.085, 0.6, 0.25, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.filterResonance": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.115, 0.06, 4.0, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.osc2Tuning": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.0367, 0.1, 0.01, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.noiseMix": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.0567, 0.4, 0.3, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.osc1Mix": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.0667, 0.4, 0.2, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.osc2Mix": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.1067, 0.3, 0.35, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.pan": template.NewParameter("val", nil),
	// }))

	// synth.AddMessenger(lfo.NewBasicLFO(0.0569, 6800.0, 1200.0, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.filterFcMax": template.NewParameter("val", nil),
	// }))

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/classic_synth.aiff", 360.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)

	listener, err := museMidi.NewListener(0, reader.Each(func(pos *reader.Position, msg midi.Message) {

		// inspect
		log.Println(msg)

		switch v := msg.(type) {
		case NoteOn:
			if v.Velocity() == 0 {
				log.Printf("velocity 0")
				synth.Synth.ReceiveMessage(map[string]any{
					"command": "trigger",
					"noteOff": fmt.Sprintf("%d", v.Key()),
				})
			} else {
				velocity := float64(v.Velocity()) / 127.0
				log.Printf("velocity %v", velocity)
				synth.Synth.ReceiveMessage(map[string]any{
					"command":   "trigger",
					"noteOn":    fmt.Sprintf("%d", v.Key()),
					"amplitude": velocity,
					"message": map[string]any{
						"frequency": notes.Mtof(int(v.Key())),
					},
				})
			}
			synth.Synth.DebugActive()
		case NoteOff:
			synth.Synth.DebugActive()
			synth.Synth.ReceiveMessage(map[string]any{
				"command": "trigger",
				"noteOff": fmt.Sprintf("%d", v.Key()),
			})
			log.Printf("note off")
			synth.Synth.DebugActive()
		}
	}))
	if err != nil {
		log.Fatalf("error opening midi listener, %v", err)
	}

	defer listener.Close()

	_ = root.RenderAudio()

	//err = root.InitializeAudio()
	//if err != nil {
	//	log.Fatalf("error opening audio stream, %v", err)
	//}
	//
	//defer root.TerminateAudio()
	//
	//a := app.New()
	//
	//a.Settings().SetTheme(&theme.Theme{})
	//
	//w := a.NewWindow("Muse")
	//
	//synth.SetupControls(w)
	//
	//w.Resize(fyne.Size{
	//	Width:  1200,
	//	Height: 400,
	//})
	//
	//w.SetContent(
	//	container.NewVBox(
	//		container.NewHBox(
	//			widget.NewButton("Start", func() {
	//				_ = root.StartAudio()
	//			}),
	//			widget.NewButton("Stop", func() {
	//				_ = root.StopAudio()
	//			}),
	//		),
	//		container.NewHScroll(synth.controls.UI()),
	//	),
	//)
	//
	//w.ShowAndRun()
}
