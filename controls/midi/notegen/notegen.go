package notegen

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/utils/pool"
	"github.com/almerlucke/muse/utils/timing"
	"gitlab.com/gomidi/midi/v2"
)

type noteInfo struct {
	channel      uint8
	key          uint8
	offTimestamp int64
}

type NoteGen struct {
	*muse.BaseControl
	activeNotes   *pool.Pool[*noteInfo]
	send          func(msg midi.Message) error
	lastTimestamp int64
	sampleRate    float64
	noteGen       genny.Generator[notes.Note]
	velocityGen   genny.Generator[float64]
	durationGen   genny.Generator[float64]
}

func New(noteGen genny.Generator[notes.Note], velocityGen genny.Generator[float64], durationGen genny.Generator[float64], send func(msg midi.Message) error) *NoteGen {
	ng := &NoteGen{
		BaseControl: muse.NewBaseControl(),
		activeNotes: pool.New[*noteInfo](),
		send:        send,
		sampleRate:  muse.CurrentConfiguration().SampleRate,
		noteGen:     noteGen,
		velocityGen: velocityGen,
		durationGen: durationGen,
	}

	ng.SetSelf(ng)

	return ng
}

func (ng *NoteGen) hasActiveNote(key uint8) bool {
	it := ng.activeNotes.Iterator()
	for v, ok := it.Next(); ok; v, ok = it.Next() {
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
			ng.activeNotes.PushValue(&noteInfo{
				channel:      0,
				key:          key,
				offTimestamp: ng.lastTimestamp + timing.MilliToSamps(durationMs, ng.sampleRate),
			})
			_ = ng.send(midi.NoteOn(0, key, velocity))
		}
	}
}

func (ng *NoteGen) Tick(timestamp int64, _ *muse.Configuration) {
	ng.lastTimestamp = timestamp

	it := ng.activeNotes.Iterator()
	for {
		if v, ok := it.Value(); ok {
			if v.offTimestamp < timestamp {
				_ = ng.send(midi.NoteOff(v.channel, v.key))
				it.Remove()
			}
			it.Next()
		} else {
			break
		}
	}
}
