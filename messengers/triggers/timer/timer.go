package timer

import (
	"github.com/almerlucke/muse"
)

type Timer struct {
	*muse.BaseMessenger
	addresses     []string
	interval      float64
	intervalMilli float64
	accum         float64
	sampleRate    float64
}

func New(intervalMilli float64, addresses []string) *Timer {
	t := &Timer{
		BaseMessenger: muse.NewBaseMessenger(),
		addresses:     addresses,
		interval:      intervalMilli * 0.001 * muse.SampleRate(),
		intervalMilli: intervalMilli,
		sampleRate:    muse.SampleRate(),
	}

	t.SetSelf(t)

	return t
}

func NewControl(intervalMilli float64) *Timer {
	return New(intervalMilli, nil)
}

func (t *Timer) ReceiveControlValue(value any, index int) {
	if index == 0 {
		if intervalMilli, ok := value.(float64); ok {
			if intervalMilli > 0 {
				t.intervalMilli = intervalMilli
				t.interval = intervalMilli * 0.001 * t.sampleRate
			}
		}
	}
}

func (t *Timer) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]interface{})

	if interval, ok := content["interval"]; ok {
		intervalMilli := interval.(float64)
		if intervalMilli > 0 {
			t.intervalMilli = intervalMilli
			t.interval = intervalMilli * 0.001 * t.sampleRate
		}
	}

	return nil
}

func (t *Timer) Tick(timestamp int64, config *muse.Configuration) {
	_ = t.Messages(timestamp, config)
}

func (t *Timer) Messages(timestamp int64, _ *muse.Configuration) []*muse.Message {
	var (
		messages       []*muse.Message
		bang           bool
		floatTimestamp = float64(timestamp)
	)

	if floatTimestamp > t.accum {
		bang = true
		for t.accum <= floatTimestamp {
			t.accum += t.interval
		}
	}

	if bang {
		t.SendControlValue(t.intervalMilli, 1)
		t.SendControlValue(muse.Bang, 0)

		for _, address := range t.addresses {
			messages = append(messages, &muse.Message{
				Address: address,
				Content: muse.Bang,
			})
		}
	}

	return messages
}
