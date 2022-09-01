package lfo

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/shaping"
	"github.com/almerlucke/muse/values"
)

type LFOTarget struct {
	Address         string
	PlaceholderName string
	ProtoMessage    values.MapPrototype
}

func NewLFOTarget(address string, placeholderName string, protoMessage values.MapPrototype) *LFOTarget {
	return &LFOTarget{PlaceholderName: placeholderName, ProtoMessage: protoMessage, Address: address}
}

type LFO struct {
	*muse.BaseMessenger
	phase     float64
	delta     float64
	min       float64
	max       float64
	shaper    shaping.Shaper
	paramName string
	targets   []*LFOTarget
}

func NewLFO(phase float64, frequency float64, min float64, max float64, shaper shaping.Shaper, targets []*LFOTarget, config *muse.Configuration, identifier string) *LFO {
	sampleRate := config.SampleRate / float64(config.BufferSize)

	return &LFO{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		delta:         frequency / sampleRate,
		phase:         phase,
		min:           min,
		max:           max,
		targets:       targets,
		shaper:        shaper,
	}
}

func (lfo *LFO) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	var out float64

	if lfo.shaper != nil {
		out = lfo.shaper.Shape(lfo.phase)
	} else {
		out = lfo.phase
	}

	out = lfo.min + (lfo.max-lfo.min)*out

	lfo.phase += lfo.delta
	for lfo.phase >= 1.0 {
		lfo.phase -= 1.0
	}
	for lfo.phase < 0.0 {
		lfo.phase += 1.0
	}

	msgs := make([]*muse.Message, len(lfo.targets))

	for index, target := range lfo.targets {
		msg := target.ProtoMessage.Map([]string{target.PlaceholderName}, []any{out})
		msgs[index] = muse.NewMessage(target.Address, msg)
	}

	return msgs
}
