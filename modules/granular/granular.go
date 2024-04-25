package granular

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils"
	"github.com/almerlucke/muse/utils/pool"
)

type Parameter interface {
	// Duration in milliseconds
	Duration() float64
	Amplitude() float64
}

type ParameterGenerator interface {
	muse.MessageReceiver
	muse.ControlReceiver
	Next(int64, *muse.Configuration) (Parameter, int64)
}

type Envelope interface {
	Synthesize([][]float64, int)
	Activate(float64, int64, Parameter, *muse.Configuration)
}

type Source interface {
	Synthesize([][]float64, int)
	Activate(int64, Parameter, *muse.Configuration)
}

type grain struct {
	envelope  Envelope
	source    Source
	sampsToGo int64
}

func (g *grain) activate(p Parameter, config *muse.Configuration) {
	g.sampsToGo = int64(p.Duration() * 0.001 * config.SampleRate)
	g.source.Activate(g.sampsToGo, p, config)
	g.envelope.Activate(p.Amplitude(), g.sampsToGo, p, config)
}

func (g *grain) synthesize(sourceBufs [][]float64, bufSize int) int {
	sampsToSynthesize := g.sampsToGo
	if sampsToSynthesize > int64(bufSize) {
		sampsToSynthesize = int64(bufSize)
	}

	g.sampsToGo -= sampsToSynthesize

	n := int(sampsToSynthesize)

	g.source.Synthesize(sourceBufs, n)
	g.envelope.Synthesize(sourceBufs, n)

	return n
}

type Granulator struct {
	*muse.BaseModule
	freeGrains    *pool.Pool[*grain]
	activeGrains  *pool.Pool[*grain]
	paramGen      ParameterGenerator
	nextParameter Parameter
	interOnset    int64
	timestamp     int64
	sourceBufs    [][]float64
	outBufs       [][]float64
}

func New(numOutputs int, sf utils.Factory[Source], ef utils.Factory[Envelope], grainPoolSize int, paramGen ParameterGenerator) *Granulator {
	config := muse.CurrentConfiguration()

	gl := &Granulator{
		BaseModule:   muse.NewBaseModule(0, numOutputs),
		freeGrains:   pool.NewPool[*grain](),
		activeGrains: pool.NewPool[*grain](),
		paramGen:     paramGen,
		sourceBufs:   make([][]float64, numOutputs), // synthesize buffer for grain source
		outBufs:      make([][]float64, numOutputs), // output buffers
	}

	for i := 0; i < numOutputs; i++ {
		gl.sourceBufs[i] = make([]float64, muse.BufferSize())
		gl.outBufs[i] = gl.Outputs[i].Buffer
	}

	for i := 0; i < grainPoolSize; i++ {
		g := &grain{}
		g.source = sf.New(config)
		g.envelope = ef.New(config)
		e := &pool.Element[*grain]{Value: g}
		gl.freeGrains.Push(e)
	}

	gl.nextParameter, gl.interOnset = paramGen.Next(0, config)

	gl.SetSelf(gl)

	return gl
}

func (gl *Granulator) ReceiveControlValue(value any, index int) {
	gl.paramGen.ReceiveControlValue(value, index)
}

func (gl *Granulator) ReceiveMessage(msg any) []*muse.Message {
	return gl.paramGen.ReceiveMessage(msg)
}

func (gl *Granulator) synthesizePool(p *pool.Pool[*grain], out [][]float64, bufSize int) {
	e := p.First()
	end := p.End()

	for e != end {
		g := e.Value

		n := g.synthesize(gl.sourceBufs, bufSize)

		for outIndex, outBuf := range out {
			for i := 0; i < n; i++ {
				outBuf[i] += gl.sourceBufs[outIndex][i]
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

func (gl *Granulator) Synthesize() bool {
	if !gl.BaseModule.Synthesize() {
		return false
	}

	// Zero buffers for new frame
	for _, outBuf := range gl.outBufs {
		for i := 0; i < gl.Config.BufferSize; i++ {
			outBuf[i] = 0.0
		}
	}

	// First run all currently active grains for full buffer size
	gl.synthesizePool(gl.activeGrains, gl.outBufs, gl.Config.BufferSize)

	outIndex := 0
	sampsLeft := int64(gl.Config.BufferSize)

	// Step through inter onsets in current cycle
	for {
		// If interonset is greater then samps left, decrement interonset and break
		if gl.interOnset > sampsLeft {
			gl.interOnset -= sampsLeft
			break
		}

		// Skip interonset samples
		sampsLeft -= gl.interOnset
		outIndex += int(gl.interOnset)

		gl.interOnset = 0

		// While interonset == 0 generate new grains and synthesize for remaining samples in this frame
		for gl.interOnset == 0 {
			e := gl.freeGrains.Pop()
			if e != nil {
				e.Value.activate(gl.nextParameter, gl.Config)

				// Synthesize remaining samples in this frame
				n := e.Value.synthesize(gl.sourceBufs, int(sampsLeft))

				for outBufIndex, outBuf := range gl.outBufs {
					for i := 0; i < n; i++ {
						outBuf[outIndex+i] += gl.sourceBufs[outBufIndex][i]
					}
				}

				// If grain is done, put it back in free list, otherwise keep it for next frame in active grains
				if e.Value.sampsToGo == 0 {
					gl.freeGrains.Push(e)
				} else {
					gl.activeGrains.Push(e)
				}
			}

			gl.nextParameter, gl.interOnset = gl.paramGen.Next(gl.timestamp, gl.Config)
		}
	}

	// Update timestamp
	gl.timestamp += int64(gl.Config.BufferSize)

	return true
}
