package granular

import (
	"container/list"
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
	NextStep(*muse.Configuration) *Step[P]
}

type SourceFactory[P Parameter] interface {
	NewSource() Source[P]
}

type Source[P Parameter] interface {
	Synthesize() float64
	Activate(P, *muse.Configuration)
}

type Grain[P Parameter] struct {
	Envelope  Envelope
	Source    Source[P]
	SampsToGo int64
	panLeft   float64
	panRight  float64
}

func (g *Grain[P]) Activate(p P, config *muse.Configuration) {
	pan := p.Panning()

	g.panLeft = math.Cos(pan * math.Pi / 2.0)
	g.panRight = math.Sin(pan * math.Pi / 2.0)
	g.SampsToGo = int64(p.Duration() * config.SampleRate)

	g.Source.Activate(p, config)

	g.Envelope.Activate(EnvelopeConfiguration{
		Amplitude:       p.Amplitude(),
		Attack:          p.Attack(),
		Release:         p.Release(),
		DurationSamples: g.SampsToGo,
		Type:            p.EnvType(),
	})
}

func (g *Grain[P]) Synthesize(out [][]float64) {
	bufSize := int64(len(out[0]))
	sampsToSynthesize := g.SampsToGo
	if sampsToSynthesize > bufSize {
		sampsToSynthesize = bufSize
	}

	g.SampsToGo -= sampsToSynthesize

	if len(out) == 2 {
		for i := int64(0); i < sampsToSynthesize; i++ {
			samp := g.Source.Synthesize() * g.Envelope.Synthesize()
			out[0][i] += samp * g.panLeft
			out[1][i] += samp * g.panRight
		}
	} else {
		for i := int64(0); i < sampsToSynthesize; i++ {
			out[0][i] += g.Source.Synthesize() * g.Envelope.Synthesize()
		}
	}
}

type Granulator[P Parameter] struct {
	*muse.BaseModule
	freeGrains      *list.List
	activeGrains    *list.List
	activatedGrains *list.List
	sequence        Sequence[P]
	nextStep        *Step[P]
}

func (gl *Granulator[P]) Initialize(identifier string, sf SourceFactory[P], grainPoolSize int, sequence Sequence[P], config *muse.Configuration) {
	gl.BaseModule = muse.NewBaseModule(0, 2, identifier)

	gl.freeGrains = list.New()
	gl.activeGrains = list.New()
	gl.activatedGrains = list.New()
	gl.sequence = sequence

	for i := 0; i < grainPoolSize; i++ {
		g := &Grain[P]{}
		g.Source = sf.NewSource()
		gl.freeGrains.PushBack(g)
	}

	gl.nextStep = sequence.NextStep(config)
}

func (gl *Granulator[P]) synthesizeList(l *list.List, out [][]float64) {
	e := l.Front()
	for e != nil {
		g := e.Value.(*Grain[P])
		g.Synthesize(out)

		prev := e
		e = e.Next()

		if g.SampsToGo == 0 {
			l.Remove(prev)
			gl.freeGrains.PushFront(prev.Value)
		}
	}
}

func (gl *Granulator[P]) moveActivated() {
	e := gl.activatedGrains.Front()
	for e != nil {
		prev := e
		e = e.Next()
		gl.activatedGrains.Remove(prev)
		gl.activeGrains.PushBack(prev.Value)
	}
}

func (gl *Granulator[P]) Synthesize(config *muse.Configuration) bool {
	if !gl.BaseModule.Synthesize(config) {
		return false
	}

	out1 := gl.OutputAtIndex(0).Buffer
	out2 := gl.OutputAtIndex(1).Buffer
	out := [][]float64{out1, out2}

	for i := 0; i < config.BufferSize; i++ {
		out1[i] = 0.0
		out2[i] = 0.0
	}

	outIndex := 0

	// First run all currently active grains
	gl.synthesizeList(gl.activeGrains, out)

	// Step through inter onsets in current cycle
	done := false
	for !done {
		sampsToGenerate := gl.nextStep.InterOnset
		sampsLeft := config.BufferSize - outIndex

		if sampsToGenerate > sampsLeft {
			sampsToGenerate = sampsLeft
			done = true
		}

		partialOut := [][]float64{nil, nil}

		for bufIndex, buf := range out {
			partialOut[bufIndex] = buf[outIndex : outIndex+sampsToGenerate]
		}

		gl.synthesizeList(gl.activatedGrains, partialOut)

		gl.nextStep.InterOnset -= sampsToGenerate
		outIndex += sampsToGenerate

		if gl.nextStep.InterOnset == 0 {
			e := gl.freeGrains.Front()
			if e != nil {
				gl.freeGrains.Remove(e)
				gl.activatedGrains.PushFront(e.Value)
				g := e.Value.(*Grain[P])
				g.Activate(gl.nextStep.Parameter, config)
			}

			gl.nextStep = gl.sequence.NextStep(config)
		}
	}

	gl.moveActivated()

	return true
}
