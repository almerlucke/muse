package polyphony

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/pool"
	"log"
)

/*
TODO: add age and steal voices if needed
*/

type Voice interface {
	muse.Module
	NoteOn(amplitude float64, message any, config *muse.Configuration)
	NoteOff()
	Note(duration float64, amplitude float64, message any, config *muse.Configuration)
	IsActive() bool
}

type Polyphony struct {
	*muse.BaseModule
	freePool   *pool.Pool[Voice]
	activePool *pool.Pool[Voice]
}

func New(numChannels int, voices []Voice) *Polyphony {
	poly := &Polyphony{
		BaseModule: muse.NewBaseModule(1, numChannels),
	}

	poly.freePool = pool.NewPool[Voice]()
	poly.activePool = pool.NewPool[Voice]()

	for _, voice := range voices {
		poly.freePool.Push(&pool.Element[Voice]{Value: voice})
	}

	poly.SetSelf(poly)

	return poly
}

func (p *Polyphony) noteOff(identifier string) {
	p.CallActiveVoices(func(v Voice) bool {
		if v.Identifier() == identifier {
			v.NoteOff()
			v.SetIdentifier("")
			return false
		}

		return true
	})
}

func (p *Polyphony) DebugActive() {
	p.CallActiveVoices(func(v Voice) bool {
		log.Printf("active voice identifier: %s", v.Identifier())
		return true
	})
}

func (p *Polyphony) AllNotesOff() {
	p.CallActiveVoices(func(v Voice) bool {
		v.NoteOff()
		return true
	})
}

func (p *Polyphony) ReceiveControlValue(value any, index int) {
	if index == 0 {
		p.ReceiveMessage(value)
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
			v := p.GetFreeVoice()
			if v != nil {
				amplitude := content["amplitude"].(float64)
				voiceMsg := content["message"]
				v.SetIdentifier(noteOnIdentifier.(string))
				v.NoteOn(amplitude, voiceMsg, p.Config)
			}
		} else if duration, ok := content["duration"]; ok {
			v := p.GetFreeVoice()
			if v != nil {
				amplitude := content["amplitude"].(float64)
				voiceMsg := content["message"]
				v.Note(duration.(float64), amplitude, voiceMsg, p.Config)
			}
		}
	} else if command == "voice" {
		// Pass message to all voices
		p.CallVoices(func(v Voice) {
			v.ReceiveMessage(msg)
		})
	}

	return nil
}

func (p *Polyphony) GetFreeVoice() Voice {
	elem := p.freePool.Pop()
	if elem != nil {
		p.activePool.Push(elem)
		return elem.Value
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

func (p *Polyphony) CallActiveVoices(f func(Voice) bool) {
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

func (p *Polyphony) CallInactiveVoices(f func(Voice) bool) {
	elem := p.freePool.First()
	end := p.freePool.End()
	for elem != end {
		ok := f(elem.Value)
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

		if prev.Value.IsActive() {
			// Add voice output to buffer
			prev.Value.Synthesize()

			for outputIndex := 0; outputIndex < len(p.Outputs); outputIndex++ {
				socket := prev.Value.OutputAtIndex(outputIndex)
				for sampIndex := 0; sampIndex < p.Config.BufferSize; sampIndex++ {
					p.Outputs[outputIndex].Buffer[sampIndex] += socket.Buffer[sampIndex]
				}
			}
		} else {
			prev.Unlink()
			p.freePool.Push(prev)
		}
	}

	return true
}
