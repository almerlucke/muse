package xfade

import "github.com/almerlucke/muse"

type XFade struct {
	*muse.BaseModule
	fade float64
}

func New(fade float64) *XFade {
	xf := &XFade{
		BaseModule: muse.NewBaseModule(3, 1),
		fade:       fade,
	}

	xf.SetSelf(xf)

	return xf
}

func (x *XFade) ReceiveControlValue(value any, index int) {
	if index == 0 {
		if fade, ok := value.(float64); ok {
			x.fade = fade
		}
	}
}

func (x *XFade) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if fade, ok := content["fade"]; ok {
		x.fade = fade.(float64)
	}

	return nil
}

func (x *XFade) Synthesize() bool {
	if !x.BaseModule.Synthesize() {
		return false
	}

	in1 := x.Inputs[0].Buffer
	in2 := x.Inputs[1].Buffer
	out := x.Outputs[0].Buffer

	fade := x.fade

	for i := 0; i < x.Config.BufferSize; i++ {
		if x.Inputs[2].IsConnected() {
			fade = x.Inputs[2].Buffer[i]
		}

		out[i] = in1[i] + (in2[i]-in1[i])*fade
	}

	return true
}
