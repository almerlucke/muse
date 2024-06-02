package classic

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/filters"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/utils"

	// shaping "github.com/almerlucke/muse/components/waveshaping"
	adsrc "github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/mixer"
	"github.com/almerlucke/muse/modules/noise"
	"github.com/almerlucke/muse/modules/osc"
	"github.com/almerlucke/muse/modules/pan"
)

type Voice struct {
	*muse.BasePatch
	ampEnv           *adsr.ADSR
	filterEnv        *adsr.ADSR
	Osc1             *osc.Osc
	Osc2             *osc.Osc
	noiseGen         *noise.Noise
	SourceMixer      *mixer.Mixer
	filter           filters.Filter
	panner           *pan.Pan
	ampEnvSetting    *adsrc.Setting
	filterEnvSetting *adsrc.Setting
	osc2Tuning       float64
	filterFcMin      float64
	filterFcMax      float64
}

func NewVoice(ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting, filter filters.Filter) *Voice {
	osc1Mix := 0.6
	osc2Mix := 0.35
	noiseMix := 0.05

	voice := &Voice{
		BasePatch:        muse.NewPatch(0, 2),
		ampEnv:           adsr.New(ampEnvSetting, adsrc.Duration, 1.0),
		filterEnv:        adsr.New(filterEnvSetting, adsrc.Duration, 1.0),
		Osc1:             osc.New(100.0, 0.0),
		Osc2:             osc.New(100.0, 0.5),
		noiseGen:         noise.New(1),
		SourceMixer:      mixer.New(3),
		filter:           filter,
		panner:           pan.New(0.5),
		ampEnvSetting:    ampEnvSetting,
		filterEnvSetting: filterEnvSetting,
		osc2Tuning:       2.03,
		filterFcMin:      50.0,
		filterFcMax:      8000,
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
		minFc := voice.filterFcMin
		maxFc := voice.filterFcMax
		if minFc > maxFc {
			tmp := maxFc
			maxFc = minFc
			minFc = tmp
		}
		return v[0]*(maxFc-minFc) + minFc
	}))

	ampVCA := voice.AddModule(functor.NewMult(2))

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

	v.ampEnv.TriggerFull(duration, amplitude, v.ampEnvSetting, adsrc.Duration)
	v.filterEnv.TriggerFull(duration, 1.0, v.filterEnvSetting, adsrc.Duration)
}

func (v *Voice) NoteOn(amplitude float64, msg any, config *muse.Configuration) {
	content := msg.(map[string]any)

	v.handleMessage(content)

	if fcRaw, ok := content["frequency"]; ok {
		fc := fcRaw.(float64)
		v.Osc1.SetFrequency(fc)
		v.Osc2.SetFrequency(fc * v.osc2Tuning)
		v.ampEnv.TriggerFull(0, amplitude, v.ampEnvSetting, adsrc.NoteOff)
		v.filterEnv.TriggerFull(0, 1.0, v.filterEnvSetting, adsrc.NoteOff)
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

	if p, ok := content["pan"]; ok {
		v.SetPan(p.(float64))
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

type Synth struct {
	*polyphony.Polyphony
}

type Setting struct {
	Osc1Mix         float64
	Osc2Mix         float64
	NoiseMix        float64
	Osc1PulseWidth  float64
	Osc2PulseWidth  float64
	FilterResonance float64
	Osc1SineMix     float64
	Osc1TriMix      float64
	Osc1SawMix      float64
	Osc1PulseMix    float64
	Osc2SineMix     float64
	Osc2TriMix      float64
	Osc2SawMix      float64
	Osc2PulseMix    float64
	Osc2Tuning      float64
	Pan             float64
	FilterFcMin     float64
	FilterFcMax     float64
}

func DefaultSetting() Setting {
	return Setting{
		Osc1Mix:         0.5,
		Osc2Mix:         0.5,
		NoiseMix:        0.0,
		Osc1PulseWidth:  0.4,
		Osc2PulseWidth:  0.3,
		FilterResonance: 0.6,
		Osc1SineMix:     0.0,
		Osc1TriMix:      0.0,
		Osc1PulseMix:    1.0,
		Osc1SawMix:      0.0,
		Osc2SineMix:     0.0,
		Osc2TriMix:      0.0,
		Osc2PulseMix:    1.0,
		Osc2SawMix:      0.0,
		Osc2Tuning:      2.05,
		Pan:             0.5,
		FilterFcMin:     50.0,
		FilterFcMax:     8000.0,
	}
}

func New(numVoices int, ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting, filterFactory utils.Factory[filters.Filter], filterConfig *filters.FilterConfig) *Synth {
	voices := make([]polyphony.Voice, numVoices)

	for i := 0; i < numVoices; i++ {
		voices[i] = NewVoice(ampEnvSetting, filterEnvSetting, filterFactory.New(filterConfig))
	}

	s := &Synth{
		Polyphony: polyphony.New(2, voices),
	}

	s.SetSelf(s)

	return s
}

func (s *Synth) Set(setting Setting) {
	s.SetOsc1Mix(setting.Osc1Mix)
	s.SetOsc2Mix(setting.Osc2Mix)
	s.SetNoiseMix(setting.NoiseMix)
	s.SetOsc1PulseWidth(setting.Osc1PulseWidth)
	s.SetOsc2PulseWidth(setting.Osc2PulseWidth)
	s.SetFilterResonance(setting.FilterResonance)
	s.SetOsc1SineMix(setting.Osc1SineMix)
	s.SetOsc1SawMix(setting.Osc1SawMix)
	s.SetOsc1TriMix(setting.Osc1TriMix)
	s.SetOsc1PulseMix(setting.Osc1PulseMix)
	s.SetOsc2SineMix(setting.Osc2SineMix)
	s.SetOsc2SawMix(setting.Osc2SawMix)
	s.SetOsc2TriMix(setting.Osc2TriMix)
	s.SetOsc2PulseMix(setting.Osc2PulseMix)
	s.SetOsc2Tuning(setting.Osc2Tuning)
	s.SetPan(setting.Pan)
	s.SetFilterFcMin(setting.FilterFcMin)
	s.SetFilterFcMax(setting.FilterFcMax)
}

func (s *Synth) SetOsc1Mix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc1Mix(mix)
	})
}

func (s *Synth) SetOsc2Mix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc2Mix(mix)
	})
}

func (s *Synth) SetNoiseMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetNoiseMix(mix)
	})
}

func (s *Synth) SetOsc1PulseWidth(pw float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc1PulseWidth(pw)
	})
}

func (s *Synth) SetOsc2PulseWidth(pw float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc2PulseWidth(pw)
	})
}

func (s *Synth) SetFilterResonance(res float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetFilterResonance(res)
	})
}

func (s *Synth) SetOsc1SineMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc1SineMix(mix)
	})
}

func (s *Synth) SetOsc1SawMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc1SawMix(mix)
	})
}

func (s *Synth) SetOsc1PulseMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc1PulseMix(mix)
	})
}

func (s *Synth) SetOsc1TriMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc1TriMix(mix)
	})
}

func (s *Synth) SetOsc2SineMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc2SineMix(mix)
	})
}

func (s *Synth) SetOsc2SawMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc2SawMix(mix)
	})
}

func (s *Synth) SetOsc2PulseMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc2PulseMix(mix)
	})
}

func (s *Synth) SetOsc2TriMix(mix float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc2TriMix(mix)
	})
}

func (s *Synth) SetOsc2Tuning(tuning float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetOsc2Tuning(tuning)
	})
}

func (s *Synth) SetPan(pan float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetPan(pan)
	})
}

func (s *Synth) SetFilterFcMin(min float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetFilterFcMin(min)
	})
}

func (s *Synth) SetFilterFcMax(max float64) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetFilterFcMin(max)
	})
}
