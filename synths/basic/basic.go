package basic

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/filters"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/utils"
	// shaping "github.com/almerlucke/muse/components/waveshaping"
	adsrc "github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/pan"
)

type Source interface {
	muse.Module
	Activate(values map[string]any)
	SetValues(values map[string]any)
	SetValue(key string, value any)
}

type Voice struct {
	*muse.BasePatch
	ampEnv           *adsr.ADSR
	filterEnv        *adsr.ADSR
	filter           filters.Filter
	panner           *pan.Pan
	source           Source
	ampEnvSetting    *adsrc.Setting
	filterEnvSetting *adsrc.Setting
	filterFcMin      float64
	filterFcMax      float64
}

func NewVoice(source Source, ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting, filter filters.Filter) *Voice {
	voice := &Voice{
		BasePatch:        muse.NewPatch(0, 2),
		ampEnv:           adsr.New(ampEnvSetting, adsrc.Duration, 1.0),
		filterEnv:        adsr.New(filterEnvSetting, adsrc.Duration, 1.0),
		filter:           filter,
		panner:           pan.New(0.5),
		source:           source,
		ampEnvSetting:    ampEnvSetting,
		filterEnvSetting: filterEnvSetting,
		filterFcMin:      40.0,
		filterFcMax:      14000.0,
	}

	voice.SetSelf(voice)

	voice.AddModule(voice.ampEnv)
	voice.AddModule(voice.filterEnv)
	voice.AddModule(voice.filter)
	voice.AddModule(voice.source)
	voice.AddModule(voice.panner)

	filterScaler := functor.New(1, func(v []float64) float64 {
		minFc := voice.filterFcMin
		maxFc := voice.filterFcMax
		if minFc > maxFc {
			tmp := maxFc
			maxFc = minFc
			minFc = tmp
		}
		return v[0]*(maxFc-minFc) + minFc
	}).AddTo(voice)

	ampVCA := functor.NewMult(2).AddTo(voice)

	voice.source.Connect(0, voice.filter, 0)
	voice.filterEnv.Connect(0, filterScaler, 0)
	filterScaler.Connect(0, voice.filter, 1)
	voice.filter.Connect(0, ampVCA, 0)
	voice.ampEnv.Connect(0, ampVCA, 1)
	ampVCA.Connect(0, voice.panner, 0)

	voice.In(voice.panner, voice.panner, 1)

	return voice
}

func (v *Voice) IsActive() bool {
	return v.ampEnv.IsActive()
}

func (v *Voice) Clear() {
	v.ampEnv.Clear()
	v.filterEnv.Clear()
}

func (v *Voice) Note(duration float64, amplitude float64, msg any, config *muse.Configuration) {
	if values, ok := msg.(map[string]any); ok {
		if attackDuration, ok := values["attackDuration"].(float64); ok {
			v.ampEnvSetting.AttackDuration = attackDuration
			v.filterEnvSetting.AttackDuration = attackDuration
		}
		if releaseDuration, ok := values["releaseDuration"].(float64); ok {
			v.ampEnvSetting.ReleaseDuration = releaseDuration
			v.filterEnvSetting.ReleaseDuration = releaseDuration
		}
		if panning, ok := values["pan"].(float64); ok {
			v.panner.SetPan(panning)
		}
		if filterFcMin, ok := values["filterFcMin"].(float64); ok {
			v.filterFcMin = filterFcMin
		}
		if filterFcMax, ok := values["filterFcMax"].(float64); ok {
			v.filterFcMax = filterFcMax
		}
		if filterRes, ok := values["filterResonance"].(float64); ok {
			v.filter.SetResonance(filterRes)
		}

		v.source.Activate(msg.(map[string]any))
	}

	v.ampEnv.TriggerFull(duration, amplitude, v.ampEnvSetting, adsrc.Duration)
	v.filterEnv.TriggerFull(duration, 1.0, v.filterEnvSetting, adsrc.Duration)
}

func (v *Voice) NoteOn(amplitude float64, msg any, config *muse.Configuration) {
	v.source.Activate(msg.(map[string]any))
	v.ampEnv.TriggerFull(0, amplitude, v.ampEnvSetting, adsrc.NoteOff)
	v.filterEnv.TriggerFull(0, 1.0, v.filterEnvSetting, adsrc.NoteOff)
}

func (v *Voice) NoteOff() {
	v.ampEnv.Release()
	v.filterEnv.Release()
}

func (v *Voice) SetValue(key string, value any) {
	v.source.SetValue(key, value)
}

func (v *Voice) SetValues(values map[string]any) {
	v.source.SetValues(values)
}

func (v *Voice) ReceiveMessage(msg any) []*muse.Message {
	if values, ok := msg.(map[string]any); ok {

		v.source.SetValues(values)
	}

	return nil
}

type Synth struct {
	*polyphony.Polyphony
}

func New(numVoices int, sourceFactory utils.Factory[Source], sourceConfig any, ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting, filterFactory utils.Factory[filters.Filter], filterConfig *filters.FilterConfig) *Synth {
	voices := make([]polyphony.Voice, numVoices)

	for i := 0; i < numVoices; i++ {
		voices[i] = NewVoice(sourceFactory.New(sourceConfig), ampEnvSetting, filterEnvSetting, filterFactory.New(filterConfig))
	}

	s := &Synth{
		Polyphony: polyphony.New(2, voices),
	}

	s.SetSelf(s)

	return s
}

func (s *Synth) SetValue(key string, value any) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetValue(key, value)
	})
}

func (s *Synth) SetValues(values map[string]any) {
	s.CallVoices(func(v polyphony.Voice) {
		v.(*Voice).SetValues(values)
	})
}
