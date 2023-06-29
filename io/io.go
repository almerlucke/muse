package io

import (
	"bytes"
	"fmt"
	"math"
	"unsafe"

	"github.com/almerlucke/muse/buffer"
	"github.com/almerlucke/muse/dsp/filters"
	"github.com/almerlucke/muse/dsp/windows"
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

type SoundFiler interface {
	NumChannels() int
	SampleRate() float64
	NumFrames() int64
	Duration() float64
	Depth() int
	Buffer(channel int, depth int) buffer.Buffer
	Lookup(pos float64, channel int, depth int, wrap bool) float64
	LookupAll(pos float64, depth int, wrap bool) []float64
}

type SoundBank map[string]SoundFiler

// SoundFile contains sound file deinterleaved samples and implements SoundFiler interface
type SoundFile struct {
	// Deinterleaved channels
	channels [][]float64

	// Sample rate
	sampleRate float64

	// Number of frames
	numFrames int64

	// Duration in seconds
	duration float64
}

// NewSoundFile load sound file from disk deinterleaved
func NewSoundFile(filePath string) (*SoundFile, error) {
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

	sf := SoundFile{}
	sf.duration = float64(info.Frames) / float64(info.Samplerate)
	sf.numFrames = info.Frames
	sf.channels = channels
	sf.sampleRate = float64(info.Samplerate)

	return &sf, nil
}

func (sf *SoundFile) NumChannels() int {
	return len(sf.channels)
}

func (sf *SoundFile) SampleRate() float64 {
	return sf.sampleRate
}

func (sf *SoundFile) NumFrames() int64 {
	return sf.numFrames
}

func (sf *SoundFile) Duration() float64 {
	return sf.duration
}

func (sf *SoundFile) Depth() int {
	return 1
}

func (sf *SoundFile) Buffer(channel int, depth int) buffer.Buffer {
	return buffer.Buffer(sf.channels[channel])
}

func (sf *SoundFile) Lookup(pos float64, channel int, depth int, wrap bool) float64 {
	return buffer.Buffer(sf.channels[channel]).Lookup(pos, wrap)
}

func (sf *SoundFile) LookupAll(pos float64, depth int, wrap bool) []float64 {
	out := make([]float64, len(sf.channels))

	l := int(sf.numFrames)
	integer1 := int(pos)
	fraction := pos - float64(integer1)

	var integer2 int

	if wrap {
		integer2 = (integer1 + 1) % l
	} else {
		integer2 = integer1 + 1
		if integer2 >= l {
			integer2 = l - 1
		}
	}

	for c := 0; c < len(sf.channels); c++ {
		buf := sf.channels[c]
		sample1 := buf[integer1]
		sample2 := buf[integer2]
		out[c] = sample1 + (sample2-sample1)*fraction
	}

	return out
}

func SpeedToMipMapDepth(speed float64) int {
	speed = math.Abs(speed)

	whole, fract := math.Modf(speed)

	depth := int(whole)
	if fract < 0.0001 && depth > 0 {
		depth -= 1
	}

	return depth
}

type MipMap struct {
	depth   int
	buffers []buffer.Buffer
}

func NewMipMap(buf buffer.Buffer, sampleRate float64, depth int) (*MipMap, error) {
	mm := &MipMap{
		depth:   depth,
		buffers: make([]buffer.Buffer, depth),
	}

	mm.buffers[0] = buf

	fc := sampleRate / 2.0 // Nyquist start

	for d := 1; d < depth; d++ {
		dfc := fc / float64(d+1)
		fir := &filters.FIR{
			Sinc: &filters.Sinc{
				CutOffFreq:   dfc,
				SamplingFreq: int(sampleRate),
				Taps:         200,
				Window:       windows.Hamming,
			},
		}

		filteredBuf, err := fir.LowPass(buf)
		if err != nil {
			return nil, err
		}

		mm.buffers[d] = buffer.Buffer(filteredBuf)
	}

	return mm, nil
}

func (mm *MipMap) Length() int {
	return len(mm.buffers[0])
}

func (mm *MipMap) Depth() int {
	return mm.depth
}

func (mm *MipMap) Lookup(pos float64, depth int, wrap bool) float64 {
	return mm.buffers[depth].Lookup(pos, wrap)
}

func (mm *MipMap) Buffer(depth int) buffer.Buffer {
	return mm.buffers[depth]
}

type MipMapSoundFile struct {
	channels   []*MipMap
	sampleRate float64
	numFrames  int64
	duration   float64
	depth      int
}

func NewMipMapSoundFile(filePath string, depth int) (*MipMapSoundFile, error) {
	sndFile, err := NewSoundFile(filePath)
	if err != nil {
		return nil, err
	}

	mmsf := &MipMapSoundFile{
		depth:      depth,
		sampleRate: sndFile.SampleRate(),
		numFrames:  sndFile.NumFrames(),
		duration:   sndFile.Duration(),
		channels:   make([]*MipMap, sndFile.NumChannels()),
	}

	for channel := 0; channel < sndFile.NumChannels(); channel++ {
		mm, err := NewMipMap(sndFile.Buffer(channel, 0), mmsf.sampleRate, depth)
		if err != nil {
			return nil, err
		}

		mmsf.channels[channel] = mm
	}

	return mmsf, nil
}

func (mmsf *MipMapSoundFile) NumChannels() int {
	return len(mmsf.channels)
}

func (mmsf *MipMapSoundFile) SampleRate() float64 {
	return mmsf.sampleRate
}

func (mmsf *MipMapSoundFile) NumFrames() int64 {
	return mmsf.numFrames
}

func (mmsf *MipMapSoundFile) Duration() float64 {
	return mmsf.duration
}

func (mmsf *MipMapSoundFile) Depth() int {
	return mmsf.depth
}

func (mmsf *MipMapSoundFile) Buffer(channel int, depth int) buffer.Buffer {
	return mmsf.channels[channel].Buffer(depth)
}

func (mmsf *MipMapSoundFile) Lookup(pos float64, channel int, depth int, wrap bool) float64 {
	return mmsf.channels[channel].Lookup(pos, depth, wrap)
}

func (mmsf *MipMapSoundFile) LookupAll(pos float64, depth int, wrap bool) []float64 {
	out := make([]float64, len(mmsf.channels))

	l := int(mmsf.numFrames)
	integer1 := int(pos)
	fraction := pos - float64(integer1)

	var integer2 int

	if wrap {
		integer2 = (integer1 + 1) % l
	} else {
		integer2 = integer1 + 1
		if integer2 >= l {
			integer2 = l - 1
		}
	}

	for c := 0; c < len(mmsf.channels); c++ {
		buf := mmsf.channels[c].Buffer(depth)
		sample1 := buf[integer1]
		sample2 := buf[integer2]
		out[c] = sample1 + (sample2-sample1)*fraction
	}

	return out
}

type WaveTableSoundFile struct {
	Tables    []buffer.Buffer
	TableSize int
}

func NewWaveTableSoundFile(filePath string, tableSize int) (*WaveTableSoundFile, error) {
	sndFile, err := NewSoundFile(filePath)
	if err != nil {
		return nil, err
	}

	numTables := int(sndFile.numFrames) / tableSize
	remaining := int(sndFile.numFrames) % tableSize

	if remaining != 0 {
		return nil, fmt.Errorf(
			"wavetable file did not contain exact multiple of table size %d: numTables = %d,  remaining = %d", tableSize, numTables, remaining,
		)
	}

	wsf := &WaveTableSoundFile{
		Tables:    make([]buffer.Buffer, numTables),
		TableSize: tableSize,
	}

	buf := sndFile.channels[0]
	offset := 0

	for i := 0; i < numTables; i++ {
		wsf.Tables[i] = buf[offset : offset+tableSize]
		offset += tableSize
	}

	return wsf, nil
}
