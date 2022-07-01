package muse

import "math"

type Voice interface {
	Module
	Activate(duration float64, message any, config *Configuration)
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

type VoiceStep struct {
	Duration   float64 // duration in milli
	InterOnset float64 // interonset in milli
	Message    any
}

type VoiceSequence interface {
	NextStep(float64, *Configuration) *VoiceStep
}

type VoicePlayer struct {
	*BaseModule
	sequence   VoiceSequence
	freePool   voicePool
	activePool voicePool
	timestamp  int64
	nextOnset  int64
	nextStep   *VoiceStep
}

func NewVoicePlayer(numChannels int, maxVoices int, sequence VoiceSequence, factory VoiceFactory, config *Configuration, identifier string) *VoicePlayer {
	player := &VoicePlayer{
		BaseModule: NewBaseModule(0, numChannels, config, identifier),
		sequence:   sequence,
	}

	player.nextStep = sequence.NextStep(0, config)
	player.nextOnset = int64(math.Ceil((player.nextStep.InterOnset * 0.001 * config.SampleRate) / float64(config.BufferSize)))

	for i := 0; i < maxVoices; i++ {
		player.freePool.Push(&voicePoolElement{voice: factory.NewVoice(config)})
	}

	return player
}

func (vp *VoicePlayer) Synthesize() bool {
	if !vp.BaseModule.Synthesize() {
		return false
	}

	timestampMilli := (float64(vp.timestamp) / vp.Config.SampleRate) * 1000.0
	done := false

	for !done {
		if vp.nextOnset == 0 {
			elem := vp.freePool.Pop()
			if elem != nil {
				vp.activePool.Push(elem)
				elem.voice.Activate(vp.nextStep.Duration, vp.nextStep.Message, vp.Config)
			}

			vp.nextStep = vp.sequence.NextStep(timestampMilli, vp.Config)
			vp.nextOnset = int64(math.Ceil((vp.nextStep.InterOnset * 0.001 * vp.Config.SampleRate) / float64(vp.Config.BufferSize)))
		} else {
			done = true
		}
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

	vp.timestamp += int64(vp.Config.BufferSize)
	vp.nextOnset--

	return true
}
