package stilson

import (
	"math"

	"github.com/almerlucke/muse"
)

var _saturation_limit = 0.95

var gainTable = []float64{
	0.999969, 0.990082, 0.980347, 0.970764, 0.961304, 0.951996, 0.94281, 0.933777, 0.924866, 0.916077, 0.90741, 0.898865, 0.890442, 0.882141,
	0.873962, 0.865906, 0.857941, 0.850067, 0.842346, 0.834686, 0.827148, 0.819733, 0.812378, 0.805145, 0.798004, 0.790955, 0.783997, 0.77713,
	0.770355, 0.763672, 0.75708, 0.75058, 0.744141, 0.737793, 0.731537, 0.725342, 0.719238, 0.713196, 0.707245, 0.701355, 0.695557, 0.689819,
	0.684174, 0.678558, 0.673035, 0.667572, 0.66217, 0.65686, 0.651581, 0.646393, 0.641235, 0.636169, 0.631134, 0.62619, 0.621277, 0.616425,
	0.611633, 0.606903, 0.602234, 0.597626, 0.593048, 0.588531, 0.584045, 0.579651, 0.575287, 0.570953, 0.566681, 0.562469, 0.558289, 0.554169,
	0.550079, 0.546051, 0.542053, 0.538116, 0.53421, 0.530334, 0.52652, 0.522736, 0.518982, 0.515289, 0.511627, 0.507996, 0.504425, 0.500885,
	0.497375, 0.493896, 0.490448, 0.487061, 0.483704, 0.480377, 0.477081, 0.473816, 0.470581, 0.467377, 0.464203, 0.46109, 0.457977, 0.454926,
	0.451874, 0.448883, 0.445892, 0.442932, 0.440033, 0.437134, 0.434265, 0.431427, 0.428619, 0.425842, 0.423096, 0.42038, 0.417664, 0.415009,
	0.412354, 0.409729, 0.407135, 0.404572, 0.402008, 0.399506, 0.397003, 0.394501, 0.392059, 0.389618, 0.387207, 0.384827, 0.382477, 0.380127,
	0.377808, 0.375488, 0.37323, 0.370972, 0.368713, 0.366516, 0.364319, 0.362122, 0.359985, 0.357849, 0.355713, 0.353607, 0.351532, 0.349457,
	0.347412, 0.345398, 0.343384, 0.34137, 0.339417, 0.337463, 0.33551, 0.333588, 0.331665, 0.329773, 0.327911, 0.32605, 0.324188, 0.322357, 0.320557,
	0.318756, 0.316986, 0.315216, 0.313446, 0.311707, 0.309998, 0.308289, 0.30658, 0.304901, 0.303223, 0.301575, 0.299927, 0.298309, 0.296692,
	0.295074, 0.293488, 0.291931, 0.290375, 0.288818, 0.287262, 0.285736, 0.284241, 0.282715, 0.28125, 0.279755, 0.27829, 0.276825, 0.275391,
	0.273956, 0.272552, 0.271118, 0.269745, 0.268341, 0.266968, 0.265594, 0.264252, 0.262909, 0.261566, 0.260223, 0.258911, 0.257599, 0.256317,
	0.255035, 0.25375,
}

func snapToZero(x float64) float64 {
	if !(x < -1.0e-8 || x > 1.0e-8) {
		return 0.0
	}

	return x
}

func saturate(input float64) float64 { //clamp without branching
	x1 := math.Abs(input + _saturation_limit)
	x2 := math.Abs(input - _saturation_limit)
	return 0.5 * (x1 - x2)
}

func crossfade(amount float64, a float64, b float64) float64 {
	return (1.0-amount)*a + amount*b
}

type StilsonMoog struct {
	*muse.BaseModule
	state     [4]float64
	p         float64
	q         float64
	out       float64
	freq      float64
	resonance float64
}

func NewStilsonMoog(freq float64, resonance float64, config *muse.Configuration, identifier string) *StilsonMoog {
	sm := &StilsonMoog{
		BaseModule: muse.NewBaseModule(3, 3, config, identifier),
		freq:       freq,
		resonance:  resonance,
	}

	sm.UpdatePQ()

	return sm
}

func (sm *StilsonMoog) UpdatePQ() {

	// Normalized cutoff between [0, 1]
	fc := sm.freq / sm.Config.SampleRate
	x2 := fc * fc
	x3 := x2 * fc

	// Frequency & amplitude correction (Cubic Fit)
	sm.p = -0.69346*x3 - 0.59515*x2 + 3.2937*fc - 1.0072

	ix := sm.p * 99.0
	ixint := int(math.Floor(ix))
	ixfrac := ix - float64(ixint)

	sm.q = sm.resonance * crossfade(ixfrac, gainTable[ixint+99], gainTable[ixint+100])
}

func (sm *StilsonMoog) Synthesize() bool {
	if !sm.BaseModule.Synthesize() {
		return false
	}

	inBuf := sm.Inputs[0].Buffer
	lowOutBuf := sm.Outputs[0].Buffer
	highOutBuf := sm.Outputs[1].Buffer
	bandOutBuf := sm.Outputs[2].Buffer

	/*
		// Scale by arbitrary value on account of our saturation function
				const float input = samples[s] * 0.65f;

				// Negative Feedback
				output = 0.25 * (input - output);

				for (int pole = 0; pole < 4; ++pole)
				{
					localState = state[pole];
					output = moog_saturate(output + p * (output - localState));
					state[pole] = output;
					output = moog_saturate(output + localState);
				}

				SNAP_TO_ZERO(output);
				samples[s] = output;
				output *= Q; // Scale stateful output by Q
	*/

	for i := 0; i < sm.Config.BufferSize; i++ {
		input := inBuf[i] * 0.65
		sm.out = 0.25 * (input - sm.out)

		for pole := 0; pole < 4; pole++ {
			tmp := sm.state[pole]
			sm.out = saturate(sm.out + sm.p*(sm.out-tmp))
			sm.state[pole] = sm.out
			sm.out = saturate(sm.out + tmp)
		}

		sm.out = snapToZero(sm.out)

		lowpass := sm.out
		highpass := input - sm.out
		bandpass := 3*sm.state[2] - lowpass

		lowOutBuf[i] = lowpass
		highOutBuf[i] = highpass
		bandOutBuf[i] = bandpass

		sm.out *= sm.q
	}

	return true
}
