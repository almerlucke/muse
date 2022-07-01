package adsr

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
)

type ADSRModule struct {
	*muse.BaseModule
	adsr     *adsrc.ADSR
	maxLevel float64
}

func NewADSRModule(steps []adsrc.Step, durationMode adsrc.DurationMode, releaseMode adsrc.ReleaseMode, maxLevel float64, config *muse.Configuration, identifier string) *ADSRModule {
	am := &ADSRModule{
		BaseModule: muse.NewBaseModule(0, 1, config, identifier),
	}

	am.maxLevel = maxLevel
	am.adsr = &adsrc.ADSR{}
	am.adsr.Initialize(steps, durationMode, releaseMode, config.SampleRate)

	return am
}

func (a *ADSRModule) ReceiveMessage(msg any) []*muse.Message {
	if messengers.IsBang(msg) {
		a.adsr.Trigger(a.maxLevel)
	}

	return nil
}

func (a *ADSRModule) Synthesize() bool {
	if !a.BaseModule.Synthesize() {
		return false
	}

	out := a.Outputs[0].Buffer

	for i := 0; i < a.Config.BufferSize; i++ {
		out[i] = a.adsr.Synthesize()
	}

	return true
}
