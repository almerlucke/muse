package main

import (
	"fmt"
	"github.com/almerlucke/genny/and"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/midi/notegen"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/utils/notes"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 96000.0,
		BufferSize: 512,
	})

	root := muse.New(0)

	fmt.Println(midi.GetInPorts())

	var out, _ = midi.OutPort(1)
	// takes the first out port, for real, consider
	// var out = OutByName("my synth")

	send, _ := midi.SendTo(out)

	root.AddMidiClock(120.0, send)

	durationGen := function.NewRandom(250.0, 3000.0)
	velocityGen := function.NewRandom(0.7, 1.0)
	noteGen := and.NewLoop[notes.Note](
		bucket.NewLoop(bucket.Indexed, notes.HungarianMinor.Root(notes.A3)...),
		bucket.NewLoop(bucket.Indexed, notes.HungarianMinor.Root(notes.A4)...),
		bucket.NewLoop(bucket.Indexed, notes.HungarianMinor.Root(notes.A5)...),
		bucket.NewLoop(bucket.Indexed, notes.HungarianMinor.Root(notes.A4)...),
	)

	ng := notegen.New(0, noteGen, velocityGen, durationGen, send).CtrlAddTo(root).CtrlIn(timer.NewControl(500, nil).CtrlAddTo(root)).(*notegen.NoteGen)

	defer func() {
		ng.NotesOff()
		root.MidiStop()
	}()

	root.MidiStart()

	_ = root.RenderAudio()
}
