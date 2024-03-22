package main

import (
	"github.com/almerlucke/sndfile"
	"log"
	"math"
	"math/rand"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/granular"
	"github.com/almerlucke/muse/modules/granular/envelopes/trapezoidal"
	"github.com/almerlucke/muse/utils/float"
	museRand "github.com/almerlucke/muse/utils/rand"
)

type LookupMode int

const (
	Wrap LookupMode = iota
	Mirror
)

type SFParam struct {
	duration   float64
	amplitude  float64
	panning    float64
	speed      float64
	offset     float64
	lookupMode LookupMode
}

func (p *SFParam) Duration() float64 {
	return p.duration
}

func (p *SFParam) Amplitude() float64 {
	return p.amplitude
}

func (p *SFParam) Attack() float64 {
	return 0.01
}

func (p *SFParam) Release() float64 {
	return 0.96
}

func (p *SFParam) Smoothness() float64 {
	return 1.0
}

func (p *SFParam) Panning() float64 {
	return p.panning
}

func (p *SFParam) Speed() float64 {
	return p.speed
}

func (p *SFParam) Offset() float64 {
	return p.offset
}

func (p *SFParam) LookupMode() LookupMode {
	return p.lookupMode
}

func NewSFParam(duration float64, amplitude float64, panning float64, speed float64, offset float64, lookupMode LookupMode) *SFParam {
	return &SFParam{
		duration:   duration,
		amplitude:  amplitude,
		panning:    panning,
		speed:      speed,
		offset:     offset,
		lookupMode: lookupMode,
	}
}

type SFSource struct {
	sf         sndfile.SoundFiler
	phase      float64
	delta      float64
	speed      float64
	panLeft    float64
	panRight   float64
	lookupMode LookupMode
}

func (s *SFSource) Synthesize(outBuffers [][]float64, bufSize int) {
	pan := [2]float64{s.panLeft, s.panRight}

	for i := 0; i < bufSize; i++ {
		pos := s.phase * float64(s.sf.NumFrames())
		s.phase = float.ZeroIfSmall(s.phase + s.delta)

		switch s.lookupMode {
		case Wrap:
			for s.phase >= 1.0 {
				s.phase -= 1.0
			}

			for s.phase < 0.0 {
				s.phase += 1.0
			}
		case Mirror:
			if s.phase >= 1.0 {
				s.phase = 2.0 - s.phase
				s.delta *= -1.0
			}
			if s.phase < 0.0 {
				s.phase = -s.phase
				s.delta *= -1.0
			}
		}

		depth := sndfile.SpeedToMipMapDepth(s.speed)
		if depth >= s.sf.Depth() {
			depth = s.sf.Depth() - 1
		}

		samps := s.sf.LookupAll(pos, depth, true)

		for outIndex, outBuf := range outBuffers {

			outBuf[i] = pan[outIndex] * samps[outIndex]
		}
	}
}

func (s *SFSource) Activate(sampsToGo int64, p granular.Parameter, c *muse.Configuration) {
	// pan := p.Panning()
	// g.panLeft = math.Cos(pan * math.Pi / 2.0)
	// g.panRight = math.Sin(pan * math.Pi / 2.0)

	sfp := p.(*SFParam)
	s.phase = sfp.offset
	s.delta = (sfp.speed * s.sf.SampleRate() / c.SampleRate) / float64(s.sf.NumFrames())
	s.speed = sfp.speed
	s.lookupMode = sfp.lookupMode

	if s.sf.NumChannels() == 2 {
		s.panLeft = math.Cos(sfp.panning * math.Pi / 2.0)
		s.panRight = math.Sin(sfp.panning * math.Pi / 2.0)
	} else {
		s.panLeft = 1.0
		s.panRight = 1.0
	}

	if s.phase >= 1.0 {
		s.phase = 0.9999
	}

	if s.phase < 0.0 {
		s.phase = 0.0
	}
}

type SFSourceFactory struct {
	SoundFile sndfile.SoundFiler
}

func (sf *SFSourceFactory) New() granular.Source {
	return &SFSource{
		sf: sf.SoundFile,
	}
}

func randBetween(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func quantize(v float64, binSize float64) float64 {
	return math.Round(v/binSize) * binSize
}

type SFParameterGenerator struct {
	parameter           SFParam // single parameter can be passed for each NextParameter call, prevent from allocating new params
	amplitudeClustering *museRand.ClusterRand
	offsetClustering    *museRand.ClusterRand
	speedClustering     *museRand.ClusterRand
	durationClustering  *museRand.ClusterRand
	onsetClustering     *museRand.ClusterRand
	panClustering       *museRand.ClusterRand
	reversePlayChance   float64
}

func (pgen *SFParameterGenerator) ReceiveControlValue(value any, index int) {
	if index == 0 {
		pgen.offsetClustering.SetCenter(value.(float64))
		pgen.offsetClustering.Update()
	} else if index == 1 {
		pgen.onsetClustering.SetCenter(value.(float64))
		pgen.onsetClustering.Update()
	} else if index == 2 {
		pgen.durationClustering.SetCenter(value.(float64))
		pgen.durationClustering.Update()
	} else if index == 3 {
		pgen.speedClustering.SetCenter(value.(float64))
		pgen.speedClustering.Update()
	} else if index == 4 {
		pgen.panClustering.SetCenter(value.(float64))
		pgen.panClustering.Update()
	}
}

func (pgen *SFParameterGenerator) ReceiveMessage(msg any) []*muse.Message {
	return nil
}

func (pgen *SFParameterGenerator) Next(timestamp int64, config *muse.Configuration) (granular.Parameter, int64) {
	param := &pgen.parameter

	param.duration = pgen.durationClustering.Rand()
	param.amplitude = pgen.amplitudeClustering.Rand()
	param.speed = pgen.speedClustering.Rand()
	param.offset = pgen.offsetClustering.Rand()
	param.panning = pgen.panClustering.Rand()
	param.lookupMode = Wrap

	if rand.Float64() < pgen.reversePlayChance {
		param.speed *= -1.0
	}

	return param, int64(pgen.onsetClustering.Rand() * 0.01 * config.SampleRate)
}

func main() {
	sfb, err := sndfile.NewMipMapSoundFile("resources/sounds/elisa.wav", 5)
	if err != nil {
		log.Fatalf("fatal err: %v", err)
	}

	numChannels := sfb.NumChannels()

	root := muse.New(numChannels)

	paramGen := &SFParameterGenerator{}

	paramGen.speedClustering = museRand.NewClusterRand(1.3, 0.5, 1.0, 0.8, 0.8)
	paramGen.amplitudeClustering = museRand.NewClusterRand(0.6, 0.3, 0.3, 0.3, 0.3)
	paramGen.durationClustering = museRand.NewClusterRand(3200.0, 2155.0, 0.8, 0.7, 0.7)
	paramGen.offsetClustering = museRand.NewClusterRand(0.2, 0.03, 0.8, 0.8, 0.8)
	paramGen.onsetClustering = museRand.NewClusterRand(10.0, 6.5, 0.4, 0.8, 0.8)
	paramGen.panClustering = museRand.NewClusterRand(0.5, 0.3, 0.3, 0.2, 0.5)
	paramGen.reversePlayChance = 0.1

	gr := root.AddModule(granular.New(numChannels, &SFSourceFactory{SoundFile: sfb}, &trapezoidal.Factory{}, 400, paramGen))

	for i := 0; i < numChannels; i++ {
		gr.Connect(i, root, i)
	}

	root.RenderAudio()

	//env.SynthesizeToFile("/Users/almerlucke/Desktop/children.aiff", 180.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
}
