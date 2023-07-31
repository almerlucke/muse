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

// bufferLen = numChannels * muse bufferSize
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

/*
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

func (sw *SoundWriter) writeToFile(filePath string, frames []float64, format sndfile.Format) error {
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

		return sw.writeToFile(filePath, unsafe.Slice((*float64)(unsafe.Pointer(&resampled[0])), len(resampled)/8), format)
	}

	return sw.writeToFile(filePath, frames, format)
}

// WriteFramesToFile writes sample frames to a sound file and close
func WriteFramesToFile(frames []float64, numChannels int, inputSampleRate int, outputSampleRate int, normalizeOutput bool, format sndfile.Format, file string) error {
	sw := NewSoundWriter(numChannels, inputSampleRate, outputSampleRate, normalizeOutput)

	sw.WriteFrames(frames)

	return sw.Finish(file, format)
}
*/
