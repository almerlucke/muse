package classic

import (
	"github.com/almerlucke/muse"
	// shaping "github.com/almerlucke/muse/components/waveshaping"
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/filters/korg35"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/mixer"
	"github.com/almerlucke/muse/modules/noise"
	"github.com/almerlucke/muse/modules/osc"
	"github.com/almerlucke/muse/modules/pan"
	"github.com/almerlucke/muse/modules/polyphony"
)

type Voice struct {
	*muse.BasePatch
	ampEnv         *adsr.ADSR
	filterEnv      *adsr.ADSR
	Osc1           *osc.Osc
	Osc2           *osc.Osc
	noiseGen       *noise.Noise
	SourceMixer    *mixer.Mixer
	filter         *korg35.Korg35LPF
	panner         *pan.Pan
	ampEnvSteps    adsrc.StepProvider
	filterEnvSteps adsrc.StepProvider
	osc2Tuning     float64
	filterFcMin    float64
	filterFcMax    float64
}

func NewVoice(config *muse.Configuration, ampEnvSteps adsrc.StepProvider, filterEnvSteps adsrc.StepProvider) *Voice {
	osc1Mix := 0.6
	osc2Mix := 0.35
	noiseMix := 0.05

	voice := &Voice{
		BasePatch:      muse.NewPatch(0, 2, config),
		ampEnv:         adsr.New(ampEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config),
		filterEnv:      adsr.New(filterEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config),
		Osc1:           osc.New(100.0, 0.0, config),
		Osc2:           osc.New(100.0, 0.5, config),
		noiseGen:       noise.New(1, config),
		SourceMixer:    mixer.New(3, config),
		filter:         korg35.New(1500.0, 0.7, 2.0, config),
		panner:         pan.New(0.5, config),
		ampEnvSteps:    ampEnvSteps,
		filterEnvSteps: filterEnvSteps,
		osc2Tuning:     2.03,
		filterFcMin:    50.0,
		filterFcMax:    8000,
	}

	voice.SetSelf(voice)

	voice.SourceMixer.SetMix([]float64{osc1Mix, osc2Mix, noiseMix})

	voice.AddModule(voice.ampEnv)
	voice.AddModule(voice.filterEnv)
	voice.AddModule(voice.Osc1)
	voice.AddModule(voice.Osc2)
	voice.AddModule(voice.noiseGen)
	voice.AddModule(voice.SourceMixer)
	voice.AddModule(voice.filter)
	voice.AddModule(voice.panner)

	filterScaler := voice.AddModule(functor.New(1, func(v []float64) float64 {
		min := voice.filterFcMin
		max := voice.filterFcMax
		if min > max {
			tmp := max
			max = min
			min = tmp
		}
		return v[0]*(max-min) + min
	}, config))

	ampVCA := voice.AddModule(functor.NewMult(2, config))

	voice.Osc1.Connect(4, voice.SourceMixer, 0)
	voice.Osc2.Connect(4, voice.SourceMixer, 1)
	voice.noiseGen.Connect(0, voice.SourceMixer, 2)
	voice.SourceMixer.Connect(0, voice.filter, 0)
	voice.filterEnv.Connect(0, filterScaler, 0)
	filterScaler.Connect(0, voice.filter, 1)
	voice.filter.Connect(0, ampVCA, 0)
	voice.ampEnv.Connect(0, ampVCA, 1)
	ampVCA.Connect(0, voice.panner, 0)
	voice.panner.Connect(0, voice, 0)
	voice.panner.Connect(1, voice, 1)

	return voice
}

func New(numVoices int, ampEnv adsrc.StepProvider, filterEnv adsrc.StepProvider, config *muse.Configuration) *polyphony.Polyphony {
	voices := make([]polyphony.Voice, numVoices)

	for i := 0; i < numVoices; i++ {
		voices[i] = NewVoice(config, ampEnv, filterEnv)
	}

	return polyphony.New(2, voices, config)
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

func (v *Voice) NoteOn(amplitude float64, msg any, config *muse.Configuration) {
	content := msg.(map[string]any)

	v.handleMessage(content)

	if fcRaw, ok := content["frequency"]; ok {
		fc := fcRaw.(float64)
		v.Osc1.SetFrequency(fc)
		v.Osc2.SetFrequency(fc * v.osc2Tuning)
		v.ampEnv.TriggerFull(0, amplitude, v.ampEnvSteps.GetSteps(), adsrc.Absolute, adsrc.NoteOff)
		v.filterEnv.TriggerFull(0, 1.0, v.filterEnvSteps.GetSteps(), adsrc.Absolute, adsrc.NoteOff)
	}
}

func (v *Voice) NoteOff() {
	v.ampEnv.Release()
	v.filterEnv.Release()
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

func (v *Voice) Osc1PulseWidth() float64 {
	return v.Osc1.PulseWidth()
}

func (v *Voice) SetOsc1PulseWidth(pw float64) {
	v.Osc1.SetPulseWidth(pw)
}

func (v *Voice) Osc2PulseWidth() float64 {
	return v.Osc2.PulseWidth()
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

func (v *Voice) SetPan(pan float64) {
	v.panner.SetPan(pan)
}

func (v *Voice) SetFilterFcMin(min float64) {
	v.filterFcMin = min
}

func (v *Voice) SetFilterFcMax(max float64) {
	v.filterFcMax = max
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

	if pan, ok := content["pan"]; ok {
		v.SetPan(pan.(float64))
	}

	if filterFcMin, ok := content["filterFcMin"]; ok {
		v.SetFilterFcMin(filterFcMin.(float64))
	}

	if filterFcMax, ok := content["filterFcMax"]; ok {
		v.SetFilterFcMax(filterFcMax.(float64))
	}
}

func (v *Voice) ReceiveMessage(msg any) []*muse.Message {
	v.handleMessage(msg.(map[string]any))
	return nil
}
