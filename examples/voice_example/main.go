package main

import (
	"github.com/almerlucke/muse"
	"github.com/mkb218/gosndfile/sndfile"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shapingc "github.com/almerlucke/muse/components/shaping"
	"github.com/almerlucke/muse/messengers"
	"github.com/almerlucke/muse/messengers/generators/sequencer"
	"github.com/almerlucke/muse/messengers/triggers/stepper"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/moog/moog1"
	"github.com/almerlucke/muse/modules/freeverb"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/shaper"
	"github.com/almerlucke/muse/modules/vartri"
)

type TestVoice struct {
	*muse.BasePatch
	adsrEnv *adsr.ADSR
	Shaper  *shaper.Shaper
	Filter  *moog1.Moog1
	phasor  *phasor.Phasor
}

func paramMapper(param int, value float64, shaper shapingc.Shaper) {
	if param == 0 {
		shaper.(*shapingc.Chain).SetSuperSawM1(value)
	}
}

func NewTestVoice(config *muse.Configuration) *TestVoice {
	testVoice := &TestVoice{
		BasePatch: muse.NewPatch(0, 1, config, ""),
	}

	steps := []adsrc.Step{
		{Level: 1.0, Duration: 20, Shape: 0.0},
		{Level: 0.3, Duration: 20, Shape: 0.0},
		{Duration: 20},
		{Duration: 500, Shape: 0.0},
	}

	adsrEnv := testVoice.AddModule(adsr.NewADSR(steps, adsrc.Absolute, adsrc.Duration, 1.0, config, "adsr"))
	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config, ""))
	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
	shape := testVoice.AddModule(shaper.NewShaper(shapingc.NewSuperSaw(), 1, paramMapper, nil, config, "shaper"))
	filter := testVoice.AddModule(moog1.NewMoog1(1700.0, 0.48, 1.75, config, "filter"))

	muse.Connect(osc, 0, shape, 0)
	muse.Connect(shape, 0, multiplier, 0)
	muse.Connect(adsrEnv, 0, multiplier, 1)
	muse.Connect(multiplier, 0, filter, 0)
	muse.Connect(filter, 0, testVoice, 0)

	testVoice.adsrEnv = adsrEnv.(*adsr.ADSR)
	testVoice.Shaper = shape.(*shaper.Shaper)
	testVoice.Filter = filter.(*moog1.Moog1)
	testVoice.phasor = osc.(*phasor.Phasor)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.adsrEnv.IsActive()
}

func (tv *TestVoice) Activate(duration float64, amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.adsrEnv.TriggerWithDuration(duration, amplitude)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func NewVoiceMessage(address string, duration float64, amplitude float64, frequency float64) []*muse.Message {
	return []*muse.Message{muse.NewMessage(
		address,
		map[string]any{
			"duration":  duration,
			"amplitude": amplitude,
			"message":   map[string]any{"osc": map[string]any{"frequency": frequency}},
		})}
}

func main() {
	env := muse.NewEnvironment(2, 3*44100, 512)

	sequencer := sequencer.NewSequencer([][]*muse.Message{
		NewVoiceMessage("voicePlayer", 250, 0.4, 200.0),
		NewVoiceMessage("voicePlayer", 125, 1.0, 400.0),
		NewVoiceMessage("voicePlayer", 500, 0.6, 500.0),
		NewVoiceMessage("voicePlayer", 125, 1.0, 600.0),
		NewVoiceMessage("voicePlayer", 250, 0.5, 100.0),
		NewVoiceMessage("voicePlayer", 250, 0.5, 50.0),
		NewVoiceMessage("voicePlayer", 750, 1.0, 50.0),
		NewVoiceMessage("voicePlayer", 500, 0.3, 100.0),
		NewVoiceMessage("voicePlayer", 375, 1.0, 250.0),
		NewVoiceMessage("voicePlayer", 250, 0.7, 750.0),
		NewVoiceMessage("voicePlayer", 250, 1.0, 375.0),
	})

	env.AddMessenger(messengers.NewGenerator(sequencer, "sequencer1"))

	env.AddMessenger(stepper.NewStepper(
		stepper.NewSliceProvider([]float64{125, -250, 125, 250, 125, -125, 250, -125, 125, -250}),
		[]string{"sequencer1"}, "",
	))

	paramVarTri1 := env.AddModule(vartri.NewVarTri(0.155, 0.0, 0.5, env.Config, "vartri1"))
	paramVarTri2 := env.AddModule(vartri.NewVarTri(0.325, 0.0, 0.5, env.Config, "vartri2"))
	superSawDrive := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*0.82 + 0.15 }, env.Config, ""))
	filterCutOff := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*2200.0 + 40.0 }, env.Config, ""))

	voices := []muse.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config)
		voices = append(voices, voice)
		// connect lfo to voices
		muse.Connect(superSawDrive, 0, voice.Shaper, 1)
		muse.Connect(filterCutOff, 0, voice.Filter, 1)
	}

	voicePlayer := env.AddModule(muse.NewVoicePlayer(1, voices, env.Config, "voicePlayer"))
	allpass := env.AddModule(allpass.NewAllpass(375.0, 375.0, 0.3, env.Config, "allpass"))
	allpassAmp := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0] * 0.8 }, env.Config, ""))
	reverb := env.AddModule(freeverb.NewFreeVerb(env.Config, "freeverb"))

	reverb.(*freeverb.FreeVerb).SetDamp(0.01)
	reverb.(*freeverb.FreeVerb).SetDry(0.7)
	reverb.(*freeverb.FreeVerb).SetWet(0.2)
	reverb.(*freeverb.FreeVerb).SetRoomSize(0.8)
	reverb.(*freeverb.FreeVerb).SetWidth(0.8)

	// connect external voice inputs to voice player so the external modules
	// are always synthesized even if no voice is active at the moment
	muse.Connect(superSawDrive, 0, voicePlayer, 0)
	muse.Connect(filterCutOff, 0, voicePlayer, 0)

	muse.Connect(paramVarTri1, 0, superSawDrive, 0)
	muse.Connect(paramVarTri2, 0, filterCutOff, 0)
	muse.Connect(voicePlayer, 0, allpass, 0)
	muse.Connect(allpass, 0, allpassAmp, 0)
	muse.Connect(voicePlayer, 0, reverb, 0)
	muse.Connect(allpassAmp, 0, reverb, 1)
	muse.Connect(reverb, 0, env, 0)
	muse.Connect(reverb, 1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/voices.aiff", 24.0, env.Config.SampleRate, sndfile.SF_FORMAT_AIFF)
}
