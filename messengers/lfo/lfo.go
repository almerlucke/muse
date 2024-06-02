package lfo

import (
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/lookup"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
)

var lfoSineShaper = lookup.NewNormalizedSineTable(512.0)

type Target struct {
	Address   string
	Shaper    shape.Shaper
	Parameter string
	Template  template.Template
}

func NewTarget(address string, shaper shape.Shaper, parameter string, template template.Template) *Target {
	return &Target{Address: address, Shaper: shaper, Parameter: parameter, Template: template}
}

func (t *Target) Messages(value float64) []*muse.Message {
	if t.Shaper == nil {
		t.Template.SetParameter(t.Parameter, value)
	} else {
		t.Template.SetParameter(t.Parameter, t.Shaper.Shape(value))
	}

	raw := t.Template.Generate()
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
	shapes     []shape.Shaper
	targets    []*Target
}

func NewControlLFO(speed float64, min float64, max float64, shapeIndex int, shapes []shape.Shaper) *LFO {
	controlRate := muse.ControlRate()

	if min > max {
		tmp := min
		min = max
		max = tmp
	}

	lfo := &LFO{
		BaseMessenger: muse.NewBaseMessenger(),
		delta:         speed / controlRate,
		speed:         speed,
		min:           min,
		max:           max,
		shapes:        shapes,
		shapeIndex:    shapeIndex,
		config:        muse.CurrentConfiguration(),
	}

	lfo.SetSelf(lfo)

	return lfo
}

func NewBasicControlLFO(speed float64, min float64, max float64) *LFO {
	return NewControlLFO(speed, min, max, 0, []shape.Shaper{lfoSineShaper})
}

func NewLFO(speed float64, targets []*Target) *LFO {
	controlRate := muse.ControlRate()

	lfo := &LFO{
		BaseMessenger: muse.NewBaseMessenger(),
		delta:         speed / controlRate,
		speed:         speed,
		targets:       targets,
		min:           0.0,
		max:           1.0,
		shapes:        []shape.Shaper{lfoSineShaper},
		config:        muse.CurrentConfiguration(),
	}

	lfo.SetSelf(lfo)

	return lfo
}

func NewBasicLFO(speed float64, scale float64, offset float64, addresses []string, param string, templ template.Template) *LFO {
	ts := make([]*Target, len(addresses))
	for i, address := range addresses {
		ts[i] = NewTarget(address, series.New(lfoSineShaper, linear.New(scale, offset)), param, templ)
	}

	return NewLFO(speed, ts)
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
		if mi, ok := value.(float64); ok {
			lfo.SetMin(mi)
		}
	case 2: // max
		if ma, ok := value.(float64); ok {
			lfo.SetMax(ma)
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

	if mi, ok := content["min"]; ok {
		lfo.SetMin(mi.(float64))
	}

	if ma, ok := content["max"]; ok {
		lfo.SetMax(ma.(float64))
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
	lfo.delta = speed / lfo.config.ControlRate()
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

	var msgs []*muse.Message

	for _, target := range lfo.targets {
		targetMsgs := target.Messages(out)
		msgs = append(msgs, targetMsgs...)
	}

	lfo.SendControlValue(controlOut, 0)

	return msgs
}
