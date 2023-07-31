package oversampler

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/buffer"
	"github.com/dh1tw/gosamplerate"
)

/*
Run inner module at higher samplerate, downsample output
*/

type Oversampler struct {
	*muse.BaseModule
	module           muse.Module
	ratio            float64
	oversamplingRate int
	src              gosamplerate.Src
	interleaveBuffer []float32
	numChannels      int
	outBuffers       []buffer.Buffer
}

func New(module muse.Module, oversamplingRate int, converterType int) (*Oversampler, error) {
	numChannels := module.NumOutputs()
	bufferSize := module.Configuration().BufferSize

	src, err := gosamplerate.New(converterType, numChannels, bufferSize*numChannels)
	if err != nil {
		return nil, err
	}

	osa := &Oversampler{
		BaseModule:       muse.NewBaseModule(0, numChannels),
		module:           module,
		numChannels:      numChannels,
		oversamplingRate: oversamplingRate,
		ratio:            1.0 / float64(oversamplingRate),
		src:              src,
		interleaveBuffer: make([]float32, bufferSize*numChannels),
		outBuffers:       make([]buffer.Buffer, numChannels),
	}

	osa.SetSelf(osa)

	for c := 0; c < numChannels; c++ {
		osa.outBuffers[c] = module.OutputAtIndex(c).Buffer
	}

	return osa, nil
}

func (osa *Oversampler) ReceiveControlValue(value any, index int) {
	osa.module.ReceiveControlValue(value, index)
}

func (osa *Oversampler) ReceiveMessage(msg any) []*muse.Message {
	return osa.module.ReceiveMessage(msg)
}

func (osa *Oversampler) Synthesize() bool {
	if !osa.BaseModule.Synthesize() {
		return false
	}

	bufIndex := 0

	for i := 0; i < osa.oversamplingRate && bufIndex < osa.Config.BufferSize; i++ {
		osa.module.PrepareSynthesis()
		osa.module.Synthesize()

		for j := 0; j < osa.Config.BufferSize; j++ {
			for c := 0; c < osa.numChannels; c++ {
				osa.interleaveBuffer[j*osa.numChannels+c] = float32(osa.outBuffers[c][j])
			}
		}

		interleavedOut, _ := osa.src.Process(osa.interleaveBuffer, float64(osa.ratio), false)

		numFrames := len(interleavedOut) / osa.numChannels
		maxFrames := osa.Config.BufferSize - bufIndex
		if numFrames > maxFrames {
			numFrames = maxFrames
		}

		for j := 0; j < numFrames; j++ {
			for c := 0; c < osa.numChannels; c++ {
				osa.Outputs[c].Buffer[bufIndex+j] = float64(interleavedOut[j*osa.numChannels+c])
			}
		}

		bufIndex += numFrames
	}

	return true
}
