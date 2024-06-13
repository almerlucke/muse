package wtscan

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/wtscan"
	"github.com/almerlucke/muse/io"
)

// Scanner wavetable osc
type Scanner struct {
	*muse.BaseModule
	*wtscan.Scanner
}

func New(sf *io.WaveTableSoundFile, fc float64, phase float64, tableIndex float64, amp float64) *Scanner {
	sc := &Scanner{
		BaseModule: muse.NewBaseModule(3, 1),
		Scanner:    wtscan.New(sf, fc, muse.SampleRate(), phase, tableIndex, amp),
	}

	sc.SetSelf(sc)

	return sc
}

func (sc *Scanner) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Frequency
		sc.SetFrequency(value.(float64))
	case 1: // Phase
		sc.SetPhase(value.(float64))
	case 2: // Table Index
		sc.SetScanIndex(value.(float64))
	case 3: // Amplitude
		sc.SetAmplitude(value.(float64))
	}
}

func (sc *Scanner) ReceiveMessage(msg any) []*muse.Message {
	if params, ok := msg.(map[string]any); ok {
		if f, ok := params["frequency"]; ok {
			sc.SetFrequency(f.(float64))
		}

		if ph, ok := params["phase"]; ok {
			sc.SetPhase(ph.(float64))
		}

		if ti, ok := params["scanIndex"]; ok {
			sc.SetScanIndex(ti.(float64))
		}

		if amp, ok := params["amplitude"]; ok {
			sc.SetAmplitude(amp.(float64))
		}
	}

	return nil
}

func (sc *Scanner) Synthesize() bool {
	if !sc.BaseModule.Synthesize() {
		return false
	}

	freqInput := sc.InputAtIndex(0)
	phaseOffsetInput := sc.InputAtIndex(1)
	scanInput := sc.InputAtIndex(2)

	out := sc.OutputAtIndex(0).Buffer

	for i := 0; i < sc.Config.BufferSize; i++ {
		if freqInput.IsConnected() {
			sc.SetFrequency(freqInput.Buffer[i])
		}

		if phaseOffsetInput.IsConnected() {
			sc.OffsetPhase(phaseOffsetInput.Buffer[i])
		}

		if scanInput.IsConnected() {
			sc.SetScanIndex(scanInput.Buffer[i])
		}

		out[i] = sc.Generate()
	}

	return true
}
