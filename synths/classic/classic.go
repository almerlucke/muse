package classic

import (
	"github.com/almerlucke/muse"
	// shaping "github.com/almerlucke/muse/components/waveshaping"
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/mixer"
	"github.com/almerlucke/muse/modules/noise"
)

type Voice struct {
	*muse.BasePatch
	ampEnv         *adsr.ADSR
	filterEnv      *adsr.ADSR
	Osc1           *blosc.Osc
	Osc2           *blosc.Osc
	noiseGen       *noise.Noise
	SourceMixer    *mixer.Mixer
	filter         *moog.Moog
	ampEnvSteps    adsrc.StepProvider
	filterEnvSteps adsrc.StepProvider
	osc2Tuning     float64
}

func NewVoice(config *muse.Configuration, ampEnvSteps adsrc.StepProvider, filterEnvSteps adsrc.StepProvider) *Voice {
	osc1Mix := 0.6
	osc2Mix := 0.35
	noiseMix := 0.05

	voice := &Voice{
		BasePatch:      muse.NewPatch(0, 1, config, ""),
		ampEnv:         adsr.NewADSR(ampEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "ampEnv"),
		filterEnv:      adsr.NewADSR(filterEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "filterEnv"),
		Osc1:           blosc.NewOsc(100.0, 0.0, config, "osc1"),
		Osc2:           blosc.NewOsc(100.0, 0.5, config, "osc2"),
		noiseGen:       noise.NewNoise(1, config),
		SourceMixer:    mixer.NewMixer(3, config, "sourceMixer"),
		filter:         moog.NewMoog(1500.0, 0.7, 1.0, config, "filter"),
		ampEnvSteps:    ampEnvSteps,
		filterEnvSteps: filterEnvSteps,
		osc2Tuning:     2.03,
	}

	voice.SourceMixer.SetMix([]float64{osc1Mix, osc2Mix, noiseMix})

	voice.AddModule(voice.ampEnv)
	voice.AddModule(voice.filterEnv)
	voice.AddModule(voice.Osc1)
	voice.AddModule(voice.Osc2)
	voice.AddModule(voice.noiseGen)
	voice.AddModule(voice.SourceMixer)
	voice.AddModule(voice.filter)

	filterScaler := voice.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0]*8000.0 + 50.0 }, config))
	ampVCA := voice.AddModule(functor.NewMult(2, config))

	muse.Connect(voice.Osc1, 4, voice.SourceMixer, 0)
	muse.Connect(voice.Osc2, 4, voice.SourceMixer, 1)
	muse.Connect(voice.noiseGen, 0, voice.SourceMixer, 2)
	muse.Connect(voice.SourceMixer, 0, voice.filter, 0)
	muse.Connect(voice.filterEnv, 0, filterScaler, 0)
	muse.Connect(filterScaler, 0, voice.filter, 1)
	muse.Connect(voice.filter, 0, ampVCA, 0)
	muse.Connect(voice.ampEnv, 0, ampVCA, 1)
	muse.Connect(ampVCA, 0, voice, 0)

	return voice
}

func (v *Voice) IsActive() bool {
	return v.ampEnv.IsActive()
}

func (v *Voice) Note(duration float64, amplitude float64, msg any, config *muse.Configuration) {
	content := msg.(map[string]any)

	v.handleMessage(content)

	if fcRaw, ok := content["frequency"]; ok {
		fc := fcRaw.(float64)
		v.Osc1.SetFrequency(fc)
		v.Osc2.SetFrequency(fc * v.osc2Tuning)
	}

	v.ampEnv.TriggerFull(duration, amplitude, v.ampEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration)
	v.filterEnv.TriggerFull(duration, 1.0, v.filterEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration)
}

func (v *Voice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	// STUB
}

func (v *Voice) NoteOff() {
	// tv.ampEnv.Release()
	// tv.filterEnv.Release()
}

func (v *Voice) SetOsc1Mix(mix float64) {
	v.SourceMixer.SetMixAt(0, mix)
}

func (v *Voice) SetOsc2Mix(mix float64) {
	v.SourceMixer.SetMixAt(1, mix)
}

func (v *Voice) SetNoiseMix(mix float64) {
	v.SourceMixer.SetMixAt(2, mix)
}

func (v *Voice) SetOsc1PulseWidth(pw float64) {
	v.Osc1.SetPulseWidth(pw)
}

func (v *Voice) SetOsc2PulseWidth(pw float64) {
	v.Osc2.SetPulseWidth(pw)
}

func (v *Voice) SetFilterResonance(res float64) {
	v.filter.SetResonance(res)
}

func (v *Voice) SetOsc1SineMix(mix float64) {
	v.Osc1.SetMixAt(0, mix)
}

func (v *Voice) SetOsc1SawMix(mix float64) {
	v.Osc1.SetMixAt(1, mix)
}

func (v *Voice) SetOsc1PulseMix(mix float64) {
	v.Osc1.SetMixAt(2, mix)
}

func (v *Voice) SetOsc1TriMix(mix float64) {
	v.Osc1.SetMixAt(3, mix)
}

func (v *Voice) SetOsc2SineMix(mix float64) {
	v.Osc2.SetMixAt(0, mix)
}

func (v *Voice) SetOsc2SawMix(mix float64) {
	v.Osc2.SetMixAt(1, mix)
}

func (v *Voice) SetOsc2PulseMix(mix float64) {
	v.Osc2.SetMixAt(2, mix)
}

func (v *Voice) SetOsc2TriMix(mix float64) {
	v.Osc2.SetMixAt(3, mix)
}

func (v *Voice) SetOsc2Tuning(tuning float64) {
	v.osc2Tuning = tuning
}

func (v *Voice) handleMessage(content map[string]any) {
	if osc1Mix, ok := content["osc1Mix"]; ok {
		v.SetOsc1Mix(osc1Mix.(float64))
	}

	if osc2Mix, ok := content["osc2Mix"]; ok {
		v.SetOsc2Mix(osc2Mix.(float64))
	}

	if noiseMix, ok := content["noiseMix"]; ok {
		v.SetNoiseMix(noiseMix.(float64))
	}

	if osc1PulseWidth, ok := content["osc1PulseWidth"]; ok {
		v.SetOsc1PulseWidth(osc1PulseWidth.(float64))
	}

	if osc2PulseWidth, ok := content["osc2PulseWidth"]; ok {
		v.SetOsc2PulseWidth(osc2PulseWidth.(float64))
	}

	if filterResonance, ok := content["filterResonance"]; ok {
		v.SetFilterResonance(filterResonance.(float64))
	}

	if osc1SineMix, ok := content["osc1SineMix"]; ok {
		v.SetOsc1SineMix(osc1SineMix.(float64))
	}

	if osc1SawMix, ok := content["osc1SawMix"]; ok {
		v.SetOsc1SawMix(osc1SawMix.(float64))
	}

	if osc1PulseMix, ok := content["osc1PulseMix"]; ok {
		v.SetOsc1PulseMix(osc1PulseMix.(float64))
	}

	if osc1TriMix, ok := content["osc1TriMix"]; ok {
		v.SetOsc1TriMix(osc1TriMix.(float64))
	}

	if osc2SineMix, ok := content["osc2SineMix"]; ok {
		v.SetOsc2SineMix(osc2SineMix.(float64))
	}

	if osc2SawMix, ok := content["osc2SawMix"]; ok {
		v.SetOsc2SawMix(osc2SawMix.(float64))
	}

	if osc2PulseMix, ok := content["osc2PulseMix"]; ok {
		v.SetOsc2PulseMix(osc2PulseMix.(float64))
	}

	if osc2TriMix, ok := content["osc2TriMix"]; ok {
		v.SetOsc2TriMix(osc2TriMix.(float64))
	}

	if osc2Tuning, ok := content["osc2Tuning"]; ok {
		v.SetOsc2Tuning(osc2Tuning.(float64))
	}
}

func (v *Voice) ReceiveMessage(msg any) []*muse.Message {
	v.handleMessage(msg.(map[string]any))
	return nil
}
