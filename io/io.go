package io

import (
	"bytes"
	"math"
	"unsafe"

	"github.com/mkb218/gosndfile/sndfile"
	"github.com/zaf/resample"
)

type SoundWriter struct {
	frames           []float64
	numChannels      int
	inputSampleRate  int
	outputSampleRate int
	normalizeOutput  bool
	peak             float64
}

func NewSoundWriter(numChannels int, inputSampleRate int, outputSampleRate int, normalizeOutput bool) *SoundWriter {
	return &SoundWriter{
		frames:           []float64{},
		numChannels:      numChannels,
		inputSampleRate:  inputSampleRate,
		outputSampleRate: outputSampleRate,
		normalizeOutput:  normalizeOutput,
	}
}

func (sw *SoundWriter) WriteFrames(frames []float64) {
	for _, v := range frames {
		va := math.Abs(v)
		if va > sw.peak {
			sw.peak = va
		}
	}

	sw.frames = append(sw.frames, frames...)
}

func (sw *SoundWriter) resample() ([]byte, error) {
	originalBytes := unsafe.Slice((*byte)(unsafe.Pointer(&sw.frames[0])), len(sw.frames)*8)

	var outBuffer bytes.Buffer

	res, err := resample.New(&outBuffer, float64(sw.inputSampleRate), float64(sw.outputSampleRate), int(sw.numChannels), resample.F64, resample.VeryHighQ)
	if err != nil {
		return nil, err
	}

	_, err = res.Write(originalBytes)
	if err != nil {
		return nil, err
	}

	return outBuffer.Bytes(), nil
}

func (sw *SoundWriter) Finish(filePath string, format sndfile.Format) error {
	frames := sw.frames

	if sw.normalizeOutput && sw.peak > 0 {
		norm := 1.0 / sw.peak
		for i, v := range frames {
			frames[i] = v * norm
		}
	}

	if sw.inputSampleRate != sw.outputSampleRate {
		resampled, err := sw.resample()
		if err != nil {
			return err
		}

		frames = unsafe.Slice((*float64)(unsafe.Pointer(&resampled[0])), len(resampled)/8)
	}

	outputInfo := sndfile.Info{}
	outputInfo.Channels = int32(sw.numChannels)
	outputInfo.Format = format | sndfile.SF_FORMAT_DOUBLE
	outputInfo.Samplerate = int32(sw.outputSampleRate)

	outputFile, err := sndfile.Open(filePath, sndfile.Write, &outputInfo)
	if err != nil {
		return err
	}

	outputFile.WriteItems(frames)

	return outputFile.Close()
}

// WriteFramesToFile writes sample frames to a sound file and close
func WriteFramesToFile(frames []float64, numChannels int, inputSampleRate int, outputSampleRate int, normalizeOutput bool, format sndfile.Format, file string) error {
	sw := NewSoundWriter(numChannels, inputSampleRate, outputSampleRate, normalizeOutput)

	sw.WriteFrames(frames)

	return sw.Finish(file, format)
}

/*
   Sound buffer
*/

// SoundFileBuffer contains sound file deinterleaved samples
type SoundFileBuffer struct {
	// Deinterleaved channels
	Channels [][]float64

	// Sample rate
	SampleRate float64

	// Number of frames
	NumFrames int64

	// Duration in seconds
	Duration float64
}

// NewSoundFileBuffer load sound file from disk deinterleaved
func NewSoundFileBuffer(filePath string) (*SoundFileBuffer, error) {
	info := sndfile.Info{}

	file, err := sndfile.Open(filePath, sndfile.Read, &info)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Create one big buffer to hold all samples
	fileBuffer := make([]float64, int64(info.Channels)*info.Frames)

	// Create separate channels by splitting buffer into info.Channels parts
	channels := make([][]float64, info.Channels)
	for i := int32(0); i < info.Channels; i++ {
		channels[i] = fileBuffer[int64(i)*info.Frames : int64(i+1)*info.Frames]
	}

	// Deinterleave in blocks
	sampleBlockSize := int64(2048) * int64(info.Channels)
	samples := make([]float64, sampleBlockSize)
	frameIndex := int64(0)

	for {
		framesRead, err := file.ReadFrames(samples)
		if err != nil {
			return nil, err
		}

		if framesRead == 0 {
			break
		}

		for i := int64(0); i < framesRead; i++ {
			for j := int64(0); j < int64(info.Channels); j++ {
				channels[j][frameIndex+i] = samples[i*int64(info.Channels)+j]
			}
		}

		frameIndex += framesRead
	}

	buffer := SoundFileBuffer{}
	buffer.Duration = float64(info.Frames) / float64(info.Samplerate)
	buffer.NumFrames = info.Frames
	buffer.Channels = channels
	buffer.SampleRate = float64(info.Samplerate)

	return &buffer, nil
}
