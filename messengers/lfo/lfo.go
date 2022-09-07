package lfo

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/shaping"
	"github.com/almerlucke/muse/values"
)

type Target struct {
	Address     string
	Shaper      shaping.Shaper
	Placeholder string
	Proto       values.MapPrototype
}

func NewTarget(address string, shaper shaping.Shaper, placeholder string, proto values.MapPrototype) *Target {
	return &Target{Address: address, Shaper: shaper, Placeholder: placeholder, Proto: proto}
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
		var targetOut float64

		if target.Shaper == nil {
			targetOut = out
		} else {
			targetOut = target.Shaper.Shape(out)
		}

		msg := target.Proto.Map([]string{target.Placeholder}, []any{targetOut})
		msgs[index] = muse.NewMessage(target.Address, msg)
	}

	return msgs
}
