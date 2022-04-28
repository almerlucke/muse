package main

import (
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
	return 0.0
}

func (p *GParam) Release() float64 {
	return 0.0
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

type GSource struct {
	phase float64
	delta float64
}

func (s *GSource) Synthesize() float64 {
	out := math.Sin(s.phase * math.Pi * 2.0)

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
}

type GSourceFactory struct{}

func (sf *GSourceFactory) NewSource() granular.Source[*GParam] {
	return &GSource{}
}

/*
type Sequence[P Parameter] interface {
	NextStep() *Step[P]
}
*/

type GSequencer struct {
}

func randBetween(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func (gs *GSequencer) NextStep(config *muse.Configuration) *granular.Step[*GParam] {
	s := &granular.Step[*GParam]{}
	s.Parameter = NewParam(randBetween(0.03, 0.07), randBetween(0.1, 0.8), randBetween(0.0, 1.0), randBetween(150.0, 1400.0))
	s.InterOnset = int(randBetween(0.001, 0.03) * config.SampleRate)
	return s
}

func main() {
	rand.Seed(time.Now().UnixNano())

	env := muse.NewEnvironment(2, 44100, 128)

	gr := &granular.Granulator[*GParam]{}
	gr.Initialize("granulator", &GSourceFactory{}, 20, &GSequencer{}, env.Config)

	env.AddModule(gr)

	muse.Connect(gr, 0, env, 0)
	muse.Connect(gr, 1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 4.0)
}
