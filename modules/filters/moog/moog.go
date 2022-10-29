package moog

import (
	"math"

	"github.com/almerlucke/muse"
)

/*
This model is based on a reference implementation of an algorithm developed by
Stefano D'Angelo and Vesa Valimaki, presented in a paper published at ICASSP in 2013.
This improved model is based on a circuit analysis and compared against a reference
Ngspice simulation. In the paper, it is noted that this particular model is
more accurate in preserving the self-oscillating nature of the real filter.
References: "An Improved Virtual Analog Model of the Moog Ladder Filter"
Original Implementation: D'Angelo, Valimaki
*/

const (
	VT = 0.312
)

type Moog struct {
	*muse.BaseModule
	v     [4]float64
	dV    [4]float64
	tV    [4]float64
	fc    float64
	res   float64
	x     float64
	g     float64
	drive float64
}

func NewMoog(fc float64, res float64, drive float64, config *muse.Configuration, identifier string) *Moog {
	m := &Moog{
		BaseModule: muse.NewBaseModule(4, 1, config, identifier),
	}

	m.SetDrive(drive)
	m.SetResonance(res)
	m.SetFrequency(fc)

	m.SetSelf(m)

	return m
}

func (m *Moog) Frequency() float64 {
	return m.fc
}

func (m *Moog) SetFrequency(fc float64) {
	m.fc = fc
	m.x = (math.Pi * fc) / m.Config.SampleRate
	m.g = 4.0 * math.Pi * VT * fc * (1.0 - m.x) / (1.0 + m.x)
}

func (m *Moog) Resonance() float64 {
	return m.res
}

func (m *Moog) SetResonance(res float64) {
	m.res = res * 4.0
}

func (m *Moog) Drive() float64 {
	return m.drive
}

func (m *Moog) SetDrive(drive float64) {
	m.drive = drive
}

func (m *Moog) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Cutoff Frequency
		m.SetFrequency(value.(float64))
	case 1: // Resonance
		m.SetResonance(value.(float64))
	case 2: // Drive
		m.SetDrive(value.(float64))
	}
}

func (m *Moog) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if fc, ok := content["frequency"].(float64); ok {
		m.SetFrequency(fc)
	}

	if res, ok := content["resonance"].(float64); ok {
		m.SetResonance(res)
	}

	if drive, ok := content["drive"].(float64); ok {
		m.SetDrive(drive)
	}

	return nil
}

func (m *Moog) Synthesize() bool {
	if !m.BaseModule.Synthesize() {
		return false
	}

	in := m.Inputs[0].Buffer
	out := m.Outputs[0].Buffer

	var dV0, dV1, dV2, dV3 float64

	for i := 0; i < m.Config.BufferSize; i++ {

		if m.Inputs[1].IsConnected() {
			m.SetFrequency(m.Inputs[1].Buffer[i])
		}

		if m.Inputs[2].IsConnected() {
			m.SetResonance(m.Inputs[2].Buffer[i])
		}

		if m.Inputs[3].IsConnected() {
			m.SetDrive(m.Inputs[3].Buffer[i])
		}

		dV0 = -m.g * (math.Tanh((m.drive*in[i]+m.res*m.v[3])/(2.0*VT)) + m.tV[0])
		m.v[0] += (dV0 + m.dV[0]) / (2.0 * m.Config.SampleRate)
		m.dV[0] = dV0
		m.tV[0] = math.Tanh(m.v[0] / (2.0 * VT))

		dV1 = m.g * (m.tV[0] - m.tV[1])
		m.v[1] += (dV1 + m.dV[1]) / (2.0 * m.Config.SampleRate)
		m.dV[1] = dV1
		m.tV[1] = math.Tanh(m.v[1] / (2.0 * VT))

		dV2 = m.g * (m.tV[1] - m.tV[2])
		m.v[2] += (dV2 + m.v[2]) / (2.0 * m.Config.SampleRate)
		m.dV[2] = dV2
		m.tV[2] = math.Tanh(m.v[2] / (2.0 * VT))

		dV3 = m.g * (m.tV[2] - m.tV[3])
		m.v[3] += (dV3 + m.dV[3]) / (2.0 * m.Config.SampleRate)
		m.dV[3] = dV3
		m.tV[3] = math.Tanh(m.v[3] / (2.0 * VT))

		out[i] = m.v[3]
	}

	return true
}
