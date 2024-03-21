package midi

import (
	"log"

	"gitlab.com/gomidi/midi"
	// . "gitlab.com/gomidi/midi/midimessage/channel" // (Channel Messages)
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
)

type Listener struct {
	driver *rtmididrv.Driver
	in     midi.In
	rd     *reader.Reader
}

func NewListener(port int, callbacks ...func(*reader.Reader)) (*Listener, error) {
	drv, err := rtmididrv.New(rtmididrv.IgnoreActiveSense(), rtmididrv.IgnoreSysex(), rtmididrv.IgnoreTimeCode())
	if err != nil {
		return nil, err
	}

	ins, err := drv.Ins()
	if err != nil {
		_ = drv.Close()
		return nil, err
	}

	in := ins[port]

	log.Printf("in %v", in)

	err = in.Open()
	if err != nil {
		_ = drv.Close()
		return nil, err
	}

	rd := reader.New(
		append(callbacks, reader.NoLogger())...,
	)

	err = rd.ListenTo(in)
	if err != nil {
		_ = in.Close()
		_ = drv.Close()
		return nil, err
	}

	return &Listener{
		driver: drv,
		in:     in,
		rd:     rd,
	}, nil
}

func (ml *Listener) Close() {
	_ = ml.in.StopListening()
	_ = ml.in.Close()
	_ = ml.driver.Close()
}
