package wtosc

import (
	"github.com/almerlucke/muse"
	wtoscc "github.com/almerlucke/muse/components/wtosc"
	"github.com/almerlucke/muse/io"
)

// Osc wavetable osc
type Osc struct {
	*muse.BaseModule
	component *wtoscc.Osc
}

func New(sf *io.WaveTableSoundFile, fc float64, phase float64, tableIndex float64, amp float64) *Osc {
	osc := &Osc{
		BaseModule: muse.NewBaseModule(3, 1),
		component:  wtoscc.New(sf, fc, muse.SampleRate(), phase, tableIndex, amp),
	}

	osc.SetSelf(osc)

	return osc
}

func (osc *Osc) Amplitude() float64 {
	return osc.component.Amplitude()
}

func (osc *Osc) SetAmplitude(amp float64) {
	osc.component.SetAmplitude(amp)
}

func (osc *Osc) Phase() float64 {
	return osc.component.Phase()
}

func (osc *Osc) SetPhase(phase float64) {
	osc.component.SetPhase(phase)
}

func (osc *Osc) Frequency() float64 {
	return osc.component.Frequency()
}

func (osc *Osc) SetFrequency(fc float64) {
	osc.component.SetFrequency(fc)
}

func (osc *Osc) TableIndex() float64 {
	return osc.component.TableIndex()
}

func (osc *Osc) SetTableIndex(tableIndex float64) {
	osc.component.SetTableIndex(tableIndex)
}

func (osc *Osc) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Frequency
		osc.SetFrequency(value.(float64))
	case 1: // Phase
		osc.SetPhase(value.(float64))
	case 2: // Table Index
		osc.SetTableIndex(value.(float64))
	case 3: // Amplitude
		osc.SetAmplitude(value.(float64))
	}
}

func (osc *Osc) ReceiveMessage(msg any) []*muse.Message {
	if params, ok := msg.(map[string]any); ok {
		if f, ok := params["frequency"]; ok {
			osc.SetFrequency(f.(float64))
		}

		if ph, ok := params["phase"]; ok {
			osc.SetPhase(ph.(float64))
		}

		if ti, ok := params["tableIndex"]; ok {
			osc.SetTableIndex(ti.(float64))
		}

		if amp, ok := params["amplitude"]; ok {
			osc.SetAmplitude(amp.(float64))
		}
	}

	return nil
}

func (osc *Osc) Synthesize() bool {
	if !osc.BaseModule.Synthesize() {
		return false
	}

	freqInput := osc.InputAtIndex(0)
	phaseOffsetInput := osc.InputAtIndex(1)
	tableIndexInput := osc.InputAtIndex(2)

	out := osc.OutputAtIndex(0).Buffer

	for i := 0; i < osc.Config.BufferSize; i++ {
		if freqInput.IsConnected() {
			osc.component.SetFrequency(freqInput.Buffer[i])
		}

		if phaseOffsetInput.IsConnected() {
			osc.component.OffsetPhase(phaseOffsetInput.Buffer[i])
		}

		if tableIndexInput.IsConnected() {
			osc.component.SetTableIndex(tableIndexInput.Buffer[i])
		}

		out[i] = osc.component.Generate()
	}

	return true
}
