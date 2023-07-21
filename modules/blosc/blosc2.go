package blosc

import (
	"math"

	"github.com/almerlucke/muse"
)

type Waveform int

const (
	SINE Waveform = iota
	COSINE
	TRIANGLE
	SQUARE
	RECTANGLE
	SAWTOOTH
	RAMP
	MODIFIED_TRIANGLE
	MODIFIED_SQUARE
	HALF_WAVE_RECTIFIED_SINE
	FULL_WAVE_RECTIFIED_SINE
	TRIANGULAR_PULSE
	TRAPEZOID_FIXED
	TRAPEZOID_VARIABLE
)

const (
	two_PI float64 = math.Pi * 2.0
)

func squareNumber(x float64) float64 {
	return x * x
}

func blep(t float64, dt float64) float64 {
	if t < dt {
		return -squareNumber(t/dt - 1)
	} else if t > 1-dt {
		return squareNumber((t-1)/dt + 1)
	}

	return 0
}

func blamp(t float64, dt float64) float64 {
	if t < dt {
		t = t/dt - 1.0
		return -1.0 / 3.0 * squareNumber(t) * t
	} else if t > 1.0-dt {
		t = (t-1.0)/dt + 1.0
		return 1.0 / 3.0 * squareNumber(t) * t
	}

	return 0
}

func integerPart(x float64) float64 {
	ip, _ := math.Modf(x)
	return ip
}

type Osc2 struct {
	*muse.BaseModule

	wf  Waveform
	fc  float64
	dt  float64
	amp float64
	pw  float64
	t   float64
}

func NewOsc2(fc float64, t float64, pw float64, amp float64, wf Waveform, config *muse.Configuration) *Osc2 {
	osc := &Osc2{
		BaseModule: muse.NewBaseModule(3, 1, config, ""),
		t:          t,
	}

	osc.SetSelf(osc)
	osc.setFrequency(fc)
	osc.setPulseWidth(pw)
	osc.setWaveform(wf)
	osc.setAmplitude(amp)

	return osc
}

func (osc *Osc2) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Frequency
		osc.setFrequency(value.(float64))
	case 1: // Pulse Width
		osc.setPulseWidth(value.(float64))
	case 2: // Amplitude
		osc.setAmplitude(value.(float64))
	case 3: // Waveform
		osc.setWaveform(Waveform(value.(int)))
	}
}

func (osc *Osc2) ReceiveMessage(msg any) []*muse.Message {
	if params, ok := msg.(map[string]any); ok {
		if f, ok := params["frequency"]; ok {
			osc.setFrequency(f.(float64))
		}

		if pw, ok := params["pulseWidth"]; ok {
			osc.setPulseWidth(pw.(float64))
		}

		if amp, ok := params["amplitude"]; ok {
			osc.setAmplitude(amp.(float64))
		}

		if wf, ok := params["waveform"]; ok {
			if wff, ok := wf.(float64); ok {
				osc.setWaveform(Waveform(int(wff)))
			}
			if wfi, ok := wf.(int); ok {
				osc.setWaveform(Waveform(wfi))
			}
		}
	}

	return nil
}

func (osc *Osc2) Synthesize() bool {
	if !osc.BaseModule.Synthesize() {
		return false
	}

	freqInput := osc.InputAtIndex(0)
	pwInput := osc.InputAtIndex(1)
	ampInput := osc.InputAtIndex(2)

	out := osc.OutputAtIndex(0).Buffer

	for i := 0; i < osc.Config.BufferSize; i++ {
		if freqInput.IsConnected() {
			osc.setFrequency(freqInput.Buffer[i])
		}
		if pwInput.IsConnected() {
			osc.setPulseWidth(pwInput.Buffer[i])
		}
		if ampInput.IsConnected() {
			osc.setAmplitude(ampInput.Buffer[i])
		}

		out[i] = osc.getAndInc()
	}

	return true
}

func (osc *Osc2) setFrequency(fc float64) {
	osc.fc = fc
	osc.dt = fc / osc.Config.SampleRate
}

func (osc *Osc2) setPhase(t float64) {
	osc.t = t
}

func (osc *Osc2) setPulseWidth(pulseWidth float64) {
	osc.pw = pulseWidth
}

func (osc *Osc2) sync(phase float64) {
	osc.t = phase

	if osc.t >= 0.0 {
		osc.t -= integerPart(osc.t)
	} else {
		osc.t += 1.0 - integerPart(osc.t)
	}
}

func (osc *Osc2) setWaveform(wf Waveform) {
	osc.wf = wf
}

func (osc *Osc2) setAmplitude(amp float64) {
	osc.amp = amp
}

func (osc *Osc2) get() float64 {
	switch osc.wf {
	case SINE:
		return osc.sin()
	case COSINE:
		return osc.cos()
	case TRIANGLE:
		return osc.tri()
	case SQUARE:
		return osc.sqr()
	case RECTANGLE:
		return osc.rect()
	case SAWTOOTH:
		return osc.saw()
	case RAMP:
		return osc.ramp()
	case MODIFIED_TRIANGLE:
		return osc.tri2()
	case MODIFIED_SQUARE:
		return osc.sqr2()
	case HALF_WAVE_RECTIFIED_SINE:
		return osc.half()
	case FULL_WAVE_RECTIFIED_SINE:
		return osc.full()
	case TRIANGULAR_PULSE:
		return osc.trip()
	case TRAPEZOID_FIXED:
		return osc.trap()
	case TRAPEZOID_VARIABLE:
		return osc.trap2()
	default:
		return 0.0
	}
}

func (osc *Osc2) inc() {
	osc.t += osc.dt
	osc.t -= integerPart(osc.t)
}

func (osc *Osc2) getAndInc() float64 {
	sample := osc.get()
	osc.inc()
	return sample
}

func (osc *Osc2) sin() float64 {
	return osc.amp * math.Sin(two_PI*osc.t)
}

func (osc *Osc2) cos() float64 {
	return osc.amp * math.Cos(two_PI*osc.t)
}

func (osc *Osc2) half() float64 {
	t2 := osc.t + 0.5
	t2 -= integerPart(t2)

	y := -2.0 / math.Pi

	if osc.t < 0.5 {
		y = 2.0*math.Sin(two_PI*osc.t) - 2.0/math.Pi
	}

	y += two_PI * osc.dt * (blamp(osc.t, osc.dt) + blamp(t2, osc.dt))

	return osc.amp * y
}

func (osc *Osc2) full() float64 {
	t1 := osc.t + 0.25
	t1 -= integerPart(t1)

	y := 2.0*math.Sin(math.Pi*t1) - 4.0/math.Pi
	y += two_PI * osc.dt * blamp(t1, osc.dt)

	return osc.amp * y
}

func (osc *Osc2) tri() float64 {
	t1 := osc.t + 0.25
	t1 -= integerPart(t1)

	t2 := osc.t + 0.75
	t2 -= integerPart(t2)

	y := osc.t * 4.0

	if y >= 3.0 {
		y -= 4.0
	} else if y > 1.0 {
		y = 2.0 - y
	}

	y += 4.0 * osc.dt * (blamp(t1, osc.dt) - blamp(t2, osc.dt))

	return osc.amp * y
}

func (osc *Osc2) tri2() float64 {
	pw := math.Max(0.0001, math.Min(0.9999, osc.pw))

	t1 := osc.t + 0.5*pw
	t1 -= integerPart(t1)

	t2 := osc.t + 1.0 - 0.5*pw
	t2 -= integerPart(t2)

	y := osc.t * 2.0

	if y >= 2.0-pw {
		y = (y - 2.0) / pw
	} else if y >= pw {
		y = 1.0 - (y-pw)/(1.0-pw)
	} else {
		y /= pw
	}

	y += osc.dt / (pw - pw*pw) * (blamp(t1, osc.dt) - blamp(t2, osc.dt))

	return osc.amp * y
}

func (osc *Osc2) trip() float64 {
	t1 := osc.t + 0.75 + 0.5*osc.pw
	t1 -= integerPart(t1)

	var y float64
	if t1 >= osc.pw {
		y = -osc.pw
	} else {
		y = 4.0 * t1
		if y >= 2.0*osc.pw {
			y = 4.0 - y/osc.pw - osc.pw
		} else {
			y = y/osc.pw - osc.pw
		}
	}

	if osc.pw > 0.0 {
		t2 := t1 + 1.0 - 0.5*osc.pw
		t2 -= integerPart(t2)

		t3 := t1 + 1.0 - osc.pw
		t3 -= integerPart(t3)
		y += 2.0 * osc.dt / osc.pw * (blamp(t1, osc.dt) - 2*blamp(t2, osc.dt) + blamp(t3, osc.dt))
	}

	return osc.amp * y
}

func (osc *Osc2) trap() float64 {
	y := 4 * osc.t
	if y >= 3.0 {
		y -= 4.0
	} else if y > 1.0 {
		y = 2.0 - y
	}
	y = math.Max(-1, math.Min(1.0, 2.0*y))

	t1 := osc.t + 0.125
	t1 -= integerPart(t1)

	t2 := t1 + 0.5
	t2 -= integerPart(t2)

	// Triangle #1
	y += 4.0 * osc.dt * (blamp(t1, osc.dt) - blamp(t2, osc.dt))

	t1 = osc.t + 0.375
	t1 -= integerPart(t1)

	t2 = t1 + 0.5
	t2 -= integerPart(t2)

	// Triangle #2
	y += 4.0 * osc.dt * (blamp(t1, osc.dt) - blamp(t2, osc.dt))

	return osc.amp * y
}

func (osc *Osc2) trap2() float64 {
	pw := math.Min(0.9999, osc.pw)
	scale := 1.0 / (1.0 - pw)

	y := 4.0 * osc.t
	if y >= 3.0 {
		y -= 4.0
	} else if y > 1.0 {
		y = 2.0 - y
	}

	y = math.Max(-1.0, math.Min(1.0, scale*y))

	t1 := osc.t + 0.25 - 0.25*pw
	t1 -= integerPart(t1)

	t2 := t1 + 0.5
	t2 -= integerPart(t2)

	// Triangle #1
	y += scale * 2.0 * osc.dt * (blamp(t1, osc.dt) - blamp(t2, osc.dt))

	t1 = osc.t + 0.25 + 0.25*pw
	t1 -= integerPart(t1)

	t2 = t1 + 0.5
	t2 -= integerPart(t2)

	// Triangle #2
	y += scale * 2.0 * osc.dt * (blamp(t1, osc.dt) - blamp(t2, osc.dt))

	return osc.amp * y
}

func (osc *Osc2) sqr() float64 {
	t2 := osc.t + 0.5
	t2 -= integerPart(t2)

	var y float64
	if osc.t < 0.5 {
		y = 1.0
	} else {
		y = -1.0
	}

	y += blep(osc.t, osc.dt) - blep(t2, osc.dt)

	return osc.amp * y
}

func (osc *Osc2) sqr2() float64 {
	t1 := osc.t + 0.875 + 0.25*(osc.pw-0.5)
	t1 -= integerPart(t1)

	t2 := osc.t + 0.375 + 0.25*(osc.pw-0.5)
	t2 -= integerPart(t2)

	// Square #1
	var y float64

	if t1 < 0.5 {
		y = 1.0
	} else {
		y = -1.0
	}

	y += blep(t1, osc.dt) - blep(t2, osc.dt)

	t1 += 0.5 * (1.0 - osc.pw)
	t1 -= integerPart(t1)

	t2 += 0.5 * (1.0 - osc.pw)
	t2 -= integerPart(t2)

	// Square #2
	if t1 < 0.5 {
		y += 1.0
	} else {
		y -= 1.0
	}

	y += blep(t1, osc.dt) - blep(t2, osc.dt)

	return osc.amp * 0.5 * y
}

func (osc *Osc2) rect() float64 {
	t2 := osc.t + 1.0 - osc.pw
	t2 -= integerPart(t2)

	y := -2.0 * osc.pw
	if osc.t < osc.pw {
		y += 2
	}

	y += blep(osc.t, osc.dt) - blep(t2, osc.dt)

	return osc.amp * y
}

func (osc *Osc2) saw() float64 {
	t1 := osc.t + 0.5
	t1 -= integerPart(t1)

	y := 2.0*t1 - 1.0
	y -= blep(t1, osc.dt)

	return osc.amp * y
}

func (osc *Osc2) ramp() float64 {
	t1 := osc.t
	t1 -= integerPart(t1)

	y := 1.0 - 2.0*t1
	y += blep(t1, osc.dt)

	return osc.amp * y
}

/*

    PolyBLEP(double sampleRate, Waveform waveform = SINE, double initialFrequency = 440.0);

    virtual ~PolyBLEP();

    void setFrequency(double freqInHz);

    void setSampleRate(double sampleRate);

    void setWaveform(Waveform waveform);

    void setPulseWidth(double pw);

    double get() const;

    void inc();

    double getAndInc();

    double getFreqInHz() const;

    void sync(double phase);

protected:
    Waveform waveform;
    double sampleRate;
    double freqInSecondsPerSample;
    double amplitude; // Frequency dependent gain [0.0..1.0]
    double pulseWidth; // [0.0..1.0]
    double t; // The current phase [0.0..1.0) of the oscillator.

    void setdt(double time);

    double sin() const;

    double cos() const;

    double half() const;

    double full() const;

    double tri() const;

    double tri2() const;

    double trip() const;

    double trap() const;

    double trap2() const;

    double sqr() const;

    double sqr2() const;

    double rect() const;

    double saw() const;

    double ramp() const;
};

#include "PolyBLEP.h"

#define _USE_MATH_DEFINES

#include <math.h>
#include <cmath>
#include <cstdint>

const double TWO_PI = 2 * M_PI;

template<typename T>
inline T square_number(const T &x) {
    return x * x;
}

// Adapted from "Phaseshaping Oscillator Algorithms for Musical Sound
// Synthesis" by Jari Kleimola, Victor Lazzarini, Joseph Timoney, and Vesa
// Valimaki.
// http://www.acoustics.hut.fi/publications/papers/smc2010-phaseshaping/
inline double blep(double t, double dt) {
    if (t < dt) {
        return -square_number(t / dt - 1);
    } else if (t > 1 - dt) {
        return square_number((t - 1) / dt + 1);
    } else {
        return 0;
    }
}

// Derived from blep().
inline double blamp(double t, double dt) {
    if (t < dt) {
        t = t / dt - 1;
        return -1 / 3.0 * square_number(t) * t;
    } else if (t > 1 - dt) {
        t = (t - 1) / dt + 1;
        return 1 / 3.0 * square_number(t) * t;
    } else {
        return 0;
    }
}

template<typename T>
inline int64_t bitwiseOrZero(const T &t) {
    return static_cast<int64_t>(t) | 0;
}

PolyBLEP::PolyBLEP(double sampleRate, Waveform waveform, double initialFrequency)
        : waveform(waveform), sampleRate(sampleRate), amplitude(1.0), t(0.0) {
    setSampleRate(sampleRate);
    setFrequency(initialFrequency);
    setWaveform(waveform);
    setPulseWidth(0.5);
}

PolyBLEP::~PolyBLEP() {

}

void PolyBLEP::setdt(double time) {
    freqInSecondsPerSample = time;
}

void PolyBLEP::setFrequency(double freqInHz) {
    setdt(freqInHz / sampleRate);
}

void PolyBLEP::setSampleRate(double sampleRate) {
    const double freqInHz = getFreqInHz();
    this->sampleRate = sampleRate;
    setFrequency(freqInHz);
}

double PolyBLEP::getFreqInHz() const {
    return freqInSecondsPerSample * sampleRate;
}

void PolyBLEP::setPulseWidth(double pulseWidth) {
    this->pulseWidth = pulseWidth;
}

void PolyBLEP::sync(double phase) {
    t = phase;
    if (t >= 0) {
        t -= bitwiseOrZero(t);
    } else {
        t += 1 - bitwiseOrZero(t);
    }
}

void PolyBLEP::setWaveform(Waveform waveform) {
    this->waveform = waveform;
}

double PolyBLEP::get() const {
    if(getFreqInHz() >= sampleRate / 4) {
        return sin();
    } else switch (waveform) {
        case SINE:
            return sin();
        case COSINE:
            return cos();
        case TRIANGLE:
            return tri();
        case SQUARE:
            return sqr();
        case RECTANGLE:
            return rect();
        case SAWTOOTH:
            return saw();
        case RAMP:
            return ramp();
        case MODIFIED_TRIANGLE:
            return tri2();
        case MODIFIED_SQUARE:
            return sqr2();
        case HALF_WAVE_RECTIFIED_SINE:
            return half();
        case FULL_WAVE_RECTIFIED_SINE:
            return full();
        case TRIANGULAR_PULSE:
            return trip();
        case TRAPEZOID_FIXED:
            return trap();
        case TRAPEZOID_VARIABLE:
            return trap2();
        default:
            return 0.0;
    }
}

void PolyBLEP::inc() {
    t += freqInSecondsPerSample;
    t -= bitwiseOrZero(t);
}

double PolyBLEP::getAndInc() {
    const double sample = get();
    inc();
    return sample;
}


*/
