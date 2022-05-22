package main

import (
	"log"
	"math"

	pshape "github.com/almerlucke/muse/components/phaseshaping"
	"github.com/almerlucke/muse/io"
	"github.com/mkb218/gosndfile/sndfile"
)

func minimoogVoyagerSawtooth(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewLinear(0.25, 0.0),
		pshape.NewFunction(func(s float64) float64 { return math.Sin(2.0 * math.Pi * s) }),
		pshape.NewBipolar(),
	}

	return pd
}

func hardSync(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewLinear(2.5, 0.0),
		pshape.NewMod1(),
		pshape.NewBipolar()}

	return pd
}

func softSyncTriangle(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 80.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewBipolar(),
		pshape.NewAbs(),
		pshape.NewLinear(1.25, 0.0),
		pshape.NewMod1(),
		pshape.NewTri(),
		pshape.NewBipolar(),
	}

	return pd
}

func jp8000triMod(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewBipolar(),
		pshape.NewAbs(),
		pshape.NewLinear(2.0, -1.0),
		pshape.NewMod1(),
		pshape.NewMult(0.7),
		pshape.NewFunction(
			func(x float64) float64 {
				return 2.0 * (x - math.Ceil(x-0.5))
			},
		),
	}

	return pd
}

func pulseWidthMod(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewLinear(1.25, 0.0),
		pshape.NewMod1(),
		pshape.NewPulse(0.4),
		pshape.NewBipolar(),
	}

	return pd
}

// g o mod o g [x(n), a ] =

func main() {
	sr := 3 * 44100.0
	numSamps := 4 * int(sr)

	pd := pulseWidthMod(sr)
	pulse := pd.Shapers[2].(*pshape.Pulse)

	samps := make([]float64, numSamps)

	for i := 0; i < numSamps; i++ {
		n := float64(i) * 1.0 / float64(numSamps)

		pulse.W = 0.2 + n*0.7

		// log.Printf("mult.M  %v", mult.M)

		samps[i] = pd.Tick()
	}

	err := io.WriteFramesToFile(samps, 1, int(sr), 44100, true, sndfile.SF_FORMAT_AIFF, "/Users/almerlucke/Desktop/shaper.aiff")
	if err != nil {
		log.Printf("err %v", err)
	}
}
