package adsr

import (
	"github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/muse"
)

type ADSR struct {
	*muse.BaseModule
	adsr     *adsr.ADSR
	level    float64
	duration float64
}

func New(setting *adsr.Setting, releaseMode adsr.ReleaseMode, level float64) *ADSR {
	a := &ADSR{
		BaseModule: muse.NewBaseModule(0, 1),
	}

	a.level = level
	a.duration = 250.0
	a.adsr = adsr.New(setting, releaseMode, muse.SampleRate())

	a.SetSelf(a)

	return a
}

func (a *ADSR) SetDuration(duration float64) {
	a.duration = duration
}

func (a *ADSR) Bang() {
	switch a.adsr.ReleaseMode() {
	case adsr.Duration:
		a.adsr.TriggerWithDuration(a.duration, a.level)
	case adsr.Automatic:
		a.adsr.Trigger(a.level)
	default:
		break
	}
}

func (a *ADSR) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Bang
		if value == muse.Bang {
			a.Bang()
		}
	case 1: // MaxLevel
		a.level = value.(float64)
	case 2: // Duration
		a.duration = value.(float64)
	}
}

func (a *ADSR) ReceiveMessage(msg any) []*muse.Message {
	if content, ok := msg.(map[string]any); ok {
		if duration, ok := content["duration"]; ok {
			a.duration = duration.(float64)
		}

		if level, ok := content["level"]; ok {
			a.level = level.(float64)
		}
	}

	if muse.IsBang(msg) {
		a.Bang()
	}

	return nil
}

func (a *ADSR) Trigger(level float64) {
	a.adsr.Trigger(level)
}

func (a *ADSR) TriggerFull(duration float64, level float64, setting *adsr.Setting, releaseMode adsr.ReleaseMode) {
	a.adsr.TriggerFull(duration, level, setting, releaseMode)
}

func (a *ADSR) TriggerWithDuration(duration float64, maxLevel float64) {
	a.adsr.TriggerWithDuration(duration, maxLevel)
}

func (a *ADSR) IsActive() bool {
	return !a.adsr.Done()
}

func (a *ADSR) Release() {
	a.adsr.Release()
}

func (a *ADSR) Clear() {
	a.adsr.Clear()
}

func (a *ADSR) Synthesize() bool {
	if !a.BaseModule.Synthesize() {
		return false
	}

	out := a.Outputs[0].Buffer

	for i := 0; i < a.Config.BufferSize; i++ {
		out[i] = a.adsr.Generate()
	}

	return true
}
