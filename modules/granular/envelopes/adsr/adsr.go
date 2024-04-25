package adsr

import (
	"github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/granular"
)

type Envelope struct {
	adsr *adsr.ADSR
}

type Parameter interface {
	ADSRSetting() *adsr.Setting
}

func (e *Envelope) New(cfg any) granular.Envelope {
	return &Envelope{
		adsr: adsr.New(&adsr.Setting{}, adsr.Automatic, cfg.(*muse.Configuration).SampleRate),
	}
}

func (e *Envelope) Activate(amplitude float64, durationSamples int64, param granular.Parameter, _ *muse.Configuration) {
	var (
		inputSetting        = param.(any).(Parameter).ADSRSetting()
		outputSetting       = *inputSetting
		envRelativeDuration float64
		grainDuration       = param.Duration()
		scale               = 1.0
	)

	envRelativeDuration += inputSetting.AttackDuration

	if !inputSetting.SkipDecay {
		envRelativeDuration += inputSetting.DecayDuration
	}

	if !inputSetting.SkipSustain {
		envRelativeDuration += inputSetting.SustainDuration
	}

	envRelativeDuration += inputSetting.ReleaseDuration

	scale = 1.0 / envRelativeDuration

	outputSetting.AttackDuration = grainDuration * scale * inputSetting.AttackDuration

	if !inputSetting.SkipDecay {
		outputSetting.DecayDuration = grainDuration * scale * inputSetting.DecayDuration
	}

	if !inputSetting.SkipSustain {
		outputSetting.SustainDuration = grainDuration * scale * inputSetting.SustainDuration
	}

	outputSetting.ReleaseDuration = grainDuration * scale * inputSetting.ReleaseDuration

	e.adsr.TriggerFull(grainDuration, amplitude, &outputSetting, adsr.Automatic)
}

func (e *Envelope) Synthesize(buf []float64, bufSize int) {
	for i := 0; i < bufSize; i++ {
		buf[i] = e.adsr.Generate()
	}
}
