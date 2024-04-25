package main

import (
	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/genny/float/interp"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/mirror"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"github.com/almerlucke/sndfile/writer"
	"math"
	"math/rand"

	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/osc"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/granular"
	grainAdsr "github.com/almerlucke/muse/modules/granular/envelopes/adsr"
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
	adsrSetting   *adsr.Setting
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

func (p *SFParam) ADSRSetting() *adsr.Setting {
	return p.adsrSetting
}

func (p *SFParam) Release() float64 {
	return p.release
}

type SFSource struct {
	osc      *osc.Osc
	aron     *chaos.Aronson
	scale    *linear.Linear
	gen      float.FrameGenerator
	waveform int
	panCur   float64
	panInc   float64
}

func NewSource(sr float64) *SFSource {
	f := func(x float64) float64 { return x * x } // Aronson adjusted
	aron := chaos.NewAronsonWithFunc(1.698, f)
	iter := iter.New([]float64{0.4, 0.4}, aron)
	mirror := mirror.New(-1.0, 1.0)
	uni := linear.NewUnipolar()
	scale := linear.New(1400.0, 50.0)
	series := series.New(mirror, uni, scale)
	wrapper := interp.New(
		shape.New(iter, series),
		interp.Linear,
		1.0/250.0,
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
	s.gen.(*interp.Interpolator).SetDelta(sfp.interpolDelta)
}

func (s *SFSource) New(cfg any) granular.Source {
	return NewSource(cfg.(*muse.Configuration).SampleRate)
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
	} else if index == 9 {
		pgen.onsetClustering.SetCenter(value.(float64))
		pgen.onsetClustering.Update()
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
	//param.adsrSetting.AttackDuration = param.attack
	//param.adsrSetting.ReleaseDuration = param.release

	return param, int64(pgen.onsetClustering.Rand() * 0.001 * config.SampleRate)
}

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 1024,
	})

	root := muse.New(2)

	paramGen := &SFParameterGenerator{}

	paramGen.parameter.adsrSetting = adsr.NewSetting(1.0, 0.05, 0.25, 0.05, 0.1, 0.6)
	paramGen.parameter.adsrSetting.AttackShape = 0.5
	paramGen.parameter.adsrSetting.DecayShape = -0.5
	paramGen.parameter.adsrSetting.ReleaseShape = -0.5
	paramGen.chaosClustering = museRand.NewClusterRand(1.567, 0.232, 0.9, 1.0, 1.0)
	paramGen.amplitudeClustering = museRand.NewClusterRand(0.1, 0.07, 0.8, 0.8, 0.8)
	paramGen.durationClustering = museRand.NewClusterRand(1400.0, 150.0, 0.9, 0.9, 0.9)
	paramGen.freqLowClustering = museRand.NewClusterRand(150, 100, 0.9, 0.9, 0.9)
	paramGen.freqHighClustering = museRand.NewClusterRand(2000, 1800, 0.9, 1.0, 1.0)
	paramGen.onsetClustering = museRand.NewClusterRand(380.0, 14.5, 0.9, 0.9, 0.9)
	paramGen.panStartClustering = museRand.NewClusterRand(0.5, 0.5, 1.0, 1.0, 1.0)
	paramGen.panEndClustering = museRand.NewClusterRand(0.5, 0.5, 1.0, 1.0, 1.0)
	paramGen.deltaClustering = museRand.NewClusterRand(0.006, 0.0049999, 0.0, 1.0, 1.0)
	paramGen.waveformClustering = museRand.NewClusterRand(3.0, 2.0, 0.9, 0.9, 0.9)
	paramGen.attackClustering = museRand.NewClusterRand(0.04, 0.03, 0.9, 1.0, 1.0)
	paramGen.releaseClustering = museRand.NewClusterRand(0.81, 0.12, 0.9, 1.0, 1.0)

	gr := root.AddModule(granular.New(2, &SFSource{}, &grainAdsr.Envelope{}, 100, paramGen))

	chaosLfo := root.AddControl(lfo.NewBasicControlLFO(0.0721, 1.36, 1.767))
	freqLowLfo := root.AddControl(lfo.NewBasicControlLFO(0.0821, 150.0, 300.0))
	freqHighLfo := root.AddControl(lfo.NewBasicControlLFO(0.0621, 800, 2000))
	deltaLfo := root.AddControl(lfo.NewBasicControlLFO(0.0521, 0.01, 0.0001))
	onsetLfo := root.AddControl(lfo.NewBasicControlLFO(0.0321, 15.01, 300.0001))
	durLfo := root.AddControl(lfo.NewBasicControlLFO(0.0421, 200.01, 2400.0001))

	gr.CtrlIn(chaosLfo, freqLowLfo, freqHighLfo, durLfo, deltaLfo, 0, 7, onsetLfo, 0, 9)

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

	_ = root.RenderToSoundFile("/home/almer/Documents/chaosNew", writer.AIFC, 300, muse.SampleRate(), true)

	// _ = root.RenderToSoundFile("/home/almer/Documents/chaosping", writer.AIFC, 30.0, 44100.0, true)

	// root.RenderToSoundFile("/Users/almerlucke/Desktop/chaosping.aiff", 180.0, root.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
	// _ = root.RenderAudio()
}
