package granular

import (
	"math"

	"github.com/almerlucke/muse"
)

type Parameter interface {
	// Duration in seconds
	Duration() float64
	// Amplitude
	Amplitude() float64
	// Stereo field panning
	Panning() float64
	// Envelope type
	EnvType() EnvelopeType
	// Attack relative portion (only if env type is not parabolice)
	Attack() float64
	// Release relative portion (only if env type is not parabolice)
	Release() float64
}

type Step[P Parameter] struct {
	Parameter  P
	InterOnset int
}

type Sequence[P Parameter] interface {
	NextStep(int64, *muse.Configuration) *Step[P]
}

type SourceFactory[P Parameter] interface {
	NewSource() Source[P]
}

type Source[P Parameter] interface {
	Synthesize() float64
	Activate(P, *muse.Configuration)
}

type grain[P Parameter] struct {
	envelope  Envelope
	source    Source[P]
	sampsToGo int64
	panLeft   float64
	panRight  float64
}

func (g *grain[P]) activate(p P, config *muse.Configuration) {
	pan := p.Panning()

	g.panLeft = math.Cos(pan * math.Pi / 2.0)
	g.panRight = math.Sin(pan * math.Pi / 2.0)
	g.sampsToGo = int64(p.Duration() * config.SampleRate)

	g.source.Activate(p, config)

	g.envelope.Activate(EnvelopeConfiguration{
		Amplitude:       p.Amplitude(),
		Attack:          p.Attack(),
		Release:         p.Release(),
		DurationSamples: g.sampsToGo,
		Type:            p.EnvType(),
	})
}

func (g *grain[P]) synthesize(out [][]float64) {
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

type grainPoolElement[P Parameter] struct {
	grain *grain[P]
	prev  *grainPoolElement[P]
	next  *grainPoolElement[P]
}

func (e *grainPoolElement[P]) Unlink() {
	e.prev.next = e.next
	e.next.prev = e.prev
}

type grainPool[P Parameter] struct {
	sentinel *grainPoolElement[P]
}

func (gp *grainPool[P]) Initialize() {
	sentinel := &grainPoolElement[P]{}
	sentinel.next = sentinel
	sentinel.prev = sentinel
	gp.sentinel = sentinel
}

func (gp *grainPool[P]) Pop() *grainPoolElement[P] {
	first := gp.sentinel.next

	if first == gp.sentinel {
		return nil
	}

	first.Unlink()

	return first
}

func (gp *grainPool[P]) Push(e *grainPoolElement[P]) {
	e.next = gp.sentinel.next
	e.prev = gp.sentinel
	gp.sentinel.next.prev = e
	gp.sentinel.next = e
}

type Granulator[P Parameter] struct {
	*muse.BaseModule
	freeGrains      grainPool[P]
	activeGrains    grainPool[P]
	activatedGrains grainPool[P]
	sequence        Sequence[P]
	nextStep        *Step[P]
	timestamp       int64
}

func (gl *Granulator[P]) Initialize(identifier string, sf SourceFactory[P], grainPoolSize int, sequence Sequence[P], config *muse.Configuration) {
	gl.BaseModule = muse.NewBaseModule(0, 2, identifier)

	gl.freeGrains.Initialize()
	gl.activeGrains.Initialize()
	gl.activatedGrains.Initialize()
	gl.sequence = sequence

	for i := 0; i < grainPoolSize; i++ {
		g := &grain[P]{}
		g.source = sf.NewSource()
		e := &grainPoolElement[P]{grain: g}
		gl.freeGrains.Push(e)
	}

	gl.nextStep = sequence.NextStep(0, config)
}

func (gl *Granulator[P]) synthesizePool(p *grainPool[P], out [][]float64) {
	e := p.sentinel.next

	for e != p.sentinel {
		g := e.grain
		g.synthesize(out)

		prev := e
		e = e.next

		if g.sampsToGo == 0 {
			prev.Unlink()
			gl.freeGrains.Push(prev)
		}
	}
}

func (gl *Granulator[P]) moveActivated() {
	e := gl.activatedGrains.sentinel.next
	for e != gl.activatedGrains.sentinel {
		prev := e
		e = e.next
		prev.Unlink()
		gl.activeGrains.Push(prev)
	}
}

func (gl *Granulator[P]) Synthesize(config *muse.Configuration) bool {
	if !gl.BaseModule.Synthesize(config) {
		return false
	}

	var out [][]float64

	if gl.NumOutputs() == 1 {
		out = [][]float64{gl.OutputAtIndex(0).Buffer}
		for i := 0; i < config.BufferSize; i++ {
			out[0][i] = 0.0
		}
	} else {
		out = [][]float64{gl.OutputAtIndex(0).Buffer, gl.OutputAtIndex(1).Buffer}
		for i := 0; i < config.BufferSize; i++ {
			out[0][i] = 0.0
			out[1][i] = 0.0
		}
	}

	outIndex := 0

	// First run all currently active grains
	gl.synthesizePool(&gl.activeGrains, out)

	// Step through inter onsets in current cycle
	done := false
	for !done {
		sampsToGenerate := gl.nextStep.InterOnset
		sampsLeft := config.BufferSize - outIndex

		if sampsToGenerate > sampsLeft {
			sampsToGenerate = sampsLeft
			done = true
		}

		partialOut := make([][]float64, gl.NumOutputs())

		for bufIndex, buf := range out {
			partialOut[bufIndex] = buf[outIndex : outIndex+sampsToGenerate]
		}

		gl.synthesizePool(&gl.activatedGrains, partialOut)

		gl.nextStep.InterOnset -= sampsToGenerate
		outIndex += sampsToGenerate

		if gl.nextStep.InterOnset == 0 {
			e := gl.freeGrains.Pop()
			if e != nil {
				gl.activatedGrains.Push(e)
				e.grain.activate(gl.nextStep.Parameter, config)
			}

			gl.nextStep = gl.sequence.NextStep(gl.timestamp, config)
		}
	}

	// Move activated grains to active pool
	gl.moveActivated()

	gl.timestamp += int64(config.BufferSize)

	return true
}
