package muse

import "github.com/almerlucke/muse/utils/pool"

type Voice interface {
	Module
	Activate(duration float64, amplitude float64, message any, config *Configuration)
	IsActive() bool
}

type VoicePlayer struct {
	*BaseModule
	freePool   pool.Pool[Voice]
	activePool pool.Pool[Voice]
}

func NewVoicePlayer(numChannels int, voices []Voice, config *Configuration, identifier string) *VoicePlayer {
	player := &VoicePlayer{
		BaseModule: NewBaseModule(1, numChannels, config, identifier),
	}

	player.freePool.Initialize()
	player.activePool.Initialize()

	for _, voice := range voices {
		player.freePool.Push(&pool.Element[Voice]{Value: voice})
	}

	return player
}

// ReceiveMessage is used to activate voices
func (vp *VoicePlayer) ReceiveMessage(msg any) []*Message {
	content := msg.(map[string]any)

	duration := content["duration"].(float64)
	amplitude := content["amplitude"].(float64)
	voiceMsg := content["message"]

	elem := vp.freePool.Pop()
	if elem != nil {
		elem.Value.Activate(duration, amplitude, voiceMsg, vp.Config)
		vp.activePool.Push(elem)
	}

	return nil
}

func (vp *VoicePlayer) Synthesize() bool {
	if !vp.BaseModule.Synthesize() {
		return false
	}

	// Clear output buffers
	for _, output := range vp.Outputs {
		output.Buffer.Clear()
	}

	// First prepare all voices for synthesis
	elem := vp.activePool.First()
	end := vp.activePool.End()
	for elem != end {
		elem.Value.PrepareSynthesis()
		elem = elem.Next
	}

	// Run active voices
	elem = vp.activePool.First()
	cnt := 0
	for elem != end {
		cnt++
		prev := elem
		elem = elem.Next

		if prev.Value.IsActive() {
			// Add voice output to buffer
			prev.Value.Synthesize()

			for outputIndex := 0; outputIndex < len(vp.Outputs); outputIndex++ {
				socket := prev.Value.OutputAtIndex(outputIndex)
				for sampIndex := 0; sampIndex < vp.Config.BufferSize; sampIndex++ {
					vp.Outputs[outputIndex].Buffer[sampIndex] += socket.Buffer[sampIndex]
				}
			}
		} else {
			prev.Unlink()
			vp.freePool.Push(prev)
		}
	}

	return true
}
