package main

import (
	"math/rand"
	"strings"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/synths/classic"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/ui/controls"
	"github.com/almerlucke/muse/utils"
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/arpeggio"
	"github.com/almerlucke/muse/value/markov"
	"github.com/almerlucke/muse/value/template"
)

type ClassicSynth struct {
	*muse.BasePatch
	controls         *controls.Group
	ampEnvSetting    *adsr.Setting
	filterEnvSetting *adsr.Setting
	poly             *polyphony.Polyphony
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

	ampEnvSetting := adsr.NewSetting(1.0, 25.0, 0.3, 80.0, 0.0, 2000.0)
	filterEnvSetting := adsr.NewSetting(0.9, 25.0, 0.5, 80.0, 0.0, 2000.0)

	synth.ampEnvSetting = ampEnvSetting
	synth.filterEnvSetting = filterEnvSetting
	synth.poly = classic.New(20, ampEnvSetting, filterEnvSetting).Named("poly").AddTo(synth).(*polyphony.Polyphony)
	synthAmp1 := functor.NewAmp(0.85).AddTo(synth).In(synth.poly)
	synthAmp2 := functor.NewAmp(0.85).AddTo(synth).In(synth.poly, 1)
	allpass1 := allpass.New(2500.0, 60000/bpm*1.666, 0.5).AddTo(synth).In(synthAmp1)
	allpass2 := allpass.New(2500.0, 60000/bpm*1.75, 0.4).AddTo(synth).In(synthAmp2)
	allpassAmp1 := functor.NewAmp(0.5).AddTo(synth).In(allpass1)
	allpassAmp2 := functor.NewAmp(0.5).AddTo(synth).In(allpass2)
	synth.chorus1 = chorus.New(false, 15, 10, 0.3, 1.42, 0.5, nil).AddTo(synth).In(synthAmp1, allpassAmp1, 0, 0).(*chorus.Chorus)
	synth.chorus2 = chorus.New(false, 15, 10, 0.31, 1.43, 0.55, nil).AddTo(synth).In(synthAmp2, allpassAmp2, 0, 0).(*chorus.Chorus)

	synth.In(synth.chorus1, synth.chorus2)
	synth.SetupControls()

	return synth
}

func (cs *ClassicSynth) SetupControls() {
	filterGroup := cs.controls.AddChild(controls.NewGroup("group.filter", "Filter"))
	filterGroup.AddControl(controls.NewSlider("voice.filterFcMin", "Filter Frequency Min", 50.0, 8000.0, 1.0, 50.0))
	filterGroup.AddControl(controls.NewSlider("voice.filterFcMax", "Filter Frequency Max", 50.0, 8000.0, 1.0, 8000.0))
	filterGroup.AddControl(controls.NewSlider("voice.filterResonance", "Resonance", 0.01, 2.0, 0.01, 0.7))

	mixerGroup := cs.controls.AddChild(controls.NewGroup("group.mixer", "Mixer"))
	mixerGroup.AddControl(controls.NewSlider("voice.osc1Mix", "Osc1 Mix", 0.0, 1.0, 0.01, 0.6))
	mixerGroup.AddControl(controls.NewSlider("voice.osc2Mix", "Osc2 Mix", 0.0, 1.0, 0.01, 0.35))
	mixerGroup.AddControl(controls.NewSlider("voice.noiseMix", "Noise Mix", 0.0, 1.0, 0.01, 0.05))

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

	if components[0] == "voice" {
		// If voice control send through to polyphony module (which will pass message to voices)
		cs.poly.ReceiveMessage(map[string]any{
			"command":     "voice",
			components[1]: newValue,
		})
	} else if components[0] == "adsr" {
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

func noteSequence(octave notes.Note) value.Valuer[any] {
	return value.NewAnd(
		[]value.Valuer[any]{
			// Row 1
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, true), 2, 2),
			// Row 2
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMajor7.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Exclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.EMinor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 2, 2),
			// Row 3
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, true), 2, 2),
			// Row 4
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMajor7.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Exclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.BMinor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 2),
			// Row 5
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.EMinor.FreqAny(octave), arpeggio.Converge, arpeggio.None, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 1, 1),
		}, true)
}

func addDrumTrack(p muse.Patch, polyName string, sounds []string, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) {
	identifier := sounds[0] + "Drum"

	p.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}))

	p.AddMessenger(banger.NewTemplateGenerator([]string{polyName}, template.Template{
		"command":   "trigger",
		"duration":  0.0,
		"amplitude": amp,
		"message": template.Template{
			"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
			"sound": value.NewSequence(utils.ToAnySlice(sounds)),
		},
	}).MsgrNamed(identifier))
}

func kickRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm5 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 3.0, rhythm4, 2.0, rhythm5, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 4.0, rhythm5, 2.0, rhythm1, 1.0)
	rhythm5.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm5, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	m := markov.NewMarkov[[]*swing.Step](markov.NewStateStarter(rhythm1), 1)

	return value.NewFlatten[*swing.Step](m)
}

func snareRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {BurstChance: 0.5, NumBurst: 3}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {BurstChance: 0.5, NumBurst: 3}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm5 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {BurstChance: 0.5, NumBurst: 3},
		{Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 3.0, rhythm4, 2.0, rhythm5, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 4.0, rhythm5, 2.0, rhythm1, 1.0)
	rhythm5.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm5, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	m := markov.NewMarkov[[]*swing.Step](markov.NewStateStarter(rhythm1), 1)

	return value.NewFlatten[*swing.Step](m)
}

func bassRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {},
		{Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 3.0, rhythm4, 2.0, rhythm1, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	m := markov.NewMarkov[[]*swing.Step](markov.NewStateStarter(rhythm1), 1)

	return value.NewFlatten[*swing.Step](m)
}

func hihatRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1}, {}, {Skip: true}, {}, {Shuffle: 0.1}, {}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Skip: true}, {}, {Shuffle: 0.1}, {}, {Skip: true}, {}, {Shuffle: 0.1},
	}, nil)

	rhythm5 := markov.NewState([]*swing.Step{
		{}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 2.0, rhythm2, 1.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 2.0, rhythm3, 1.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 2.0, rhythm4, 1.0, rhythm5, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 3.0, rhythm5, 1.0, rhythm1, 1.0)
	rhythm5.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm5, 3.0, rhythm1, 1.0, rhythm2, 1.0)

	m := markov.NewMarkov[[]*swing.Step](markov.NewStateStarter(rhythm1), 1)

	return value.NewFlatten[*swing.Step](m)
}

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 256,
	})

	root := muse.New(2)

	bpm := 80.0

	soundBank := io.SoundBank{}

	soundBank["hihat"], _ = io.NewSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	soundBank["kick"], _ = io.NewSoundFile("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav")
	soundBank["snare"], _ = io.NewSoundFile("resources/drums/snare/Cymatics - Humble Adequate Snare - E.wav")
	soundBank["808_1"], _ = io.NewSoundFile("resources/drums/808/Cymatics - Humble 808 4 - F.wav")
	soundBank["808_2"], _ = io.NewSoundFile("resources/drums/808/Cymatics - Humble 808 3 - F.wav")
	soundBank["808_3"], _ = io.NewSoundFile("resources/drums/fx/Cymatics - Orchid Impact FX 2.wav")
	soundBank["808_4"], _ = io.NewSoundFile("resources/drums/fx/Cymatics - Orchid Reverse Crash 2.wav")
	soundBank["shaker"], _ = io.NewSoundFile("resources/drums/shots/Cymatics - Orchid Shaker - Drew.wav")

	drumMachine := drums.NewDrums(soundBank, 20).Named("drums").AddTo(root)

	addDrumTrack(root, "drums", []string{"hihat"}, int(bpm), 8, 0.875, 1.125, 0.6, hihatRhythm())
	addDrumTrack(root, "drums", []string{"kick"}, int(bpm), 8, 0.875, 1.125, 1.0, kickRhythm())
	addDrumTrack(root, "drums", []string{"snare"}, int(bpm), 8, 1.0, 1.0, 0.7, snareRhythm())
	addDrumTrack(root, "drums", []string{"808_1", "808_2", "808_3", "808_4"}, int(bpm), 2, 1.0, 1.0, 0.7, bassRhythm())
	addDrumTrack(root, "drums", []string{"shaker"}, int(bpm), 2, 1.0, 1.0, 1.0, kickRhythm())

	synth := NewClassicSynth(bpm).AddTo(root).(*ClassicSynth)

	root.In(
		drumMachine, 0,
		drumMachine, 1,
		modules.Amp(synth, 0, 0.7).AddTo(root), 0, 0,
		modules.Amp(synth, 1, 0.7).AddTo(root), 0, 1,
	)

	synth.AddMessenger(banger.NewTemplateGenerator([]string{"poly"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{375.0, 750.0, 1000.0, 250.0, 250.0, 375.0, 750.0}),
		"amplitude": value.NewConst[any](1.0),
		"message": template.Template{
			"osc1SineMix":  value.NewSequence([]any{1.0, 0.7, 0.5, 0.3, 0.1, 0.0, 0.1, 0.2, 0.3, 0.5, 0.7}),
			"osc1SawMix":   value.NewSequence([]any{0.0, 0.1, 0.3, 0.5, 0.7, 1.0, 0.7, 0.5, 0.3, 0.2, 0.1}),
			"osc1PulseMix": value.NewSequence([]any{0.2, 0.5, 0.7, 1.0, 0.7, 0.5, 0.2, 0.1, 0.0}),
			"osc1TriMix":   value.NewSequence([]any{0.7, 1.0, 1.0, 0.7, 0.3, 0.1, 0.0, 0.0, 0.0}),
			"osc2SineMix":  value.NewSequence([]any{0.7, 1.0, 1.0, 0.7, 0.3, 0.1, 0.0, 0.0, 0.0}),
			"osc2SawMix":   value.NewSequence([]any{0.2, 0.5, 0.7, 1.0, 0.7, 0.5, 0.2, 0.1, 0.0}),
			"osc2PulseMix": value.NewSequence([]any{0.0, 0.1, 0.3, 0.5, 0.7, 1.0, 0.7, 0.5, 0.3, 0.2, 0.1}),
			"osc2TriMix":   value.NewSequence([]any{1.0, 0.7, 0.5, 0.3, 0.1, 0.0, 0.1, 0.2, 0.3, 0.5, 0.7}),
			"frequency":    noteSequence(notes.O3),
		},
	}).MsgrNamed("control"))

	synth.AddMessenger(stepper.NewStepper(
		swing.New(int(bpm), 4,
			value.NewSequence([]*swing.Step{{}, {Skip: true}}),
		),
		[]string{"control"},
	))

	synth.AddMessenger(lfo.NewBasicLFO(0.14, 0.7, 0.15, []string{"synth"}, "val", template.Template{
		"voice.osc1PulseWidth": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.103, 0.7, 0.15, []string{"synth"}, "val", template.Template{
		"voice.osc2PulseWidth": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.085, 0.8, 0.25, []string{"synth"}, "val", template.Template{
		"voice.filterResonance": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.115, 0.08, 4.0, []string{"synth"}, "val", template.Template{
		"voice.osc2Tuning": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0367, 0.1, 0.01, []string{"synth"}, "val", template.Template{
		"voice.noiseMix": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0567, 0.4, 0.3, []string{"synth"}, "val", template.Template{
		"voice.osc1Mix": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0667, 0.4, 0.2, []string{"synth"}, "val", template.Template{
		"voice.osc2Mix": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.1067, 0.3, 0.35, []string{"synth"}, "val", template.Template{
		"voice.pan": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0767, 20.0, 5.0, []string{"synth"}, "val", template.Template{
		"adsr.filter.attackDuration": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0867, 0.2, 0.1, []string{"synth"}, "val", template.Template{
		"adsr.filter.decayLevel": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0817, 120.0, 10.0, []string{"synth"}, "val", template.Template{
		"adsr.filter.decayDuration": template.NewParameter("val", nil),
	}))

	root.RenderAudio()

	// root.RenderToSoundFile("/Users/almerlucke/Desktop/classic_control.aifc", 240.0, 44100.0, true)

	// synth.AddMessenger(lfo.NewBasicLFO(0.0569, 6800.0, 1200.0, []string{"synth"}, env.Config, "val", template.Template{
	// 	"voice.filterFcMax": template.NewParameter("val", nil),
	// }))

	// root.RenderToSoundFile("/Users/almerlucke/Desktop/classic_synth.aiff", 240.0, 44100.0, true, sndfile.SF_FORMAT_AIFF)

	// err := root.InitializeAudio()
	// if err != nil {
	// 	log.Fatalf("error opening audio stream, %v", err)
	// }

	// defer root.TerminateAudio()

	// a := app.New()

	// a.Settings().SetTheme(&theme.Theme{})

	// w := a.NewWindow("Muse")

	// w.Resize(fyne.Size{
	// 	Width:  700,
	// 	Height: 400,
	// })

	// w.SetContent(
	// 	container.NewVBox(
	// 		container.NewHBox(
	// 			widget.NewButton("Start", func() {
	// 				root.StartAudio()
	// 			}),
	// 			widget.NewButton("Stop", func() {
	// 				root.StopAudio()
	// 			}),
	// 		),
	// 		container.NewHScroll(synth.controls.UI()),
	// 	),
	// )

	// w.ShowAndRun()
}
