package muse

import (
	"github.com/almerlucke/muse/io"
	"github.com/mkb218/gosndfile/sndfile"
)

type Environment struct {
	*BasePatch
	Config *Configuration
}

func NewEnvironment(numOutputs int, sampleRate float64, bufferSize int) *Environment {
	config := &Configuration{SampleRate: sampleRate, BufferSize: bufferSize}

	return &Environment{
		BasePatch: NewPatch(0, numOutputs, config, "environment"),
		Config:    config,
	}
}

func (e *Environment) Synthesize() bool {
	e.PostMessages()
	e.PrepareSynthesis()
	return e.BasePatch.Synthesize()
}

func (e *Environment) SynthesizeToFile(filePath string, numSeconds float64, format sndfile.Format) error {
	numChannels := e.NumOutputs()

	swr := io.NewSoundWriter(numChannels, int(e.Config.SampleRate), 44100, true)

	interleaveBuffer := make([]float64, e.NumOutputs()*e.Config.BufferSize)

	framesToProduce := int64(e.Config.SampleRate * numSeconds)

	for framesToProduce > 0 {
		e.Synthesize()

		interleaveIndex := 0

		numFrames := e.Config.BufferSize

		if framesToProduce <= int64(e.Config.BufferSize) {
			numFrames = int(framesToProduce)
		}

		for i := 0; i < numFrames; i++ {
			for c := 0; c < numChannels; c++ {
				interleaveBuffer[interleaveIndex] = e.OutputAtIndex(c).Buffer[i]
				interleaveIndex++
			}
		}

		swr.WriteFrames(interleaveBuffer[:numFrames*numChannels])

		framesToProduce -= int64(numFrames)
	}

	return swr.Finish(filePath, format)
}
