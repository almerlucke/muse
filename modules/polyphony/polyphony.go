package polyphony

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/pool"
)

/*
TODO: add age and steal voices if needed
*/

type Voice interface {
	muse.Module
	NoteOn(amplitude float64, message any, config *muse.Configuration)
	NoteOff()
	Note(duration float64, amplitude float64, message any, config *muse.Configuration)
	Clear()
	IsActive() bool
}

type voiceInfo struct {
	age            int64
	isStolen       bool
	nextMsg        map[string]any
	nextIdentifier string
	voice          Voice
}

type Polyphony struct {
	*muse.BaseModule
	freePool   *pool.Pool[*voiceInfo]
	activePool *pool.Pool[*voiceInfo]
}

func New(numChannels int, voices []Voice) *Polyphony {
	poly := &Polyphony{
		BaseModule: muse.NewBaseModule(1, numChannels),
	}

	poly.freePool = pool.New[*voiceInfo]()
	poly.activePool = pool.New[*voiceInfo]()

	for _, voice := range voices {
		poly.freePool.Push(&pool.Element[*voiceInfo]{Value: &voiceInfo{
			voice: voice,
		}})
	}

	poly.SetSelf(poly)

	return poly
}

func (p *Polyphony) noteOff(identifier string) {
	p.CallActiveVoiceInfo(func(info *voiceInfo) bool {
		if info.isStolen && info.nextIdentifier == identifier {
			info.isStolen = false
			info.nextMsg = nil
			info.nextIdentifier = ""
		} else if info.voice.Identifier() == identifier {
			info.voice.NoteOff()
			info.voice.SetIdentifier("")
			return false
		}

		return true
	})
}

func (p *Polyphony) AllNotesOff() {
	p.CallActiveVoices(func(v Voice) bool {
		v.NoteOff()
		v.SetIdentifier("")
		return true
	})
}

func (p *Polyphony) ReceiveControlValue(value any, index int) {
	if index == 0 {
		p.ReceiveMessage(value)
	}
}

func (p *Polyphony) handleTriggerMessage(msg map[string]any, identifier string, duration float64, isNoteOn bool) {
	v := p.getFreeVoice()
	if v != nil {
		if isNoteOn {
			v.SetIdentifier(identifier)
			v.NoteOn(msg["amplitude"].(float64), msg["message"], p.Config)
		} else {
			v.Note(duration, msg["amplitude"].(float64), msg["message"], p.Config)
		}
	} else {
		// Steal oldest active voice
		info := p.getOldestActiveVoiceInfo()
		if info != nil {
			info.isStolen = true
			info.nextMsg = msg
			info.nextIdentifier = identifier
		}
	}
}

func (p *Polyphony) activateStolenVoiceInfo(info *voiceInfo) {
	if noteOnIdentifier, ok := info.nextMsg["noteOn"]; ok {
		info.voice.Clear()
		info.voice.SetIdentifier(noteOnIdentifier.(string))
		info.voice.NoteOn(info.nextMsg["amplitude"].(float64), info.nextMsg["message"], p.Config)
		info.isStolen = false
		info.nextMsg = nil
		info.age = 0
		info.nextIdentifier = ""
	} else if duration, ok := info.nextMsg["duration"]; ok {
		info.voice.Clear()
		info.voice.Note(duration.(float64), info.nextMsg["amplitude"].(float64), info.nextMsg["message"], p.Config)
		info.isStolen = false
		info.nextMsg = nil
		info.age = 0
		info.nextIdentifier = ""
	}
}

// ReceiveMessage is used to activate voices
func (p *Polyphony) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	command := content["command"].(string)
	if command == "trigger" {
		// Trigger a voice
		if noteOffIdentifier, ok := content["noteOff"]; ok {
			p.noteOff(noteOffIdentifier.(string))
		} else if noteOnIdentifier, ok := content["noteOn"]; ok {
			p.handleTriggerMessage(content, noteOnIdentifier.(string), 0, true)
		} else if duration, ok := content["duration"]; ok {
			p.handleTriggerMessage(content, "", duration.(float64), false)
		}
	} else if command == "voice" {
		// Pass message to all voices
		p.CallVoices(func(v Voice) {
			v.ReceiveMessage(msg)
		})
	}

	return nil
}

func (p *Polyphony) getOldestActiveVoiceInfo() *voiceInfo {
	elem := p.activePool.First()
	end := p.activePool.End()

	var oldest *voiceInfo

	for elem != end {
		if !elem.Value.isStolen && (oldest == nil || elem.Value.age > oldest.age) {
			oldest = elem.Value
		}
		elem = elem.Next
	}

	return oldest
}

func (p *Polyphony) getFreeVoice() Voice {
	elem := p.freePool.Pop()
	if elem != nil {
		p.activePool.Push(elem)
		elem.Value.age = 0
		return elem.Value.voice
	}

	return nil
}

func (p *Polyphony) CallVoices(f func(Voice)) {
	p.CallActiveVoices(func(v Voice) bool {
		f(v)
		return true
	})
	p.CallInactiveVoices(func(v Voice) bool {
		f(v)
		return true
	})
}

func (p *Polyphony) CallActiveVoiceInfo(f func(*voiceInfo) bool) {
	elem := p.activePool.First()
	end := p.activePool.End()
	for elem != end {
		ok := f(elem.Value)
		if !ok {
			break
		}
		elem = elem.Next
	}
}

func (p *Polyphony) CallActiveVoices(f func(Voice) bool) {
	p.CallActiveVoiceInfo(func(info *voiceInfo) bool {
		return f(info.voice)
	})
}

func (p *Polyphony) CallInactiveVoices(f func(Voice) bool) {
	elem := p.freePool.First()
	end := p.freePool.End()
	for elem != end {
		ok := f(elem.Value.voice)
		if !ok {
			break
		}
		elem = elem.Next
	}
}

func (p *Polyphony) Synthesize() bool {
	if !p.BaseModule.Synthesize() {
		return false
	}

	// Clear output buffers
	for _, output := range p.Outputs {
		output.Buffer.Clear()
	}

	// First prepare all voices for synthesis
	p.CallActiveVoices(func(v Voice) bool {
		v.PrepareSynthesis()
		return true
	})

	// Run active voices
	end := p.activePool.End()
	elem := p.activePool.First()
	for elem != end {
		prev := elem
		elem = elem.Next
		info := prev.Value
		voice := prev.Value.voice

		if voice.IsActive() {
			// Add voice output to buffer
			voice.Synthesize()

			if info.isStolen {
				// Fade out voice over 1 buffer cycle
				x := 1.0 / float64(p.Config.BufferSize)
				for outputIndex := 0; outputIndex < len(p.Outputs); outputIndex++ {
					socket := voice.OutputAtIndex(outputIndex)
					for sampIndex := 0; sampIndex < p.Config.BufferSize; sampIndex++ {
						p.Outputs[outputIndex].Buffer[sampIndex] += socket.Buffer[sampIndex] * (1.0 - x*float64(sampIndex))
					}
				}
				p.activateStolenVoiceInfo(info)
			} else {
				for outputIndex := 0; outputIndex < len(p.Outputs); outputIndex++ {
					socket := voice.OutputAtIndex(outputIndex)
					for sampIndex := 0; sampIndex < p.Config.BufferSize; sampIndex++ {
						p.Outputs[outputIndex].Buffer[sampIndex] += socket.Buffer[sampIndex]
					}
				}
			}
		} else if info.isStolen {
			p.activateStolenVoiceInfo(info)
		} else {
			prev.Unlink()
			p.freePool.Push(prev)
		}
	}

	return true
}
