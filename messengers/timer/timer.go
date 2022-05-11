package timer

import (
	"github.com/almerlucke/muse"
)

type Timer struct {
	*muse.BaseMessenger
	addresses    []string
	interval     int64
	lastMultiple int64
}

func NewTimer(intervalSeconds float64, addresses []string, config *muse.Configuration, identifier string) *Timer {
	return &Timer{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		addresses:     addresses,
		interval:      int64(intervalSeconds * config.SampleRate),
	}
}

func (t *Timer) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	messages := []*muse.Message{}

	bang := false

	multiple := timestamp / t.interval

	if timestamp == 0 || multiple != t.lastMultiple {
		bang = true
	}

	t.lastMultiple = multiple

	if bang {
		for _, address := range t.addresses {
			messages = append(messages, &muse.Message{
				Address: address,
				Content: "bang",
			})
		}
	}

	return messages
}
