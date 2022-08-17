package adsr

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
)

type ADSR struct {
	*muse.BaseModule
	adsr     *adsrc.ADSR
	maxLevel float64
}

func NewADSR(steps []adsrc.Step, durationMode adsrc.DurationMode, releaseMode adsrc.ReleaseMode, maxLevel float64, config *muse.Configuration, identifier string) *ADSR {
	am := &ADSR{
		BaseModule: muse.NewBaseModule(0, 1, config, identifier),
	}

	am.maxLevel = maxLevel
	am.adsr = &adsrc.ADSR{}
	am.adsr.Initialize(steps, durationMode, releaseMode, config.SampleRate)

	return am
}

func (a *ADSR) ReceiveMessage(msg any) []*muse.Message {
	if messengers.IsBang(msg) {
		a.adsr.Trigger(a.maxLevel)
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

func (a *ADSR) Synthesize() bool {
	if !a.BaseModule.Synthesize() {
		return false
	}

	out := a.Outputs[0].Buffer

	for i := 0; i < a.Config.BufferSize; i++ {
		out[i] = a.adsr.Synthesize()
	}

	return true
}
