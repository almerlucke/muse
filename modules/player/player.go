package player

import (
	"math"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
)

type Player struct {
	*muse.BaseModule
	buffer  *io.SoundFileBuffer
	phase   float64
	inc     float64
	oneShot bool
	done    bool
}

func NewPlayer(buffer *io.SoundFileBuffer, speed float64, oneShot bool, config *muse.Configuration, identifier string) *Player {
	inc := (speed * buffer.SampleRate / config.SampleRate) / float64(buffer.NumFrames)
	if oneShot && inc < 0 {
		inc = math.Abs(inc)
	}

	return &Player{
		BaseModule: muse.NewBaseModule(0, len(buffer.Channels), config, identifier),
		inc:        inc,
		oneShot:    oneShot,
		done:       oneShot,
		buffer:     buffer,
	}
}

func (p *Player) ReceiveMessage(msg any) []*muse.Message {
	mp := msg.(map[string]any)
	speed := mp["speed"].(float64)

	p.inc = (speed * p.buffer.SampleRate / p.Config.SampleRate) / float64(p.buffer.NumFrames)

	if p.oneShot {
		p.done = false
		p.phase = 0.0
	}

	return nil
}

func (p *Player) IsActive() bool {
	return !p.oneShot || (p.oneShot && !p.done)
}

func (p *Player) Synthesize() bool {
	if !p.BaseModule.Synthesize() {
		return false
	}

	for i := 0; i < p.Config.BufferSize; i++ {
		if p.done {
			for _, out := range p.Outputs {
				out.Buffer[i] = 0.0
			}
			continue
		}

		x := p.phase * float64(p.buffer.NumFrames)
		xi1 := int64(x)
		xi2 := xi1 + 1
		xf := x - float64(xi1)

		if xi2 >= p.buffer.NumFrames {
			if p.oneShot {
				xi2 = p.buffer.NumFrames - 1
			} else {
				xi2 = 0
			}
		}

		for outIndex, out := range p.Outputs {
			xi1v := p.buffer.Channels[outIndex][xi1]
			out.Buffer[i] = xi1v + (p.buffer.Channels[outIndex][xi2]-xi1v)*xf
		}

		p.phase += p.inc

		if p.oneShot {
			if p.phase >= 1.0 {
				p.done = true
			}
		} else {
			for p.phase >= 1.0 {
				p.phase -= 1.0
			}
			for p.phase < 0.0 {
				p.phase += 1.0
			}
		}
	}

	return true
}
