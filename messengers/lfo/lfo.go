package lfo

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/shaping"
	val "github.com/almerlucke/muse/values"
)

type Target struct {
	Address     string
	Shaper      shaping.Shaper
	Placeholder string
	Proto       val.Prototype
}

func NewTarget(address string, shaper shaping.Shaper, placeholder string, proto val.Prototype) *Target {
	return &Target{Address: address, Shaper: shaper, Placeholder: placeholder, Proto: proto}
}

func (t *Target) replacements(value float64) []*val.Replacement {
	if t.Shaper == nil {
		return []*val.Replacement{val.NewReplacement(t.Placeholder, value)}
	}

	return []*val.Replacement{val.NewReplacement(t.Placeholder, t.Shaper.Shape(value))}
}

func (t *Target) Message(value float64) *muse.Message {
	return muse.NewMessage(t.Address, t.Proto.Map(t.replacements(value)))
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

	msgs := make([]*muse.Message, len(lfo.targets))

	for index, target := range lfo.targets {
		msgs[index] = target.Message(out)
	}

	return msgs
}
