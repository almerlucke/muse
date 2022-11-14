package phasor

type Phasor struct {
	phase     float64
	inc       float64
	outVector [1]float64
}

func NewPhasor(fc float64, fs float64, phase float64) *Phasor {
	p := &Phasor{}
	p.SetFrequency(fc, fs)
	p.SetPhase(phase)
	return p
}

func (p *Phasor) Phase() float64 {
	return p.phase
}

func (p *Phasor) SetPhase(phase float64) {
	p.phase = phase
}

func (p *Phasor) SetFrequency(fc float64, fs float64) {
	p.inc = fc / fs
}

func (p *Phasor) NumDimensions() int {
	return 1
}

func (p *Phasor) Tick() []float64 {
	out := p.phase

	p.phase += p.inc

	for p.phase >= 1.0 {
		p.phase -= 1.0
	}

	for p.phase < 0.0 {
		p.phase += 1.0
	}

	p.outVector[0] = out
	return p.outVector[:]
}
