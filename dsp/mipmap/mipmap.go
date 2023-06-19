package mipmap

import (
	"math"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/dsp/filters"
	"github.com/almerlucke/muse/dsp/windows"
	"github.com/almerlucke/muse/io"
)

func SpeedToDepth(speed float64) int {
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
	buffers []muse.Buffer
}

func NewMipMap(buffer muse.Buffer, sampleRate float64, depth int) (*MipMap, error) {
	mm := &MipMap{
		depth:   depth,
		buffers: make([]muse.Buffer, depth),
	}

	mm.buffers[0] = buffer

	fc := sampleRate / 2.0 // Nyquist start

	for d := 1; d < depth; d++ {
		dfc := fc / float64(d+1)
		fir := &filters.FIR{
			Sinc: &filters.Sinc{
				CutOffFreq:   dfc,
				SamplingFreq: int(sampleRate),
				Taps:         40,
				Window:       windows.Hamming,
			},
		}

		filteredBuf, err := fir.LowPass(buffer)
		if err != nil {
			return nil, err
		}

		mm.buffers[d] = muse.Buffer(filteredBuf)
	}

	return mm, nil
}

func (mm *MipMap) Length() int {
	return len(mm.buffers[0])
}

func (mm *MipMap) Depth() int {
	return mm.depth
}

func (mm *MipMap) Lookup(pos float64, depth int) float64 {
	return mm.buffers[depth].Lookup(pos)
}

func (mm *MipMap) Buffer(depth int) muse.Buffer {
	return mm.buffers[depth]
}

type MipMapSoundFileBuffer struct {
	Channels   []*MipMap
	SampleRate float64
	NumFrames  int64
	Duration   float64
	Depth      int
}

func NewMipMapSoundFileBuffer(filePath string, depth int) (*MipMapSoundFileBuffer, error) {
	sndFileBuf, err := io.NewSoundFileBuffer(filePath)
	if err != nil {
		return nil, err
	}

	mmsfb := &MipMapSoundFileBuffer{
		Depth:      depth,
		SampleRate: sndFileBuf.SampleRate,
		NumFrames:  sndFileBuf.NumFrames,
		Duration:   sndFileBuf.Duration,
		Channels:   make([]*MipMap, len(sndFileBuf.Channels)),
	}

	for channel := 0; channel < len(sndFileBuf.Channels); channel++ {
		mm, err := NewMipMap(sndFileBuf.Channels[channel], sndFileBuf.SampleRate, depth)
		if err != nil {
			return nil, err
		}

		mmsfb.Channels[channel] = mm
	}

	return mmsfb, nil
}

func (mmsfb *MipMapSoundFileBuffer) Buffer(channel int, depth int) muse.Buffer {
	return mmsfb.Channels[channel].Buffer(depth)
}

func (mmsfb *MipMapSoundFileBuffer) Lookup(pos float64, channel int, depth int) float64 {
	return mmsfb.Channels[channel].Lookup(pos, depth)
}
