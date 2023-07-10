package midi

import (
	"log"

	"gitlab.com/gomidi/midi"
	// . "gitlab.com/gomidi/midi/midimessage/channel" // (Channel Messages)
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
)

type MidiListener struct {
	driver *rtmididrv.Driver
	in     midi.In
	rd     *reader.Reader
}

func NewMidiListener(port int, callbacks ...func(*reader.Reader)) (*MidiListener, error) {
	drv, err := rtmididrv.New(rtmididrv.IgnoreActiveSense(), rtmididrv.IgnoreSysex(), rtmididrv.IgnoreTimeCode())
	if err != nil {
		return nil, err
	}

	ins, err := drv.Ins()
	if err != nil {
		drv.Close()
		return nil, err
	}

	in := ins[port]

	log.Printf("in %v", in)

	err = in.Open()
	if err != nil {
		drv.Close()
		return nil, err
	}

	rd := reader.New(
		append(callbacks, reader.NoLogger())...,
	)

	err = rd.ListenTo(in)
	if err != nil {
		in.Close()
		drv.Close()
		return nil, err
	}

	return &MidiListener{
		driver: drv,
		in:     in,
		rd:     rd,
	}, nil
}

func (ml *MidiListener) Close() {
	ml.in.StopListening()
	ml.in.Close()
	ml.driver.Close()
}
