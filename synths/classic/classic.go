package classic

import (
	"github.com/almerlucke/muse"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/filters/moog"
)

type Voice struct {
	*muse.BasePatch
	ampEnv    *adsr.ADSR
	filterEnv *adsr.ADSR
	osc1      *blosc.Osc
	filter    *moog.Moog
	shaper    shaping.Shaper
	// ampStepProvider adsrc.ADSRStepProvider
}

// func NewTestVoice(config *muse.Configuration, ampStepProvider adsrc.ADSRStepProvider) *TestVoice {
// 	testVoice := &TestVoice{
// 		BasePatch:       muse.NewPatch(0, 1, config, ""),
// 		ampStepProvider: ampStepProvider,
// 		shaper:          shaping.NewJP8000triMod(),
// 	}

// 	ampEnv := testVoice.AddModule(adsr.NewADSR(ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "ampAdsr"))
// 	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config, ""))
// 	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
// 	shape := testVoice.AddModule(waveshaper.NewWaveShaper(testVoice.shaper, 0, nil, nil, config, "shaper"))

// 	muse.Connect(osc, 0, shape, 0)
// 	muse.Connect(shape, 0, multiplier, 0)
// 	muse.Connect(ampEnv, 0, multiplier, 1)
// 	muse.Connect(multiplier, 0, testVoice, 0)

// 	testVoice.ampEnv = ampEnv.(*adsr.ADSR)
// 	testVoice.phasor = osc.(*phasor.Phasor)

// 	return testVoice
// }

// func (tv *TestVoice) IsActive() bool {
// 	return tv.ampEnv.IsActive()
// }

// func (tv *TestVoice) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
// 	msg := message.(map[string]any)

// 	tv.ampEnv.TriggerFull(duration, amplitude, tv.ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration)
// 	tv.phasor.ReceiveMessage(msg["osc"])
// }

// func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
// 	// STUB
// }

// func (tv *TestVoice) NoteOff() {
// 	tv.ampEnv.Release()
// 	tv.filterEnv.Release()
// }

// func (tv *TestVoice) ReceiveMessage(msg any) []*muse.Message {
// 	// content := msg.(map[string]any)
// 	return nil
// }
