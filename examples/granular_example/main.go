package main

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/granular"
)

type GParam struct {
	duration  float64
	amplitude float64
	panning   float64
	frequency float64
}

func (p *GParam) Duration() float64 {
	return p.duration
}

func (p *GParam) Amplitude() float64 {
	return p.amplitude
}

func (p *GParam) Panning() float64 {
	return p.panning
}

func (p *GParam) EnvType() granular.EnvelopeType {
	return granular.Parabolic
}

func (p *GParam) Attack() float64 {
	return 0.1
}

func (p *GParam) Release() float64 {
	return 0.5
}

func (p *GParam) Frequency() float64 {
	return p.frequency
}

func NewParam(d float64, a float64, p float64, f float64) *GParam {
	return &GParam{
		duration:  d,
		amplitude: a,
		panning:   p,
		frequency: f,
	}
}

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

func (p *SFParam) EnvType() granular.EnvelopeType {
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
	buffer *muse.SoundFileBuffer
	offset float64
	phase  float64
	delta  float64
}

func (s *SFSource) Synthesize() float64 {
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

	out := s.buffer.Channels[0][lookup1]
	return out + (s.buffer.Channels[0][lookup2]-out)*fraction
}

func (s *SFSource) Activate(p *SFParam, c *muse.Configuration) {
	s.phase = 0.0
	s.offset = p.offset
	s.delta = (p.speed * s.buffer.SampleRate / c.SampleRate) / float64(s.buffer.NumFrames)
}

type GSource struct {
	phase float64
	delta float64
}

func (s *GSource) Synthesize() float64 {
	out := math.Sin(s.phase * math.Pi * 2.0)
	// out := s.phase

	s.phase += s.delta

	for s.phase >= 1.0 {
		s.phase -= 1.0
	}

	for s.phase < 0.0 {
		s.phase += 1.0
	}

	return out
}

func (s *GSource) Activate(p *GParam, c *muse.Configuration) {
	s.delta = p.frequency / c.SampleRate
	s.phase = 0.0
}

type SFSourceFactory struct {
	Samples *muse.SoundFileBuffer
}

func (sf *SFSourceFactory) NewSource() granular.Source[*SFParam] {
	return &SFSource{
		buffer: sf.Samples,
	}
}

type GSourceFactory struct{}

func (sf *GSourceFactory) NewSource() granular.Source[*GParam] {
	return &GSource{}
}

type GSequencer struct {
}

func randBetween(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func quantize(v float64, binSize float64) float64 {
	return math.Round(v/binSize) * binSize
}

type SFSequencer struct {
}

func (gs *SFSequencer) NextStep(timestamp int64, config *muse.Configuration) *granular.Step[*SFParam] {
	s := &granular.Step[*SFParam]{}

	offset := math.Mod(float64(timestamp)/config.SampleRate, 32.0) / 32.0

	s.InterOnset = int(randBetween(0.00003, 0.0012) * config.SampleRate)
	s.Parameter = NewSFParam(
		randBetween(0.004, 0.38),
		randBetween(0.2, 1.0),
		randBetween(0.0, 1.0),
		randBetween(1.0-0.1*offset, 1.0+0.1*offset),
		randBetween(offset*0.97-0.03*offset, offset*0.97+0.03*offset),
	)

	return s
}

func (gs *GSequencer) NextStep(timestamp int64, config *muse.Configuration) *granular.Step[*GParam] {
	s := &granular.Step[*GParam]{}
	seconds := math.Mod(float64(timestamp)/config.SampleRate, 32)

	if seconds < 8 {
		s.Parameter = NewParam(
			quantize(randBetween(0.04, 3.0), 0.1),
			randBetween(0.1, 0.8),
			randBetween(0.0, 1.0),
			quantize(randBetween(100.0, 550.0), 100))
		s.InterOnset = int(quantize(randBetween(0.03, 0.12), 0.02) * config.SampleRate)
	} else if seconds < 16 {
		s.Parameter = NewParam(
			quantize(randBetween(0.04, 0.05), 0.01),
			randBetween(0.7, 0.9),
			randBetween(0.0, 1.0),
			quantize(randBetween(200.0, 250.0), 40))
		s.InterOnset = int(quantize(randBetween(0.003, 0.012), 0.001) * config.SampleRate)
	} else if seconds < 24 {
		s.Parameter = NewParam(
			quantize(randBetween(0.04, 1.0), 0.1),
			randBetween(0.1, 0.8),
			randBetween(0.0, 1.0),
			quantize(randBetween(200.0, 350.0), 100))
		s.InterOnset = int(quantize(randBetween(0.01, 0.1), 0.02) * config.SampleRate)
	} else if seconds < 32 {
		s.Parameter = NewParam(
			quantize(randBetween(0.02, 0.05), 0.01),
			randBetween(0.1, 0.8),
			randBetween(0.0, 1.0),
			quantize(randBetween(100.0, 2550.0), 100))
		s.InterOnset = int(quantize(randBetween(0.03, 0.12), 0.02) * config.SampleRate)
	}

	return s
}

func main() {
	rand.Seed(time.Now().UnixNano())

	env := muse.NewEnvironment(2, 44100, 1024)

	sfb, err := muse.NewSoundFileBuffer("/Users/almerlucke/Downloads/female-laugh_C#_minor.wav")
	if err != nil {
		log.Fatalf("fatal err: %v", err)
	}

	gr := &granular.Granulator[*SFParam]{}
	gr.Initialize(&SFSourceFactory{Samples: sfb}, 400, &SFSequencer{}, env.Config, "granulator")

	env.AddModule(gr)

	muse.Connect(gr, 0, env, 0)
	muse.Connect(gr, 1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 32.0)
}
