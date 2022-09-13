package waveshaper

import (
	"github.com/almerlucke/muse"
	shaping "github.com/almerlucke/muse/components/waveshaping"
)

type ParamMapFunction func(int, float64, shaping.Shaper)

type MessageMapFunction func(any, shaping.Shaper)

type WaveShaper struct {
	*muse.BaseModule
	shaper      shaping.Shaper
	paramMapper ParamMapFunction
	msgMapper   MessageMapFunction
}

func NewWaveShaper(shaper shaping.Shaper, numParams int, paramMapper ParamMapFunction, msgMapper MessageMapFunction, config *muse.Configuration, identifier string) *WaveShaper {
	return &WaveShaper{
		BaseModule:  muse.NewBaseModule(numParams+1, 1, config, identifier),
		shaper:      shaper,
		paramMapper: paramMapper,
		msgMapper:   msgMapper,
	}
}

func (s *WaveShaper) ReceiveMessage(msg any) []*muse.Message {
	if s.msgMapper != nil {
		s.msgMapper(msg, s.shaper)
	}

	return nil
}

func (s *WaveShaper) Synthesize() bool {
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
