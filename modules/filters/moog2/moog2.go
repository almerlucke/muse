package moog2

import (
	"math"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/mmath"
)

// MoogFilter::MoogFilter(unsigned char Ftype, float Ffreq, float Fq,
// 	unsigned int srate, int bufsize)
// 	:Filter(srate, bufsize), sr(srate), gain(1.0f)
// {
// 	setfreq_and_q(Ffreq/srate, Fq);
// 	settype(Ftype); // q must be set before

// 	// initialize state but not to 0, in case of being used as denominator
// 	for (unsigned int i = 0; i<(sizeof(state)/sizeof(*state)); i++)
// 	{
// 			state[i] = 0.0001f;
// 	}
// }

// inline float MoogFilter::step(float input)
// {
// 	// transconductance
// 	// gM(vIn) = tanh( vIn ) / vIn
// 	float gm0 = tanhXdivX(state[0]);
// 	// to ease calculations the others are 1.0 (so daturation)
// 	// denominators
// 	float d0 = 1.0f / (1.0f + c*gm0);
// 	float dp1 = 1.0f / (1.0f + c);
// 	float dp2 = dp1*dp1;
// 	float dp3 = dp2*dp1;

// 	// instantaneous response estimate
// 	float y3Estimate =
// 			cp4 * d0 * gm0 * dp3 * input +
// 			cp3 * gm0 * dp3     * d0 * state[0] +
// 			cp2 * dp3 * state[1] +
// 			c   * dp2 * state[2] +
// 						dp1 * state[3];

// 	// mix input and gained feedback estimate for
// 	// cheaper feedback gain compensation. Idea from
// 	// Antti Huovilainen and Vesa Välimäki, "NEW APPROACHES TO DIGITAL SUBTRACTIVE SYNTHESIS"
// 	float u = input - tanhX(feedbackGain * (y3Estimate - 0.5f*input));
// 	// output of all stages
// 	float y0 = gm0 * d0 * (state[0] + c * u);
// 	float y1 = dp1 * (state[1] + c * y0);
// 	float y2 = dp1 * (state[2] + c * y1);
// 	float y3 = dp1 * (state[3] + c * y2);

// 	// update state
// 	state[0] += ct2 * (u - y0);
// 	state[1] += ct2 * (y0 - y1);
// 	state[2] += ct2 * (y1 - y2);
// 	state[3] += ct2 * (y2 - y3);

// 	// calculate multimode filter output
// 	return (a0 * u
// 				+ a1 * y0
// 				+ a2 * y1
// 				+ a3 * y2
// 				+ a4 * y3);
// }

// void MoogFilter::filterout(float *smp)
// {
// 	for (int i = 0; i < buffersize; i ++)
// 	{
// 			smp[i] = step(tanhX(smp[i]*gain));
// 			smp[i] *= outgain;
// 	}
// }

// void MoogFilter::setgain(float dBgain)
// {
// 	gain = dB2rap(dBgain);
// }

// void MoogFilter::settype(unsigned char ftype)
// {
// 	switch (ftype)
// 	{
// 			case 0:
// 					a0 = 1.0f; a1 =-4.0f; a2 = 6.0f; a3 =-4.0f; a4 = 1.0f;
// 					break;
// 			case 1:
// 					a0 = 0.0f; a1 = 0.0f; a2 = 4.0f; a3 =-8.0f; a4 = 4.0f;
// 					break;
// 			case 2:
// 			default:
// 					a0 = 0.0f; a1 = 0.0f; a2 = 0.0f; a3 = 0.0f; a4 = passbandCompensation;
// 					break;
// 	}
// }

func tan_2(x float64) float64 {
	//Pade approximation tan(x) hand tuned to map fCutoff
	x2 := x * x
	//~ return ((9.54f*x*((11.08f - x2)))/(105.0f - x2*(45.0f + x2))); // more accurate but instable at high frequencies
	return x + 0.15*x2 + 0.3*x2*x2 // faster, no division (by zero)
}

func tanhX(x float64) float64 {
	// Pade approximation of tanh(x) bound to [-1 .. +1]
	// https://mathr.co.uk/blog/2017-09-06_approximating_hyperbolic_tangent.html
	x2 := x * x
	return x * (105.0 + 10.0*x2) / (105.0 + (45.0+x2)*x2)
}

func tanhXdivX(x float64) float64 {
	//add DC offset to raise even harmonics (like transistor bias current)
	x += 0.1
	// pre calc often used term
	x2 := x * x
	// Pade approximation for tanh(x)/x used in filter stages (5x per sample)
	//~ return ((15.0+x2)/(15.0+6.0*x2));
	// faster approximation without division
	// tuned for more distortion in self oscillation
	return 1.0 - (0.35 * x2) + (0.06 * x2 * x2)
}

type Moog2 struct {
	*muse.BaseModule

	fType                 int
	fc                    float64
	q                     float64
	gain                  float64
	feedbackGain          float64
	a0, a1, a2, a3, a4    float64
	state                 [4]float64
	passbandCompensation  float64
	c, ct2, cp2, cp3, cp4 float64
}

func (m *Moog2) setGain(gain float64) {
	m.gain = mmath.Db2Rap(gain)
}

func (m *Moog2) setFreq(fc float64) {
	m.fc = fc

	ff := fc / m.Config.SampleRate
	// pre warp cutoff to map to reality
	m.c = tan_2(math.Pi * ff)
	// limit cutoff to prevent overflow
	m.c = mmath.Limit(m.c, 0.0006, 1.5)
	// pre calculate some stuff outside the hot zone
	m.ct2 = m.c * 2.0
	m.cp2 = m.c * m.c
	m.cp3 = m.cp2 * m.c
	m.cp4 = m.cp2 * m.cp2
}

func (m *Moog2) setQ(q float64) {
	m.q = q
	// flattening the Q input
	// self oscillation begins around 4.0
	// mapped to match the ANALOG filter class
	m.feedbackGain = math.Cbrt(q/1000.0)*4.0 + 0.3
	// compensation factor for passband reduction by the negative feedback
	m.passbandCompensation = 1.0 + mmath.Limit(m.feedbackGain, 0.0, 1.0)
}

func (m *Moog2) setType(fType int) {
	switch fType {
	case 0:
		m.a0 = 1.0
		m.a1 = -4.0
		m.a2 = 6.0
		m.a3 = -4.0
		m.a4 = 1.0
	case 1:
		m.a0 = 0.0
		m.a1 = 0.0
		m.a2 = 4.0
		m.a3 = -8.0
		m.a4 = 4.0
	case 2:
		m.a0 = 0.0
		m.a1 = 0.0
		m.a2 = 0.0
		m.a3 = 0.0
		m.a4 = m.passbandCompensation
	default:
		break
	}
}
