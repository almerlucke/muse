package main

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/modules/granular"
	"github.com/mkb218/gosndfile/sndfile"
)

type SFParam struct {
	duration  float64
	amplitude float64
	panning   float64
	speed     float64
	offset    float64
}

func (p *SFParam) Duration() float64 {
	return p.duration
}

func (p *SFParam) Amplitude() float64 {
	return p.amplitude
}

func (p *SFParam) Panning() float64 {
	return p.panning
}

func (p *SFParam) EnvType() granular.DefaultEnvelopeType {
	return granular.Parabolic
}

func (p *SFParam) Attack() float64 {
	return 0.1
}

func (p *SFParam) Release() float64 {
	return 0.5
}

func (p *SFParam) Speed() float64 {
	return p.speed
}

func (p *SFParam) Offset() float64 {
	return p.offset
}

func NewSFParam(d float64, a float64, p float64, s float64, o float64) *SFParam {
	return &SFParam{
		duration:  d,
		amplitude: a,
		panning:   p,
		speed:     s,
		offset:    o,
	}
}

type SFSource struct {
	buffer *io.SoundFileBuffer
	offset float64
	phase  float64
	delta  float64
}

func (s *SFSource) Synthesize(outBuffers [][]float64, bufSize int) {

	for i := 0; i < bufSize; i++ {
		offset := s.offset + s.phase

		for offset >= 1.0 {
			offset -= 1.0
		}

		for offset < 0.0 {
			offset += 1.0
		}

		offset = offset * float64(s.buffer.NumFrames)
		lookup1 := int64(offset)
		fraction := offset - float64(lookup1)
		lookup2 := lookup1 + 1

		if lookup2 >= s.buffer.NumFrames {
			lookup2 = 0
		}

		s.phase += s.delta

		for s.phase >= 1.0 {
			s.phase -= 1.0
		}

		for s.phase < 0.0 {
			s.phase += 1.0
		}

		for outIndex, outBuf := range outBuffers {
			out := s.buffer.Channels[outIndex][lookup1]

			outBuf[i] = out + (s.buffer.Channels[outIndex][lookup2]-out)*fraction
		}
	}
}

func (s *SFSource) Activate(p granular.Parameter, c *muse.Configuration) {
	// pan := p.Panning()
	// g.panLeft = math.Cos(pan * math.Pi / 2.0)
	// g.panRight = math.Sin(pan * math.Pi / 2.0)

	sfp := p.(*SFParam)
	s.phase = 0.0
	s.offset = sfp.offset
	s.delta = (sfp.speed * s.buffer.SampleRate / c.SampleRate) / float64(s.buffer.NumFrames)
}

type SFSourceFactory struct {
	Samples *io.SoundFileBuffer
}

func (sf *SFSourceFactory) New() granular.Source {
	return &SFSource{
		buffer: sf.Samples,
	}
}

func randBetween(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func quantize(v float64, binSize float64) float64 {
	return math.Round(v/binSize) * binSize
}

type SFParameterGenerator struct {
	parameter SFParam // single parameter can be passed for each NextParameter call, prevent from allocating new params
}

func (f *SFParameterGenerator) Next(timestamp int64, config *muse.Configuration) (granular.Parameter, int64) {
	offset := math.Mod(float64(timestamp)/config.SampleRate, 34.0) / 34.0

	interOnset := int64(randBetween(0.00003, 0.0012) * config.SampleRate)

	f.parameter.duration = randBetween(4.0, 400.0)
	f.parameter.amplitude = randBetween(0.2, 1.0)
	f.parameter.panning = randBetween(0.0, 1.0)
	f.parameter.speed = randBetween(1.0-0.2*offset, 1.0+0.2*offset)
	f.parameter.offset = randBetween(offset*0.8-0.20*offset, offset*0.8+0.2*offset)

	return &f.parameter, interOnset
}

func main() {
	rand.Seed(time.Now().UnixNano())

	sfb, err := io.NewSoundFileBuffer("resources/sounds/John_1-1.wav")
	if err != nil {
		log.Fatalf("fatal err: %v", err)
	}

	numChannels := len(sfb.Channels)

	env := muse.NewEnvironment(numChannels, 3*44100, 1024)

	gr := env.AddModule(granular.NewGranulator(numChannels, &SFSourceFactory{Samples: sfb}, &granular.DefaultEnvelopeFactory{}, 400, &SFParameterGenerator{}, env.Config, "granulator"))

	for i := 0; i < numChannels; i++ {
		gr.Connect(i, env, i)
	}

	env.SynthesizeToFile("/Users/almerlucke/Desktop/john.aiff", 34.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
}
