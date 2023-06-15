package lfo

import (
	"github.com/almerlucke/muse"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/value/template"
)

var lfoSineShaper = shaping.NewNormalizedSineTable(512.0)

type Target struct {
	Address   string
	Shaper    shaping.Shaper
	Parameter string
	Template  template.Template
}

func NewTarget(address string, shaper shaping.Shaper, parameter string, template template.Template) *Target {
	return &Target{Address: address, Shaper: shaper, Parameter: parameter, Template: template}
}

func (t *Target) Messages(value float64) []*muse.Message {
	if t.Shaper == nil {
		t.Template.SetParameter(t.Parameter, value)
	} else {
		t.Template.SetParameter(t.Parameter, t.Shaper.Shape(value))
	}

	raw := t.Template.Value()
	msgs := make([]*muse.Message, len(raw))

	for i, msg := range raw {
		msgs[i] = muse.NewMessage(t.Address, msg)
	}

	return msgs
}

type LFO struct {
	*muse.BaseMessenger
	config     *muse.Configuration
	phase      float64
	delta      float64
	speed      float64
	min        float64
	max        float64
	shapeIndex int
	shapes     []shaping.Shaper
	targets    []*Target
}

func NewControlLFO(speed float64, min float64, max float64, shapeIndex int, shapes []shaping.Shaper, config *muse.Configuration, identifier string) *LFO {
	sampleRate := config.SampleRate / float64(config.BufferSize)

	if min > max {
		tmp := min
		min = max
		max = tmp
	}

	lfo := &LFO{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		delta:         speed / sampleRate,
		speed:         speed,
		min:           min,
		max:           max,
		shapes:        shapes,
		shapeIndex:    shapeIndex,
		config:        config,
	}

	lfo.SetSelf(lfo)

	return lfo
}

func NewBasicControlLFO(speed float64, min float64, max float64, config *muse.Configuration, identifier string) *LFO {
	return NewControlLFO(speed, min, max, 0, []shaping.Shaper{lfoSineShaper}, config, identifier)
}

func NewLFO(speed float64, targets []*Target, config *muse.Configuration, identifier string) *LFO {
	sampleRate := config.SampleRate / float64(config.BufferSize)

	lfo := &LFO{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		delta:         speed / sampleRate,
		speed:         speed,
		targets:       targets,
		min:           0.0,
		max:           1.0,
		shapes:        []shaping.Shaper{lfoSineShaper},
		config:        config,
	}

	lfo.SetSelf(lfo)

	return lfo
}

func NewBasicLFO(speed float64, scale float64, offset float64, addresses []string, config *muse.Configuration, param string, templ template.Template) *LFO {
	ts := make([]*Target, len(addresses))
	for i, address := range addresses {
		ts[i] = NewTarget(address, shaping.NewSerial(lfoSineShaper, shaping.NewLinear(scale, offset)), param, templ)
	}

	return NewLFO(speed, ts, config, "")
}

func (lfo *LFO) ReceiveControlValue(value any, index int) {
	// Index == 0 -> speed (float)
	// Index == 1 -> min (float)
	// Index == 2 -> max (float)
	// Index == 3 -> shape index (int or float)

	switch index {
	case 0: // speed
		if speed, ok := value.(float64); ok {
			lfo.SetSpeed(speed)
		}
	case 1: // min
		if min, ok := value.(float64); ok {
			lfo.SetMin(min)
		}
	case 2: // max
		if max, ok := value.(float64); ok {
			lfo.SetMax(max)
		}
	case 3: // shape index
		lfo.SetShapeIndex(value)
	}
}

func (lfo *LFO) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if speed, ok := content["speed"]; ok {
		lfo.SetSpeed(speed.(float64))
	}

	if min, ok := content["min"]; ok {
		lfo.SetMin(min.(float64))
	}

	if max, ok := content["max"]; ok {
		lfo.SetMax(max.(float64))
	}

	if shapeIndex, ok := content["shapeIndex"]; ok {
		lfo.SetShapeIndex(shapeIndex)
	}

	return nil
}

func (lfo *LFO) Speed() float64 {
	return lfo.speed
}

func (lfo *LFO) SetSpeed(speed float64) {
	sampleRate := lfo.config.SampleRate / float64(lfo.config.BufferSize)
	lfo.delta = speed / sampleRate
	lfo.speed = speed
}

func (lfo *LFO) Min() float64 {
	return lfo.min
}

func (lfo *LFO) SetMin(min float64) {
	if min > lfo.max {
		lfo.min = lfo.max
		lfo.max = min
	} else {
		lfo.min = min
	}
}

func (lfo *LFO) Max() float64 {
	return lfo.max
}

func (lfo *LFO) SetMax(max float64) {
	if max < lfo.min {
		lfo.max = lfo.min
		lfo.min = max
	} else {
		lfo.max = max
	}
}

func (lfo *LFO) ShapeIndex() int {
	return lfo.shapeIndex
}

func (lfo *LFO) SetShapeIndex(anyIndex any) {
	index := lfo.shapeIndex

	if newIndex, ok := anyIndex.(float64); ok {
		index = int(newIndex)
	} else if newIndex, ok := anyIndex.(int); ok {
		index = newIndex
	}

	if index < len(lfo.shapes) {
		lfo.shapeIndex = index
	}
}

func (lfo *LFO) Tick(timestamp int64, config *muse.Configuration) {
	out := (lfo.max-lfo.min)*lfo.shapes[lfo.shapeIndex].Shape(lfo.phase) + lfo.min

	lfo.phase += lfo.delta

	for lfo.phase >= 1.0 {
		lfo.phase -= 1.0
	}

	for lfo.phase < 0.0 {
		lfo.phase += 1.0
	}

	lfo.SendControlValue(out, 0)
}

func (lfo *LFO) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	out := lfo.phase
	controlOut := (lfo.max-lfo.min)*lfo.shapes[lfo.shapeIndex].Shape(lfo.phase) + lfo.min

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

	lfo.SendControlValue(controlOut, 0)

	return msgs
}
