package player

import (
	"math"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/dsp/mipmap"
)

type MMPlayer struct {
	*muse.BaseModule
	mmBuf   *mipmap.MipMapSoundFileBuffer
	phase   float64
	inc     float64
	speed   float64
	depth   int
	amp     float64
	oneShot bool
	done    bool
}

func SpeedToDepth(speed float64) int {
	speed = math.Abs(speed)

	whole, fract := math.Modf(speed)

	depth := int(whole)
	if fract < 0.0001 && depth > 0 {
		depth -= 1
	}

	return depth
}

func NewMMPlayer(mmBuf *mipmap.MipMapSoundFileBuffer, speed float64, amp float64, oneShot bool, config *muse.Configuration, identifier string) *MMPlayer {
	inc := (speed * mmBuf.SampleRate / config.SampleRate) / float64(mmBuf.NumFrames)
	depth := SpeedToDepth(speed)

	if depth >= mmBuf.Depth {
		depth = mmBuf.Depth - 1
	}

	if oneShot {
		inc = math.Abs(inc)
	}

	p := &MMPlayer{
		BaseModule: muse.NewBaseModule(0, len(mmBuf.Channels), config, identifier),
		inc:        inc,
		speed:      speed,
		oneShot:    oneShot,
		done:       oneShot,
		mmBuf:      mmBuf,
		depth:      depth,
		amp:        amp,
	}

	p.SetSelf(p)

	return p
}

func (p *MMPlayer) SetBuffer(mmBuf *mipmap.MipMapSoundFileBuffer) {
	p.mmBuf = mmBuf
	p.inc = (p.speed * p.mmBuf.SampleRate / p.Config.SampleRate) / float64(p.mmBuf.NumFrames)
}

func (p *MMPlayer) Speed() float64 {
	return p.speed
}

func (p *MMPlayer) SetSpeed(speed float64) {
	p.inc = (speed * p.mmBuf.SampleRate / p.Config.SampleRate) / float64(p.mmBuf.NumFrames)
	p.speed = speed
	depth := SpeedToDepth(speed)

	if depth >= p.mmBuf.Depth {
		depth = p.mmBuf.Depth - 1
	}

	p.depth = depth
}

func (p *MMPlayer) Bang() {
	if p.oneShot {
		p.done = false
		p.phase = 0.0
	}
}

func (p *MMPlayer) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Bang
		if value == muse.Bang {
			p.Bang()
		}
	case 1: // Speed
		p.SetSpeed(value.(float64))
	}
}

func (p *MMPlayer) ReceiveMessage(msg any) []*muse.Message {
	if content, ok := msg.(map[string]any); ok {
		if speed, ok := content["speed"]; ok {
			p.SetSpeed(speed.(float64))
		}
	}

	if muse.IsBang(msg) {
		p.Bang()
	}

	return nil
}

func (p *MMPlayer) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	p.activate(amplitude, message, config)
}

func (p *MMPlayer) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
	p.activate(amplitude, message, config)
}

func (p *MMPlayer) NoteOff() {}

func (p *MMPlayer) activate(amplitude float64, message any, config *muse.Configuration) {
	content := message.(map[string]any)

	p.amp = amplitude

	speed := p.speed

	// if sound, ok := content["sound"]; ok {
	// 	p.SetSound(sound.(string))
	// }

	if newSpeed, ok := content["speed"]; ok {
		speed = newSpeed.(float64)
	}

	p.SetSpeed(speed)

	p.done = false
	p.phase = 0.0
}

func (p *MMPlayer) IsActive() bool {
	return !p.oneShot || (p.oneShot && !p.done)
}

func (p *MMPlayer) Synthesize() bool {
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

		x := p.phase * float64(p.mmBuf.NumFrames)
		xi1 := int64(x)
		xi2 := xi1 + 1
		xf := x - float64(xi1)

		if xi2 >= p.mmBuf.NumFrames {
			if p.oneShot {
				xi2 = p.mmBuf.NumFrames - 1
			} else {
				xi2 = 0
			}
		}

		for outIndex, out := range p.Outputs {
			xi1v := p.mmBuf.Channels[outIndex].Buffer(p.depth)[xi1]
			out.Buffer[i] = p.amp * (xi1v + (p.mmBuf.Channels[outIndex].Buffer(p.depth)[xi2]-xi1v)*xf)
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
