package main

import (
	"math"
	"math/rand"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/generator"
	"github.com/almerlucke/muse/components/interpolator"
	"github.com/almerlucke/muse/components/iterator"
	"github.com/almerlucke/muse/components/iterator/chaos"
	"github.com/almerlucke/muse/components/osc"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/granular"
	"github.com/almerlucke/muse/modules/granular/envelopes/trapezoidal"
	museRand "github.com/almerlucke/muse/utils/rand"
)

type SFParam struct {
	duration      float64
	amplitude     float64
	chaos         float64
	interpolDelta float64
	waveform      int
	attack        float64
	release       float64
	freqLow       float64
	freqHigh      float64
	panStart      float64
	panEnd        float64
}

func (p *SFParam) Duration() float64 {
	return p.duration
}

func (p *SFParam) Amplitude() float64 {
	return p.amplitude
}

func (p *SFParam) Panning() float64 {
	return 0.5
}

func (p *SFParam) Smoothness() float64 {
	return 1.0
}

func (p *SFParam) Attack() float64 {
	return p.attack
}

func (p *SFParam) Release() float64 {
	return p.release
}

func NewSFParam(duration float64, amplitude float64, attack float64, release float64, panStart float64, panEnd float64, chaos float64, delta float64, waveform int, freqLow float64, freqhigh float64) *SFParam {
	return &SFParam{
		duration:      duration,
		amplitude:     amplitude,
		panStart:      panStart,
		panEnd:        panEnd,
		chaos:         chaos,
		interpolDelta: delta,
		attack:        attack,
		release:       release,
		freqLow:       freqLow,
		freqHigh:      freqhigh,
		waveform:      waveform,
	}
}

type SFSource struct {
	osc      *osc.Osc
	aron     *chaos.Aronson
	scale    *waveshaping.Linear
	gen      generator.Generator
	waveform int
	panCur   float64
	panInc   float64
}

func NewSource(sr float64) *SFSource {
	f := func(x float64) float64 { return x * x } // Aronson adjusted
	aron := chaos.NewAronsonWithFunc(1.698, f)
	iter := iterator.New([]float64{0.4, 0.4}, aron)
	mirror := waveshaping.NewMirror(-1.0, 1.0)
	uni := waveshaping.NewUnipolar()
	scale := waveshaping.NewLinear(1400.0, 50.0)
	serial := waveshaping.NewSerial(mirror, uni, scale)
	wrapper := interpolator.New(
		waveshaping.NewGeneratorWrapper(iter, []waveshaping.Shaper{serial, serial}),
		interpolator.Linear,
		250,
	)

	return &SFSource{
		osc:   osc.New(100.0, 0.0, sr),
		aron:  aron,
		scale: scale,
		gen:   wrapper,
	}
}

func (s *SFSource) Synthesize(outBuffers [][]float64, bufSize int) {
	pan := [2]float64{}
	waveformMap := [5]int{0, 0, 0, 3, 3}

	for i := 0; i < bufSize; i++ {
		pan[0] = math.Cos(s.panCur * math.Pi / 2.0)
		pan[1] = math.Sin(s.panCur * math.Pi / 2.0)

		s.panCur += s.panInc

		s.osc.SetFrequency(s.gen.Generate()[0])
		out := s.osc.Tick()[waveformMap[s.waveform]]

		for outIndex, outBuf := range outBuffers {
			outBuf[i] = pan[outIndex] * out
		}
	}
}

func (s *SFSource) Activate(sampsToGo int64, p granular.Parameter, c *muse.Configuration) {
	sfp := p.(*SFParam)
	s.aron.A = sfp.chaos
	s.scale.Shift = sfp.freqLow
	s.scale.Scale = sfp.freqHigh - sfp.freqLow
	s.panCur = sfp.panStart
	s.panInc = (sfp.panEnd - sfp.panStart) / float64(sampsToGo)
	s.waveform = sfp.waveform
	s.osc.SetMix([4]float64{0.25, 0.25, 0.25, 0.25})
	s.gen.(*interpolator.Interpolator).SetDelta(sfp.interpolDelta)
}

type SFSourceFactory struct {
	sr float64
}

func NewSourceFactory(sr float64) *SFSourceFactory {
	return &SFSourceFactory{sr: sr}
}

func (sf *SFSourceFactory) New() granular.Source {
	return NewSource(sf.sr)
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
	onsetClustering     *museRand.ClusterRand
	chaosClustering     *museRand.ClusterRand
	deltaClustering     *museRand.ClusterRand
	freqLowClustering   *museRand.ClusterRand
	freqHighClustering  *museRand.ClusterRand
	durationClustering  *museRand.ClusterRand
	waveformClustering  *museRand.ClusterRand
	panStartClustering  *museRand.ClusterRand
	panEndClustering    *museRand.ClusterRand
	attackClustering    *museRand.ClusterRand
	releaseClustering   *museRand.ClusterRand
}

func (pgen *SFParameterGenerator) ReceiveControlValue(value any, index int) {
	if index == 0 {
		pgen.chaosClustering.SetCenter(value.(float64))
		pgen.chaosClustering.Update()
	} else if index == 1 {
		pgen.freqLowClustering.SetCenter(value.(float64))
		pgen.freqLowClustering.Update()
	} else if index == 2 {
		pgen.freqHighClustering.SetCenter(value.(float64))
		pgen.freqHighClustering.Update()
	} else if index == 3 {
		pgen.durationClustering.SetCenter(value.(float64))
		pgen.durationClustering.Update()
	} else if index == 4 {
		pgen.onsetClustering.SetCenter(value.(float64))
		pgen.onsetClustering.Update()
	} else if index == 5 {
		pgen.panStartClustering.SetCenter(value.(float64))
		pgen.panStartClustering.Update()
	} else if index == 6 {
		pgen.panEndClustering.SetCenter(value.(float64))
		pgen.panEndClustering.Update()
	} else if index == 7 {
		pgen.deltaClustering.SetCenter(value.(float64))
		pgen.deltaClustering.Update()
	} else if index == 8 {
		pgen.waveformClustering.SetCenter(value.(float64))
		pgen.waveformClustering.Update()
	}
}

func (pgen *SFParameterGenerator) ReceiveMessage(msg any) []*muse.Message {
	return nil
}

func (pgen *SFParameterGenerator) Next(timestamp int64, config *muse.Configuration) (granular.Parameter, int64) {
	param := &pgen.parameter

	param.duration = pgen.durationClustering.Rand()
	param.amplitude = pgen.amplitudeClustering.Rand()
	param.chaos = pgen.chaosClustering.Rand()
	param.freqLow = pgen.freqLowClustering.Rand()
	param.freqHigh = pgen.freqHighClustering.Rand()
	param.panStart = pgen.panStartClustering.Rand()
	param.panEnd = pgen.panEndClustering.Rand()
	param.interpolDelta = pgen.deltaClustering.Rand()
	param.waveform = int(pgen.waveformClustering.Rand()) % 5
	param.attack = pgen.attackClustering.Rand()
	param.release = pgen.releaseClustering.Rand()

	return param, int64(pgen.onsetClustering.Rand() * 0.001 * config.SampleRate)
}

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0 * 3,
		BufferSize: 1024,
	})

	root := muse.New(2)

	paramGen := &SFParameterGenerator{}

	paramGen.chaosClustering = museRand.NewClusterRand(1.567, 0.232, 0.9, 1.0, 1.0)
	paramGen.amplitudeClustering = museRand.NewClusterRand(0.1, 0.07, 0.8, 0.8, 0.8)
	paramGen.durationClustering = museRand.NewClusterRand(1400.0, 1250.0, 0.9, 0.9, 0.9)
	paramGen.freqLowClustering = museRand.NewClusterRand(150, 100, 0.9, 0.9, 0.9)
	paramGen.freqHighClustering = museRand.NewClusterRand(2000, 1800, 0.9, 1.0, 1.0)
	paramGen.onsetClustering = museRand.NewClusterRand(380.0, 374.5, 0.9, 0.9, 0.9)
	paramGen.panStartClustering = museRand.NewClusterRand(0.5, 0.5, 1.0, 1.0, 1.0)
	paramGen.panEndClustering = museRand.NewClusterRand(0.5, 0.5, 1.0, 1.0, 1.0)
	paramGen.deltaClustering = museRand.NewClusterRand(0.006, 0.0049999, 0.0, 1.0, 1.0)
	paramGen.waveformClustering = museRand.NewClusterRand(3.0, 2.0, 0.9, 0.9, 0.9)
	paramGen.attackClustering = museRand.NewClusterRand(0.04, 0.03, 0.9, 1.0, 1.0)
	paramGen.releaseClustering = museRand.NewClusterRand(0.81, 0.12, 0.9, 1.0, 1.0)

	gr := root.AddModule(granular.New(2, NewSourceFactory(root.Config.SampleRate), &trapezoidal.Factory{}, 40, paramGen))

	chaosLfo := root.AddControl(lfo.NewBasicControlLFO(0.0721, 1.36, 1.767))
	freqLowLfo := root.AddControl(lfo.NewBasicControlLFO(0.0821, 150.0, 300.0))
	freqHighLfo := root.AddControl(lfo.NewBasicControlLFO(0.0621, 800, 2000))
	deltaLfo := root.AddControl(lfo.NewBasicControlLFO(0.0521, 0.01, 0.0001))

	gr.CtrlIn(chaosLfo, freqLowLfo, freqHighLfo, deltaLfo, 0, 7)

	// offsetLfo := env.AddControl(lfo.NewBasicControlLFO(0.04, 0.1, 0.9, env.Config, ""))
	// offsetLfo.CtrlConnect(0, gr, 0)

	// onsetLfo := env.AddControl(lfo.NewBasicControlLFO(0.031, 5.2, 40.8, env.Config, ""))
	// onsetLfo.CtrlConnect(0, gr, 1)

	// durationLfo := env.AddControl(lfo.NewBasicControlLFO(0.021, 135.2, 450.8, env.Config, ""))
	// durationLfo.CtrlConnect(0, gr, 2)

	// panLfo := env.AddControl(lfo.NewBasicControlLFO(0.011, 0.3, 0.7, env.Config, ""))
	// panLfo.CtrlConnect(0, gr, 4)

	for i := 0; i < 2; i++ {
		gr.Connect(i, root, i)
	}

	root.RenderToSoundFile("/Users/almerlucke/Desktop/chaosping.aiff", 30.0, 44100.0, io.WAV)

	// root.RenderToSoundFile("/Users/almerlucke/Desktop/chaosping.aiff", 180.0, root.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
	// root.RenderAudio()
}
