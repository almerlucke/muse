package delay

import "github.com/almerlucke/muse"

type Delay struct {
	*muse.BaseMessenger
	messenger      muse.Messenger
	control        muse.Control
	beginTimestamp int64
	delay          int64
}

func NewControlDelay(control muse.Control, delay float64, config *muse.Configuration) *Delay {
	return NewDelay(nil, control, delay, config)
}

func NewDelay(messenger muse.Messenger, control muse.Control, delay float64, config *muse.Configuration) *Delay {
	d := &Delay{
		BaseMessenger: muse.NewBaseMessenger(""),
		control:       control,
		messenger:     messenger,
		delay:         int64(delay * config.SampleRate * 0.001),
	}
	d.SetSelf(d)
	return d
}

func (d *Delay) Tick(timestamp int64, config *muse.Configuration) {
	if d.beginTimestamp == 0 {
		d.beginTimestamp = timestamp
	}

	if (timestamp - d.beginTimestamp) >= d.delay {
		if d.control != nil {
			d.control.Tick(timestamp, config)
		}
	}
}

func (d *Delay) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	if d.beginTimestamp == 0 {
		d.beginTimestamp = timestamp
	}

	if (timestamp - d.beginTimestamp) >= d.delay {
		if d.control != nil {
			d.control.Tick(timestamp, config)
		}
		if d.messenger != nil {
			return d.messenger.Messages(timestamp, config)
		}
	}

	return nil
}
