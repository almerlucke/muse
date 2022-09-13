package lfo

import (
	"github.com/almerlucke/muse"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/values/prototype"
)

type Target struct {
	Address     string
	Shaper      shaping.Shaper
	Placeholder string
	Proto       prototype.Prototype
}

func NewTarget(address string, shaper shaping.Shaper, placeholder string, proto prototype.Prototype) *Target {
	return &Target{Address: address, Shaper: shaper, Placeholder: placeholder, Proto: proto}
}

func (t *Target) replacements(value float64) []*prototype.Replacement {
	if t.Shaper == nil {
		return []*prototype.Replacement{prototype.NewReplacement(t.Placeholder, value)}
	}

	return []*prototype.Replacement{prototype.NewReplacement(t.Placeholder, t.Shaper.Shape(value))}
}

func (t *Target) Messages(value float64) []*muse.Message {
	raw := t.Proto.Map(t.replacements(value))
	msgs := make([]*muse.Message, len(raw))

	for i, msg := range raw {
		msgs[i] = muse.NewMessage(t.Address, msg)
	}

	return msgs
}

type LFO struct {
	*muse.BaseMessenger
	phase   float64
	delta   float64
	targets []*Target
}

func NewLFO(frequency float64, targets []*Target, config *muse.Configuration, identifier string) *LFO {
	sampleRate := config.SampleRate / float64(config.BufferSize)

	return &LFO{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		delta:         frequency / sampleRate,
		targets:       targets,
	}
}

func (lfo *LFO) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	out := lfo.phase

	lfo.phase += lfo.delta

	for lfo.phase >= 1.0 {
		lfo.phase -= 1.0
	}
	for lfo.phase < 0.0 {
		lfo.phase += 1.0
	}

	msgs := []*muse.Message{}

	for _, target := range lfo.targets {
		targetMsgs := target.Messages(out)
		msgs = append(msgs, targetMsgs...)
	}

	return msgs
}
