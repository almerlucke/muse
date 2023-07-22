package adsr

import (
	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
)

type ADSR struct {
	*muse.BaseModule
	adsr     *adsrc.ADSR
	maxLevel float64
	duration float64
}

func New(steps []adsrc.Step, durationMode adsrc.DurationMode, releaseMode adsrc.ReleaseMode, maxLevel float64, config *muse.Configuration) *ADSR {
	a := &ADSR{
		BaseModule: muse.NewBaseModule(0, 1, config, ""),
	}

	a.maxLevel = maxLevel
	a.duration = 250.0
	a.adsr = &adsrc.ADSR{}
	a.adsr.Initialize(steps, durationMode, releaseMode, config.SampleRate)

	a.SetSelf(a)

	return a
}

func (a *ADSR) SetDuration(duration float64) {
	a.duration = duration
}

func (a *ADSR) Bang() {
	switch a.adsr.ReleaseMode() {
	case adsrc.Duration:
		a.adsr.TriggerWithDuration(a.duration, a.maxLevel)
	case adsrc.Automatic:
		a.adsr.Trigger(a.maxLevel)
	}
}

func (a *ADSR) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Bang
		if value == muse.Bang {
			a.Bang()
		}
	case 1: // MaxLevel
		a.maxLevel = value.(float64)
	case 2: // Duration
		a.duration = value.(float64)
	}
}

func (a *ADSR) ReceiveMessage(msg any) []*muse.Message {
	if content, ok := msg.(map[string]any); ok {
		if duration, ok := content["duration"]; ok {
			a.duration = duration.(float64)
		}

		if maxLevel, ok := content["maxLevel"]; ok {
			a.maxLevel = maxLevel.(float64)
		}
	}

	if muse.IsBang(msg) {
		a.Bang()
	}

	return nil
}

func (a *ADSR) TriggerFull(duration float64, maxLevel float64, steps []adsrc.Step, durationMode adsrc.DurationMode, releaseMode adsrc.ReleaseMode) {
	a.adsr.TriggerFull(duration, maxLevel, steps, durationMode, releaseMode)
}

func (a *ADSR) TriggerWithDuration(duration float64, maxLevel float64) {
	a.adsr.TriggerWithDuration(duration, maxLevel)
}

func (a *ADSR) IsActive() bool {
	return !a.adsr.IsFinished()
}

func (a *ADSR) Release() {
	a.adsr.Release()
}

func (a *ADSR) Synthesize() bool {
	if !a.BaseModule.Synthesize() {
		return false
	}

	out := a.Outputs[0].Buffer

	for i := 0; i < a.Config.BufferSize; i++ {
		out[i] = a.adsr.Tick()[0]
	}

	return true
}
