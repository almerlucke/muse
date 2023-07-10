package main

import (
	"github.com/almerlucke/muse/components/generator"
	"github.com/almerlucke/muse/components/ops"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/mkb218/gosndfile/sndfile"
)

func main() {
	dx7 := ops.NewDX7Algo1(waveshaping.NewSineTable(2048), 100.0, 44100.0)

	dx7.Operator(0).LevelEnvelope().Rates = [4]float64{0.95, 0.8, 0.8, 0.8}
	dx7.Operator(1).LevelEnvelope().Rates = [4]float64{0.95, 0.8, 0.8, 0.8}
	dx7.Operator(2).LevelEnvelope().Rates = [4]float64{0.95, 0.8, 0.8, 0.8}
	dx7.Operator(3).LevelEnvelope().Rates = [4]float64{0.95, 0.8, 0.8, 0.8}
	dx7.Operator(4).LevelEnvelope().Rates = [4]float64{0.95, 0.8, 0.8, 0.8}
	dx7.Operator(5).LevelEnvelope().Rates = [4]float64{0.95, 0.8, 0.8, 0.8}
	dx7.PitchEnvelope().Levels = [4]float64{0.516, 0.503, 0.495, 0.5}
	dx7.PitchEnvelope().Rates = [4]float64{0.95, 0.8, 0.8, 0.8}
	dx7.SetReleaseMode(ops.EnvelopeDurationRelease)
	dx7.NoteOn(400.0, 1.0, 2.0)
	generator.WriteToFile(dx7, "/Users/almerlucke/Desktop/test.aiff", 8.0, 44100, sndfile.SF_FORMAT_AIFF)

	// rate 1.0 -> 0.001 s --- 0.001 s
	// rate 0.0 -> 90 s --- 210 s
	//

	// env := ops.NewEnvelope([4]float64{1.0, 0.5, 0.3, 0.0}, [4]float64{0.95, 0.95, 0.95, 0.95}, 44100.0, ops.EnvelopeAutomaticRelease)
	// vec := []float64{}

	// env.TriggerHard(0)

	// for i := 0; i < 4000; i++ {
	// 	v := env.Tick()
	// 	//log.Printf("%f", v)
	// 	vec = append(vec, v)
	// }

	// // env.NoteOff()

	// for i := 0; i < 16000; i++ {
	// 	v := env.Tick()
	// 	//log.Printf("%f", v)
	// 	vec = append(vec, v)
	// }

	// plot.PlotVector(vec, 1000, 500, "/Users/almerlucke/Desktop/env.png")

	// vec = []float64{}
	// for i := 0; i < 2000; i++ {
	// 	v := ops.RateToSeconds(float64(i)/2000.0, ops.Rising, 2.5)
	// 	// log.Printf("%f", v)
	// 	vec = append(vec, v)
	// }

	// plot.PlotVector(vec, 1000, 500, "/Users/almerlucke/Desktop/decay.png")
}
