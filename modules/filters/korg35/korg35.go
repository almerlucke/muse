package korg35

import (
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

type Korg35LPF struct {
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

func NewKorg35LPF(fc float64, config *muse.Configuration, id string) *Korg35LPF {
	return &Korg35LPF{
		BaseModule: muse.NewBaseModule(1, 1, config, id),
		lpf1:       newOnePole(fc, config.SampleRate),
		lpf2:       newOnePole(fc, config.SampleRate),
		hpf1:       newOnePole(fc, config.SampleRate),
		fc:         fc,
		k:          0.01,
		sat:        1.0,
		nlp:        true,
	}
}

func (klpf *Korg35LPF) reset() {
	klpf.lpf1.reset()
	klpf.lpf2.reset()
	klpf.hpf1.reset()
}

func (klpf *Korg35LPF) update() {
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

func (klpf *Korg35LPF) Cutoff() float64 {
	return klpf.fc
}

func (klpf *Korg35LPF) SetCutoff(fc float64) {
	klpf.fc = fc
	klpf.update()
}

func (klpf *Korg35LPF) Resonance() float64 {
	return klpf.k
}

func (klpf *Korg35LPF) SetResonance(res float64) {
	klpf.k = res
	klpf.update()
}

func (klpf *Korg35LPF) Saturation() float64 {
	return klpf.sat
}

func (klpf *Korg35LPF) SetSaturation(sat float64) {
	klpf.sat = sat
}

func (klpf *Korg35LPF) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Cutoff Frequency
		klpf.SetCutoff(value.(float64))
	case 1: // Resonance (0.01 - 2.0)
		klpf.SetResonance(value.(float64))
	case 2: // Saturation
		klpf.SetSaturation(value.(float64))
	}
}

func (klpf *Korg35LPF) ReceiveMessage(msg any) *muse.Message {
	content := msg.(map[string]any)

	if cutoff, ok := content["frequency"]; ok {
		klpf.SetCutoff(cutoff.(float64))
	}

	if res, ok := content["resonance"]; ok {
		klpf.SetResonance(res.(float64))
	}

	if sat, ok := content["saturation"]; ok {
		klpf.SetSaturation(sat.(float64))
	}

	return nil
}

func (klpf *Korg35LPF) Synthesize() bool {
	if !klpf.BaseModule.Synthesize() {
		return false
	}

	out := klpf.Outputs[0].Buffer
	in := klpf.Inputs[0].Buffer

	for i := 0; i < klpf.Config.BufferSize; i++ {
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

/*

Korg35LPF.prototype.run = function(xn){
  var y1 = this.lpf1.run(xn);
  var S35 = this.hpf1.getFeedbackOutput() + this.lpf2.getFeedbackOutput();

  var u = this.a * (y1 + S35);

  if (this.nlp) {
    u = tanh(this._sat * u);
  }

  var y = this.k * this.lpf2.run(u);

  this.hpf1.run(y);

  if (this.k > 0) y *= 1/this.k;

  return y;
};

function tanh(x){
  x = Math.exp(2 * x);
  return (x - 1) / (x + 1);
}
*/
