package timer

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/utils/timing"
)

type Timer struct {
	*muse.BaseMessenger
	addresses     []string
	interval      float64
	intervalMilli float64
	gen           genny.Generator[float64]
	accum         float64
	sampleRate    float64
}

func New(intervalMilli float64, addresses []string, gen genny.Generator[float64]) *Timer {
	if gen != nil {
		intervalMilli = gen.Generate()
	}

	t := &Timer{
		BaseMessenger: muse.NewBaseMessenger(),
		addresses:     addresses,
		interval:      timing.MilliToSampsf(intervalMilli, muse.SampleRate()),
		intervalMilli: intervalMilli,
		gen:           gen,
		sampleRate:    muse.SampleRate(),
	}

	t.accum = t.interval
	t.SetSelf(t)

	return t
}

func NewControl(intervalMilli float64, gen genny.Generator[float64]) *Timer {
	return New(intervalMilli, nil, gen)
}

func (t *Timer) ReceiveControlValue(value any, index int) {
	if index == 0 {
		if intervalMilli, ok := value.(float64); ok {
			if intervalMilli > 0 {
				t.intervalMilli = intervalMilli
				t.interval = timing.MilliToSampsf(intervalMilli, t.sampleRate)
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
			t.interval = timing.MilliToSampsf(intervalMilli, t.sampleRate)
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
			if t.gen != nil {
				if t.gen.Done() {
					t.gen.Reset()
				}
				t.intervalMilli = t.gen.Generate()
				t.interval = timing.MilliToSampsf(t.intervalMilli, t.sampleRate)
			}
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
