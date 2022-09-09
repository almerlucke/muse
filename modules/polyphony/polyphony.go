package polyphony

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/pool"
)

type Voice interface {
	muse.Module
	NoteOn(amplitude float64, message any, config *muse.Configuration)
	NoteOff()
	Activate(duration float64, amplitude float64, message any, config *muse.Configuration)
	IsActive() bool
}

type Polyphony struct {
	*muse.BaseModule
	freePool   *pool.Pool[Voice]
	activePool *pool.Pool[Voice]
}

func NewPolyphony(numChannels int, voices []Voice, config *muse.Configuration, identifier string) *Polyphony {
	player := &Polyphony{
		BaseModule: muse.NewBaseModule(1, numChannels, config, identifier),
	}

	player.freePool = pool.NewPool[Voice]()
	player.activePool = pool.NewPool[Voice]()

	for _, voice := range voices {
		player.freePool.Push(&pool.Element[Voice]{Value: voice})
	}

	return player
}

func (p *Polyphony) noteOff(identifier string) {
	elem := p.activePool.First()
	end := p.activePool.End()
	for elem != end {
		if elem.Value.Identifier() == identifier {
			elem.Value.NoteOff()
			break
		}
		elem = elem.Next
	}
}

func (p *Polyphony) AllNotesOff() {
	elem := p.activePool.First()
	end := p.activePool.End()
	for elem != end {
		elem.Value.NoteOff()
		elem = elem.Next
	}
}

// ReceiveMessage is used to activate voices
func (p *Polyphony) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	command := content["command"].(string)
	if command == "trigger" {
		// Trigger a voice
		elem := p.freePool.Pop()
		if elem != nil {
			if noteOffIdentifier, ok := content["noteOff"]; ok {
				p.noteOff(noteOffIdentifier.(string))
			} else if noteOnIdentifier, ok := content["noteOn"]; ok {
				amplitude := content["amplitude"].(float64)
				voiceMsg := content["message"]
				elem.Value.SetIdentifier(noteOnIdentifier.(string))
				elem.Value.NoteOn(amplitude, voiceMsg, p.Config)
			} else if duration, ok := content["duration"]; ok {
				amplitude := content["amplitude"].(float64)
				voiceMsg := content["message"]
				elem.Value.Activate(duration.(float64), amplitude, voiceMsg, p.Config)
			}

			p.activePool.Push(elem)
		}
	} else if command == "voice" {
		// Pass message to active voices
		elem := p.activePool.First()
		end := p.activePool.End()
		for elem != end {
			if elem.Value.IsActive() {
				elem.Value.ReceiveMessage(msg)
			}
			elem = elem.Next
		}
	}

	return nil
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
	elem := p.activePool.First()
	end := p.activePool.End()
	for elem != end {
		elem.Value.PrepareSynthesis()
		elem = elem.Next
	}

	// Run active voices
	elem = p.activePool.First()
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