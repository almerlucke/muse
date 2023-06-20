package player

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
)

type Player struct {
	*muse.BaseModule
	sf          io.SoundFiler
	startOffset float64
	endOffset   float64
	phase       float64
	inc         float64
	speed       float64
	amp         float64
	depth       int
	oneShot     bool
	done        bool
	soundBank   io.SoundBank
}

func NewPlayer(sf io.SoundFiler, speed float64, amp float64, oneShot bool, config *muse.Configuration, identifier string) *Player {
	return NewPlayerX(sf, speed, amp, 0.0, sf.Duration(), oneShot, config, identifier)
}

func NewPlayerX(sf io.SoundFiler, speed float64, amp float64, startOffset float64, endOffset float64, oneShot bool, config *muse.Configuration, identifier string) *Player {
	inc := (speed * sf.SampleRate() / config.SampleRate) / float64(sf.NumFrames())

	depth := io.SpeedToMipMapDepth(speed)
	if depth >= sf.Depth() {
		depth = sf.Depth() - 1
	}

	p := &Player{
		BaseModule: muse.NewBaseModule(0, sf.NumChannels(), config, identifier),
		inc:        inc,
		speed:      speed,
		oneShot:    oneShot,
		done:       oneShot,
		sf:         sf,
		amp:        amp,
	}

	p.startOffset = p.normalizeDurationOffset(startOffset)
	p.endOffset = p.normalizeDurationOffset(endOffset)

	if p.inc < 0.0 {
		p.phase = p.endOffset
	} else {
		p.phase = p.startOffset
	}

	p.SetSelf(p)

	return p
}

// offset in seconds
func (p *Player) normalizeDurationOffset(offset float64) float64 {
	numFrames := float64(p.sf.NumFrames())
	frameOffset := offset * p.sf.SampleRate()
	if frameOffset >= numFrames {
		frameOffset = numFrames - 1.0
	}
	return frameOffset / numFrames
}

func (p *Player) SetSoundBank(soundBank io.SoundBank) {
	p.soundBank = soundBank
}

func (p *Player) SetSound(sound string) {
	if p.soundBank != nil {
		if sf, ok := p.soundBank[sound]; ok {
			p.SetSoundFile(sf)
		}
	}
}

func (p *Player) SetSoundFile(sf io.SoundFiler) {
	p.sf = sf
	p.inc = (p.speed * p.sf.SampleRate() / p.Config.SampleRate) / float64(p.sf.NumFrames())
	depth := io.SpeedToMipMapDepth(p.speed)
	if depth >= sf.Depth() {
		depth = sf.Depth() - 1
	}
}

func (p *Player) Speed() float64 {
	return p.speed
}

func (p *Player) SetSpeed(speed float64) {
	p.inc = (speed * p.sf.SampleRate() / p.Config.SampleRate) / float64(p.sf.NumFrames())
	p.speed = speed
	depth := io.SpeedToMipMapDepth(speed)
	if depth >= p.sf.Depth() {
		depth = p.sf.Depth() - 1
	}
}

func (p *Player) SetStartOffset(offset float64) {
	p.startOffset = p.normalizeDurationOffset(offset)
}

func (p *Player) SetEndOffset(offset float64) {
	p.endOffset = p.normalizeDurationOffset(offset)
}

func (p *Player) Bang() {
	if p.oneShot {
		p.done = false
		if p.inc < 0.0 {
			p.phase = p.endOffset
		} else {
			p.phase = p.startOffset
		}
	}
}

func (p *Player) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Bang
		if value == muse.Bang {
			p.Bang()
		}
	case 1: // Speed
		p.SetSpeed(value.(float64))
	case 2: // Start offset
		p.SetStartOffset(value.(float64))
	case 3: // End offset
		p.SetEndOffset(value.(float64))
	}
}

func (p *Player) ReceiveMessage(msg any) []*muse.Message {
	if content, ok := msg.(map[string]any); ok {
		if speed, ok := content["speed"]; ok {
			p.SetSpeed(speed.(float64))
		}

		if startOffset, ok := content["startOffset"]; ok {
			p.SetStartOffset(startOffset.(float64))
		}

		if endOffset, ok := content["endOffset"]; ok {
			p.SetEndOffset(endOffset.(float64))
		}
	}

	if muse.IsBang(msg) {
		p.Bang()
	}

	return nil
}

func (p *Player) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	p.activate(amplitude, message, config)
}

func (p *Player) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
	p.activate(amplitude, message, config)
}

func (p *Player) NoteOff() {}

func (p *Player) activate(amplitude float64, message any, config *muse.Configuration) {
	content := message.(map[string]any)

	p.amp = amplitude

	speed := p.speed

	if sound, ok := content["sound"]; ok {
		p.SetSound(sound.(string))
	}

	if newSpeed, ok := content["speed"]; ok {
		speed = newSpeed.(float64)
	}

	p.SetSpeed(speed)

	if startOffset, ok := content["startOffset"]; ok {
		p.SetStartOffset(startOffset.(float64))
	}

	if endOffset, ok := content["endOffset"]; ok {
		p.SetEndOffset(endOffset.(float64))
	}

	p.done = false

	if p.inc < 0 {
		p.phase = p.endOffset
	} else {
		p.phase = p.startOffset
	}
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

		samps := p.sf.LookupAll(p.phase*float64(p.sf.NumFrames()), 0, !p.oneShot)
		for outIndex, out := range p.Outputs {
			out.Buffer[i] = samps[outIndex]
		}

		p.phase += p.inc

		if p.oneShot {
			if p.inc < 0.0 {
				if p.phase <= p.startOffset {
					p.done = true
				}
			} else if p.phase >= p.endOffset {
				p.done = true
			}
		} else {
			r := p.endOffset - p.startOffset
			np := p.phase - p.startOffset

			for np >= r {
				np -= r
			}
			for np < 0.0 {
				np += r
			}

			p.phase = p.startOffset + np
		}
	}

	return true
}
