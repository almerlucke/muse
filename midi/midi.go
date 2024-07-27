package midi

import (
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

func OpenOutPort(port int) (drivers.Out, func(msg midi.Message) error, error) {
	out, err := midi.OutPort(port)
	if err != nil {
		return nil, nil, err
	}

	err = out.Open()
	if err != nil {
		return nil, nil, err
	}

	send, err := midi.SendTo(out)
	if err != nil {
		_ = out.Close()
		return nil, nil, err
	}

	return out, send, nil
}
