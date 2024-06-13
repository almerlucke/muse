package korg35

import (
	"github.com/almerlucke/muse/modules/filters"
	"math"

	"github.com/almerlucke/muse"
)

func prewarp(f float64, sr float64) float64 {
	return math.Tan(math.Pi * f / sr)
}

func tanh(x float64) float64 {
	x = math.Exp(2.0 * x)
	return (x - 1.0) / (x + 1.0)
}

type onePole struct {
	fc float64
	sr float64
	a  float64
	b  float64
	z1 float64
}

func newOnePole(fc float64, sr float64) *onePole {
	op := &onePole{
		fc: fc,
		sr: sr,
		a:  1.0,
		b:  1.0,
	}

	op.reset()
	op.update()

	return op
}

func (op *onePole) reset() {
	op.z1 = 0.0
}

func (op *onePole) getFeedbackOutput() float64 {
	return op.z1 * op.b
}

func (op *onePole) update() {
	wa := prewarp(op.fc, op.sr)
	op.a = wa / (1.0 + wa)
}

func (op *onePole) setCutoff(fc float64) {
	op.fc = fc
	op.update()
}

func (op *onePole) lpTick(xn float64) float64 {
	vn := (xn - op.z1) * op.a
	out := vn + op.z1
	op.z1 = vn + out
	return out
}

func (op *onePole) hpTick(xn float64) float64 {
	vn := (xn - op.z1) * op.a
	lpOut := vn + op.z1
	op.z1 = vn + lpOut
	hpOut := xn - lpOut

	return hpOut
}

func (op *onePole) apTick(xn float64) float64 {
	vn := (xn - op.z1) * op.a
	lpOut := vn + op.z1
	op.z1 = vn + lpOut
	hpOut := xn - lpOut
	apOut := lpOut - hpOut

	return apOut
}

type LPF struct {
	*muse.BaseModule
	lpf1 *onePole
	lpf2 *onePole
	hpf1 *onePole
	fc   float64
	a    float64
	k    float64
	sat  float64
	nlp  bool
}

type Factory struct{}

func (f *Factory) New(cfg any) filters.Filter {
	fCfg := cfg.(*filters.FilterConfig)
	return New(fCfg.Frequency, fCfg.Resonance, fCfg.Drive)
}

func DefaultConfig() *filters.FilterConfig {
	return filters.NewFilterConfig(1500.0, 0.9, 2.0, 0)
}

func New(fc float64, res float64, sat float64) *LPF {
	sr := muse.SampleRate()
	korg := &LPF{
		BaseModule: muse.NewBaseModule(4, 1),
		lpf1:       newOnePole(fc, sr),
		lpf2:       newOnePole(fc, sr),
		hpf1:       newOnePole(fc, sr),
		fc:         fc,
		k:          res,
		sat:        sat,
		nlp:        true,
	}

	korg.SetSelf(korg)

	return korg
}

func (klpf *LPF) reset() {
	klpf.lpf1.reset()
	klpf.lpf2.reset()
	klpf.hpf1.reset()
}

func (klpf *LPF) update() {
	g := prewarp(klpf.fc, klpf.Config.SampleRate)
	G := g / (1 + g)
	k := klpf.k

	klpf.lpf1.a = G
	klpf.lpf2.a = G
	klpf.hpf1.a = G

	klpf.lpf2.b = (k - k*G) / (1 + g)
	klpf.hpf1.b = -1 / (1 + g)

	klpf.a = 1 / (1 - k*G + k*G*G)
}

func (klpf *LPF) Frequency() float64 {
	return klpf.fc
}

func (klpf *LPF) SetFrequency(fc float64) {
	klpf.fc = fc
	klpf.update()
}

func (klpf *LPF) Resonance() float64 {
	return klpf.k
}

func (klpf *LPF) SetResonance(res float64) {
	klpf.k = res
	klpf.update()
}

func (klpf *LPF) Drive() float64 {
	return klpf.sat
}

func (klpf *LPF) SetDrive(drive float64) {
	klpf.sat = drive
}

func (klpf *LPF) Type() int { return 0 }

func (klpf *LPF) SetType(_ int) {}

func (klpf *LPF) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Cutoff Frequency
		klpf.SetFrequency(value.(float64))
	case 1: // Resonance (0.01 - 2.0)
		klpf.SetResonance(value.(float64))
	case 2: // Saturation
		klpf.SetDrive(value.(float64))
	}
}

func (klpf *LPF) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if fc, ok := content["frequency"]; ok {
		klpf.SetFrequency(fc.(float64))
	}

	if res, ok := content["resonance"]; ok {
		klpf.SetResonance(res.(float64))
	}

	if sat, ok := content["drive"]; ok {
		klpf.SetDrive(sat.(float64))
	}

	return nil
}

func (klpf *LPF) Synthesize() bool {
	if !klpf.BaseModule.Synthesize() {
		return false
	}

	out := klpf.Outputs[0].Buffer
	in := klpf.Inputs[0].Buffer

	for i := 0; i < klpf.Config.BufferSize; i++ {
		if klpf.Inputs[1].IsConnected() {
			klpf.SetFrequency(klpf.Inputs[1].Buffer[i])
		}

		if klpf.Inputs[2].IsConnected() {
			klpf.SetResonance(klpf.Inputs[2].Buffer[i])
		}

		if klpf.Inputs[3].IsConnected() {
			klpf.SetDrive(klpf.Inputs[3].Buffer[i])
		}

		y1 := klpf.lpf1.lpTick(in[i])

		s35 := klpf.hpf1.getFeedbackOutput() + klpf.lpf2.getFeedbackOutput()
		u := klpf.a * (y1 + s35)

		if klpf.nlp {
			u = tanh(klpf.sat * u)
		}

		y := klpf.k * klpf.lpf2.lpTick(u)

		klpf.hpf1.hpTick(y)

		if klpf.k > 0.0 {
			y *= 1.0 / klpf.k
		}

		out[i] = y
	}

	return true
}
