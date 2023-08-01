package io

import (
	"github.com/almerlucke/muse/buffer"
	"github.com/almerlucke/muse/io/sndfile/aifc"
	"github.com/dh1tw/gosamplerate"
)

type Writer struct {
	src              gosamplerate.Src
	bufferSize       int
	interleaveBuffer []float32
	sndfile          *aifc.AIFC
	numChannels      int
	ratio            float64
	normalize        bool
	max              float32
}

func NewWriter(filePath string, numChannels int, bufferSize int, inputSampleRate int, outputSampleRate int, converterType int, normalize bool) (*Writer, error) {
	src, err := gosamplerate.New(converterType, numChannels, bufferSize*numChannels)
	if err != nil {
		return nil, err
	}

	sndfile, err := aifc.Open(filePath, int16(numChannels), float64(outputSampleRate))
	if err != nil {
		return nil, err
	}

	return &Writer{
		src:              src,
		numChannels:      numChannels,
		ratio:            float64(outputSampleRate) / float64(inputSampleRate),
		sndfile:          sndfile,
		bufferSize:       bufferSize,
		interleaveBuffer: make([]float32, numChannels*bufferSize),
	}, nil
}

func (wr *Writer) WriteBuffers(buffers []buffer.Buffer, endOfInput bool) error {
	for i := 0; i < wr.bufferSize; i++ {
		for j := 0; j < wr.numChannels; j++ {
			samp := float32(buffers[j][i])
			if samp > wr.max {
				wr.max = samp
			}
			wr.interleaveBuffer[i*wr.numChannels+j] = samp
		}
	}

	output, err := wr.src.Process(wr.interleaveBuffer, wr.ratio, endOfInput)
	if err != nil {
		return err
	}

	if len(output) > 0 {
		err = wr.sndfile.WriteItems(output)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wr *Writer) Close() error {
	normalizeErr := wr.sndfile.Normalize(wr.max)
	deleteErr := gosamplerate.Delete(wr.src)
	closeErr := wr.sndfile.Close()

	if normalizeErr != nil {
		return normalizeErr
	}
	if deleteErr != nil {
		return deleteErr
	}
	if closeErr != nil {
		return closeErr
	}

	return nil
}
