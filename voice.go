package muse

type Voice interface {
	Module
	Activate(duration float64, amplitude float64, message any, config *Configuration)
	IsActive() bool
}

type voicePoolElement struct {
	voice Voice
	prev  *voicePoolElement
	next  *voicePoolElement
}

func (e *voicePoolElement) Unlink() {
	e.prev.next = e.next
	e.next.prev = e.prev
}

type voicePool struct {
	sentinel *voicePoolElement
}

func (vp *voicePool) Initialize() {
	sentinel := &voicePoolElement{}
	sentinel.next = sentinel
	sentinel.prev = sentinel
	vp.sentinel = sentinel
}

func (vp *voicePool) Pop() *voicePoolElement {
	first := vp.sentinel.next

	if first == vp.sentinel {
		return nil
	}

	first.Unlink()

	return first
}

func (vp *voicePool) Push(e *voicePoolElement) {
	e.next = vp.sentinel.next
	e.prev = vp.sentinel
	vp.sentinel.next.prev = e
	vp.sentinel.next = e
}

type VoiceFactory interface {
	NewVoice(*Configuration) Voice
}

type VoicePlayer struct {
	*BaseModule
	freePool   voicePool
	activePool voicePool
}

func NewVoicePlayer(numChannels int, voices []Voice, config *Configuration, identifier string) *VoicePlayer {
	player := &VoicePlayer{
		BaseModule: NewBaseModule(1, numChannels, config, identifier),
	}

	player.freePool.Initialize()
	player.activePool.Initialize()

	for _, voice := range voices {
		player.freePool.Push(&voicePoolElement{voice: voice})
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
		vp.activePool.Push(elem)
		elem.voice.Activate(duration, amplitude, voiceMsg, vp.Config)
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
	elem := vp.activePool.sentinel.next
	for elem != vp.activePool.sentinel {
		elem.voice.PrepareSynthesis()
		elem = elem.next
	}

	// Run active voices
	elem = vp.activePool.sentinel.next
	for elem != vp.activePool.sentinel {
		prev := elem
		elem = elem.next

		if prev.voice.IsActive() {
			// Add voice output to buffer
			prev.voice.Synthesize()

			for outputIndex := 0; outputIndex < len(vp.Outputs); outputIndex++ {
				socket := prev.voice.OutputAtIndex(outputIndex)
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
