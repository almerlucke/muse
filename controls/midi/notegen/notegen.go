package notegen

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/containers/list"
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/utils/timing"
	"gitlab.com/gomidi/midi/v2"
)

type noteInfo struct {
	key          uint8
	offTimestamp int64
}

type NoteGen struct {
	*muse.BaseControl
	activeNotes   *list.List[*noteInfo]
	send          func(msg midi.Message) error
	lastTimestamp int64
	sampleRate    float64
	noteGen       genny.Generator[notes.Note]
	velocityGen   genny.Generator[float64]
	durationGen   genny.Generator[float64]
	channel       uint8
}

func New(channel uint8, noteGen genny.Generator[notes.Note], velocityGen genny.Generator[float64], durationGen genny.Generator[float64], send func(msg midi.Message) error) *NoteGen {
	ng := &NoteGen{
		BaseControl: muse.NewBaseControl(),
		activeNotes: list.New[*noteInfo](),
		send:        send,
		sampleRate:  muse.CurrentConfiguration().SampleRate,
		noteGen:     noteGen,
		velocityGen: velocityGen,
		durationGen: durationGen,
		channel:     channel,
	}

	ng.SetSelf(ng)

	return ng
}

func (ng *NoteGen) hasActiveNote(key uint8) bool {
	for it := ng.activeNotes.Iterator(true); !it.Finished(); {
		v, _ := it.Next()

		if v.key == key {
			return true
		}
	}

	return false
}

func (ng *NoteGen) ReceiveControlValue(value any, index int) {
	if muse.IsBang(value) {
		if ng.durationGen.Done() {
			ng.durationGen.Reset()
		}

		if ng.noteGen.Done() {
			ng.noteGen.Reset()
		}

		if ng.velocityGen.Done() {
			ng.velocityGen.Reset()
		}

		durationMs := ng.durationGen.Generate()
		key := uint8(ng.noteGen.Generate())
		velocity := uint8(ng.velocityGen.Generate() * 127.0)

		if !ng.hasActiveNote(key) {
			ng.activeNotes.Push(&noteInfo{
				key:          key,
				offTimestamp: ng.lastTimestamp + timing.MilliToSamps(durationMs, ng.sampleRate),
			})
			_ = ng.send(midi.NoteOn(ng.channel, key, velocity))
		}
	}
}

func (ng *NoteGen) Tick(timestamp int64, _ *muse.Configuration) {
	ng.lastTimestamp = timestamp

	ng.activeNotes.ForEachElement(func(e *list.Element[*noteInfo], index int) {
		if e.Value.offTimestamp <= timestamp {
			_ = ng.send(midi.NoteOff(ng.channel, e.Value.key))
			e.Unlink()
		}
	})
}

func (ng *NoteGen) NotesOff() {
	ng.activeNotes.ForEach(func(info *noteInfo, index int) {
		_ = ng.send(midi.NoteOff(ng.channel, info.key))
	})

	ng.activeNotes.Clear()
}
