package timer

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
)

type Timer struct {
	*muse.BaseMessenger
	addresses     []string
	interval      float64
	intervalMilli float64
	lastMultiple  int64
	sampleRate    float64
}

func NewTimer(intervalMilli float64, addresses []string, config *muse.Configuration, identifier string) *Timer {
	return &Timer{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		addresses:     addresses,
		interval:      intervalMilli * 0.001 * config.SampleRate,
		intervalMilli: intervalMilli,
		sampleRate:    config.SampleRate,
	}
}

func NewControlTimer(intervalMilli float64, config *muse.Configuration, identifier string) *Timer {
	return &Timer{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		interval:      intervalMilli * 0.001 * config.SampleRate,
		intervalMilli: intervalMilli,
		sampleRate:    config.SampleRate,
	}
}

func (t *Timer) ReceiveControlValue(value any, index int) {
	if intervalMilli, ok := value.(float64); ok {
		if intervalMilli > 0 {
			t.intervalMilli = intervalMilli
			t.interval = intervalMilli * 0.001 * t.sampleRate
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

func (t *Timer) tick(timestamp int64, config *muse.Configuration) (bool, float64) {
	bang := false

	multiple := int64(float64(timestamp) / t.interval)

	if timestamp == 0 || multiple != t.lastMultiple {
		bang = true
	}

	t.lastMultiple = multiple

	return bang, t.intervalMilli
}

func (t *Timer) Tick(timestamp int64, config *muse.Configuration) {
	bang, duration := t.tick(timestamp, config)

	if bang {
		t.SendControlValue(duration, 1)
		t.SendControlValue(banger.Bang, 0)
	}
}

func (t *Timer) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	messages := []*muse.Message{}
	bang, duration := t.tick(timestamp, config)

	if bang {
		t.SendControlValue(duration, 1)
		t.SendControlValue(banger.Bang, 0)

		for _, address := range t.addresses {
			messages = append(messages, &muse.Message{
				Address: address,
				Content: banger.Bang,
			})
		}
	}

	return messages
}
