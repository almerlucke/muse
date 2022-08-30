package granular

import (
	"math"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/pool"
)

type Parameter interface {
	// Duration in milliseconds
	Duration() float64
	// Amplitude
	Amplitude() float64
	// Stereo field panning
	Panning() float64
	// Envelope type
	EnvType() EnvelopeType
	// Attack relative portion (only if env type is not parabolic)
	Attack() float64
	// Release relative portion (only if env type is not parabolic)
	Release() float64
}

type ParameterFactory interface {
	NextParameter(int64, *muse.Configuration) (Parameter, int64)
}

type SourceFactory interface {
	NewSource() Source
}

type Source interface {
	Synthesize() float64
	Activate(Parameter, *muse.Configuration)
}

type grain struct {
	envelope  Envelope
	source    Source
	sampsToGo int64
	panLeft   float64
	panRight  float64
}

func (g *grain) activate(p Parameter, config *muse.Configuration) {
	pan := p.Panning()

	g.panLeft = math.Cos(pan * math.Pi / 2.0)
	g.panRight = math.Sin(pan * math.Pi / 2.0)
	g.sampsToGo = int64(p.Duration() * 0.001 * config.SampleRate)

	g.source.Activate(p, config)

	g.envelope.Activate(EnvelopeConfiguration{
		Amplitude:       p.Amplitude(),
		Attack:          p.Attack(),
		Release:         p.Release(),
		DurationSamples: g.sampsToGo,
		Type:            p.EnvType(),
	})
}

func (g *grain) synthesize(out [][]float64) {
	bufSize := int64(len(out[0]))
	sampsToSynthesize := g.sampsToGo
	if sampsToSynthesize > bufSize {
		sampsToSynthesize = bufSize
	}

	g.sampsToGo -= sampsToSynthesize

	if len(out) == 2 {
		for i := int64(0); i < sampsToSynthesize; i++ {
			samp := g.source.Synthesize() * g.envelope.Synthesize()
			out[0][i] += samp * g.panLeft
			out[1][i] += samp * g.panRight
		}
	} else {
		for i := int64(0); i < sampsToSynthesize; i++ {
			out[0][i] += g.source.Synthesize() * g.envelope.Synthesize()
		}
	}
}

type Granulator struct {
	*muse.BaseModule
	freeGrains       pool.Pool[*grain]
	activeGrains     pool.Pool[*grain]
	activatedGrains  pool.Pool[*grain]
	parameterFactory ParameterFactory
	nextParameter    Parameter
	interOnset       int64
	timestamp        int64
}

func NewGranulator(sf SourceFactory, grainPoolSize int, parameterFactory ParameterFactory, config *muse.Configuration, identifier string) *Granulator {
	grl := &Granulator{}
	grl.Initialize(sf, grainPoolSize, parameterFactory, config, identifier)
	return grl
}

func (gl *Granulator) Initialize(sf SourceFactory, grainPoolSize int, parameterFactory ParameterFactory, config *muse.Configuration, identifier string) {
	gl.BaseModule = muse.NewBaseModule(0, 2, config, identifier)

	gl.freeGrains.Initialize()
	gl.activeGrains.Initialize()
	gl.activatedGrains.Initialize()
	gl.parameterFactory = parameterFactory

	for i := 0; i < grainPoolSize; i++ {
		g := &grain{}
		g.source = sf.NewSource()
		e := &pool.Element[*grain]{Value: g}
		gl.freeGrains.Push(e)
	}

	gl.nextParameter, gl.interOnset = parameterFactory.NextParameter(0, config)
}

func (gl *Granulator) synthesizePool(p *pool.Pool[*grain], out [][]float64) {
	e := p.First()
	end := p.End()

	for e != end {
		g := e.Value
		g.synthesize(out)

		prev := e
		e = e.Next

		if g.sampsToGo == 0 {
			prev.Unlink()
			gl.freeGrains.Push(prev)
		}
	}
}

func (gl *Granulator) moveActivated() {
	e := gl.activatedGrains.First()
	end := gl.activatedGrains.End()
	for e != end {
		prev := e
		e = e.Next
		prev.Unlink()
		gl.activeGrains.Push(prev)
	}
}

func (gl *Granulator) Synthesize() bool {
	if !gl.BaseModule.Synthesize() {
		return false
	}

	var out [2][]float64
	var partialOut [2][]float64

	numOutputs := gl.NumOutputs()

	if numOutputs == 1 {
		out[0] = gl.OutputAtIndex(0).Buffer
		for i := 0; i < gl.Config.BufferSize; i++ {
			out[0][i] = 0.0
		}
	} else {
		out[0] = gl.OutputAtIndex(0).Buffer
		out[1] = gl.OutputAtIndex(1).Buffer
		for i := 0; i < gl.Config.BufferSize; i++ {
			out[0][i] = 0.0
			out[1][i] = 0.0
		}
	}

	outIndex := 0

	// First run all currently active grains
	gl.synthesizePool(&gl.activeGrains, out[:numOutputs])

	// Step through inter onsets in current cycle
	done := false
	for !done {
		sampsToGenerate := gl.interOnset
		sampsLeft := int64(gl.Config.BufferSize - outIndex)

		if sampsToGenerate > sampsLeft {
			sampsToGenerate = sampsLeft
			done = true
		}

		for bufIndex, buf := range out {
			partialOut[bufIndex] = buf[outIndex : outIndex+int(sampsToGenerate)]
		}

		gl.synthesizePool(&gl.activatedGrains, partialOut[:numOutputs])

		gl.interOnset -= sampsToGenerate
		outIndex += int(sampsToGenerate)

		for gl.interOnset == 0 {
			e := gl.freeGrains.Pop()
			if e != nil {
				gl.activatedGrains.Push(e)
				e.Value.activate(gl.nextParameter, gl.Config)
			}

			gl.nextParameter, gl.interOnset = gl.parameterFactory.NextParameter(gl.timestamp, gl.Config)
		}
	}

	// Move activated grains to active pool
	gl.moveActivated()

	gl.timestamp += int64(gl.Config.BufferSize)

	return true
}
