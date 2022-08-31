package granular

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/pool"
)

type Parameter interface {
	// Duration in milliseconds
	Duration() float64
	// Amplitude
	Amplitude() float64
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
	Synthesize([][]float64, int)
	Activate(Parameter, *muse.Configuration)
}

type grain struct {
	envelope  Envelope
	source    Source
	sampsToGo int64
}

func (g *grain) activate(p Parameter, config *muse.Configuration) {
	// pan := p.Panning()
	// g.panLeft = math.Cos(pan * math.Pi / 2.0)
	// g.panRight = math.Sin(pan * math.Pi / 2.0)

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

func (g *grain) synthesize(outBuf [][]float64, envBuf []float64, bufSize int) {
	numOut := len(outBuf)

	sampsToSynthesize := g.sampsToGo
	if sampsToSynthesize > int64(bufSize) {
		sampsToSynthesize = int64(bufSize)
	}

	g.sampsToGo -= sampsToSynthesize

	n := int(sampsToSynthesize)

	g.source.Synthesize(outBuf, n)
	g.envelope.SynthesizeBuffer(envBuf, n)

	for outIndex := 0; outIndex < numOut; outIndex++ {
		for i := 0; i < n; i++ {
			outBuf[outIndex][i] *= envBuf[i]
		}
	}
}

type Granulator struct {
	*muse.BaseModule
	freeGrains        *pool.Pool[*grain]
	activeGrains      *pool.Pool[*grain]
	activatedGrains   *pool.Pool[*grain]
	parameterFactory  ParameterFactory
	nextParameter     Parameter
	interOnset        int64
	timestamp         int64
	synthesizeBuffer  [][]float64
	envelopeBuffer    []float64
	outBuffers        [][]float64
	partialOutBuffers [][]float64
}

func NewGranulator(numOutputs int, sf SourceFactory, grainPoolSize int, parameterFactory ParameterFactory, config *muse.Configuration, identifier string) *Granulator {
	gl := &Granulator{
		BaseModule:        muse.NewBaseModule(0, numOutputs, config, identifier),
		freeGrains:        pool.NewPool[*grain](),
		activeGrains:      pool.NewPool[*grain](),
		activatedGrains:   pool.NewPool[*grain](),
		parameterFactory:  parameterFactory,
		synthesizeBuffer:  make([][]float64, numOutputs),
		envelopeBuffer:    make([]float64, config.BufferSize),
		outBuffers:        make([][]float64, numOutputs),
		partialOutBuffers: make([][]float64, numOutputs),
	}

	for i := 0; i < numOutputs; i++ {
		gl.synthesizeBuffer[i] = make([]float64, config.BufferSize)
		gl.outBuffers[i] = gl.Outputs[i].Buffer
	}

	for i := 0; i < grainPoolSize; i++ {
		g := &grain{}
		g.source = sf.NewSource()
		e := &pool.Element[*grain]{Value: g}
		gl.freeGrains.Push(e)
	}

	gl.nextParameter, gl.interOnset = parameterFactory.NextParameter(0, config)

	return gl
}

func (gl *Granulator) synthesizePool(p *pool.Pool[*grain], out [][]float64, bufSize int) {
	e := p.First()
	end := p.End()

	for e != end {
		g := e.Value

		g.synthesize(gl.synthesizeBuffer, gl.envelopeBuffer, bufSize)

		for outIndex, outBuf := range out {
			for i := 0; i < bufSize; i++ {
				outBuf[i] += gl.synthesizeBuffer[outIndex][i]
			}
		}

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

	for _, outBuf := range gl.outBuffers {
		for i := 0; i < gl.Config.BufferSize; i++ {
			outBuf[i] = 0.0
		}
	}

	outIndex := 0

	// First run all currently active grains
	gl.synthesizePool(gl.activeGrains, gl.outBuffers, gl.Config.BufferSize)

	// Step through inter onsets in current cycle
	done := false
	for !done {
		sampsToGenerate := gl.interOnset
		sampsLeft := int64(gl.Config.BufferSize - outIndex)

		if sampsToGenerate > sampsLeft {
			sampsToGenerate = sampsLeft
			done = true
		}

		for bufIndex, buf := range gl.outBuffers {
			gl.partialOutBuffers[bufIndex] = buf[outIndex : outIndex+int(sampsToGenerate)]
		}

		gl.synthesizePool(gl.activatedGrains, gl.partialOutBuffers, int(sampsToGenerate))

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
