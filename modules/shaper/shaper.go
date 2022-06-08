package shaper

import (
	"github.com/almerlucke/muse"
	shapingc "github.com/almerlucke/muse/components/shaping"
)

type ParamMapFunction func(int, float64, shapingc.Shaper)

type MessageMapFunction func(any, shapingc.Shaper)

type Shaper struct {
	*muse.BaseModule
	shaper      shapingc.Shaper
	paramMapper ParamMapFunction
	msgMapper   MessageMapFunction
}

func NewShaper(shaper shapingc.Shaper, numParams int, paramMapper ParamMapFunction, msgMapper MessageMapFunction, config *muse.Configuration, identifier string) *Shaper {
	return &Shaper{
		BaseModule:  muse.NewBaseModule(numParams+1, 1, config, identifier),
		shaper:      shaper,
		paramMapper: paramMapper,
		msgMapper:   msgMapper,
	}
}

func (s *Shaper) ReceiveMessage(msg any) []*muse.Message {
	if s.msgMapper != nil {
		s.msgMapper(msg, s.shaper)
	}

	return nil
}

func (s *Shaper) Synthesize() bool {
	if !s.BaseModule.Synthesize() {
		return false
	}

	output := s.Outputs[0].Buffer
	input := s.Inputs[0].Buffer

	for i := 0; i < s.Config.BufferSize; i++ {
		// Map param inputs to shaper object
		if s.paramMapper != nil {
			for j := 1; j < len(s.Inputs); j++ {
				if s.Inputs[j].IsConnected() {
					s.paramMapper(j-1, s.Inputs[j].Buffer[i], s.shaper)
				}
			}
		}

		// Shape input phase
		output[i] = s.shaper.Shape(input[i])
	}

	return true
}
