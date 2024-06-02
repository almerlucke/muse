package freeverb

import (
	"github.com/almerlucke/muse"
)

/*
   Constants
*/

const (
	denormGuard     = 1e-15
	numcombs        = 8
	numallpasses    = 4
	muted           = 0
	fixedgain       = 0.015
	scalewet        = 3
	scaledry        = 2
	scaledamp       = 0.4
	scaleroom       = 0.28
	offsetroom      = 0.7
	initialroom     = 0.5
	initialdamp     = 0.5
	initialwet      = 0.4 / scalewet
	initialdry      = 0.2
	initialwidth    = 1.0
	initialmode     = 0
	initialfeedback = 0.5
	freezemode      = 0.5
	stereospread    = 23
	combtuningL1    = 1116
	combtuningR1    = combtuningL1 + stereospread
	combtuningL2    = 1188
	combtuningR2    = combtuningL2 + stereospread
	combtuningL3    = 1277
	combtuningR3    = combtuningL3 + stereospread
	combtuningL4    = 1356
	combtuningR4    = combtuningL4 + stereospread
	combtuningL5    = 1422
	combtuningR5    = combtuningL5 + stereospread
	combtuningL6    = 1491
	combtuningR6    = combtuningL6 + stereospread
	combtuningL7    = 1557
	combtuningR7    = combtuningL7 + stereospread
	combtuningL8    = 1617
	combtuningR8    = combtuningL8 + stereospread
	allpasstuningL1 = 556
	allpasstuningR1 = allpasstuningL1 + stereospread
	allpasstuningL2 = 441
	allpasstuningR2 = allpasstuningL2 + stereospread
	allpasstuningL3 = 341
	allpasstuningR3 = allpasstuningL3 + stereospread
	allpasstuningL4 = 225
	allpasstuningR4 = allpasstuningL4 + stereospread
)

type fvAllpass struct {
	buffer   []float64
	bufidx   int
	feedback float64
}

func newAllpass(buflen int, feedback float64) *fvAllpass {
	return &fvAllpass{
		buffer:   make([]float64, buflen),
		feedback: feedback,
	}
}

func (allpass *fvAllpass) mute() {
	for i := 0; i < len(allpass.buffer); i++ {
		allpass.buffer[i] = 0.0
	}
}

func (allpass *fvAllpass) process(input float64) float64 {
	bufout := allpass.buffer[allpass.bufidx] + denormGuard

	output := -input + bufout

	allpass.buffer[allpass.bufidx] = input + (bufout * allpass.feedback)

	allpass.bufidx++

	if allpass.bufidx >= len(allpass.buffer) {
		allpass.bufidx = 0
	}

	return output
}

/*
   Comb
*/

type fvComb struct {
	feedback    float64
	filterstore float64
	damp1       float64
	damp2       float64
	buffer      []float64
	bufidx      int
}

func newComb(buflen int, feedback float64) *fvComb {
	comb := &fvComb{
		buffer:   make([]float64, buflen),
		feedback: feedback,
	}
	comb.setDamp(initialdamp)
	return comb
}

func (c *fvComb) setDamp(val float64) {
	c.damp1 = val
	c.damp2 = 1.0 - val
}

func (c *fvComb) mute() {
	for i := 0; i < len(c.buffer); i++ {
		c.buffer[i] = 0.0
	}
}

func (c *fvComb) process(input float64) float64 {
	output := c.buffer[c.bufidx] + denormGuard
	c.filterstore = output*c.damp2 + c.filterstore*c.damp1
	c.buffer[c.bufidx] = input + c.filterstore*c.feedback
	c.bufidx++
	if c.bufidx >= len(c.buffer) {
		c.bufidx = 0
	}
	return output
}

/*
   Module
*/

// FreeVerb module
type FreeVerb struct {
	*muse.BaseModule
	combL     []*fvComb
	combR     []*fvComb
	allpassL  []*fvAllpass
	allpassR  []*fvAllpass
	gain      float64
	roomsize  float64
	roomsize1 float64
	damp      float64
	damp1     float64
	wet       float64
	wet1      float64
	wet2      float64
	dry       float64
	width     float64
	mode      float64
}

// NewFreeVerbModule generate new freeverb module
func New() *FreeVerb {

	scale := muse.SampleRate() / 44100.0

	fv := &FreeVerb{
		BaseModule: muse.NewBaseModule(2, 2),
	}

	fv.combL = make([]*fvComb, numcombs)
	fv.combR = make([]*fvComb, numcombs)
	fv.combL[0] = newComb(int(combtuningL1*scale), initialfeedback)
	fv.combR[0] = newComb(int(combtuningR1*scale), initialfeedback)
	fv.combL[1] = newComb(int(combtuningL2*scale), initialfeedback)
	fv.combR[1] = newComb(int(combtuningR2*scale), initialfeedback)
	fv.combL[2] = newComb(int(combtuningL3*scale), initialfeedback)
	fv.combR[2] = newComb(int(combtuningR3*scale), initialfeedback)
	fv.combL[3] = newComb(int(combtuningL4*scale), initialfeedback)
	fv.combR[3] = newComb(int(combtuningR4*scale), initialfeedback)
	fv.combL[4] = newComb(int(combtuningL5*scale), initialfeedback)
	fv.combR[4] = newComb(int(combtuningR5*scale), initialfeedback)
	fv.combL[5] = newComb(int(combtuningL6*scale), initialfeedback)
	fv.combR[5] = newComb(int(combtuningR6*scale), initialfeedback)
	fv.combL[6] = newComb(int(combtuningL7*scale), initialfeedback)
	fv.combR[6] = newComb(int(combtuningR7*scale), initialfeedback)
	fv.combL[7] = newComb(int(combtuningL8*scale), initialfeedback)
	fv.combR[7] = newComb(int(combtuningR8*scale), initialfeedback)

	fv.allpassL = make([]*fvAllpass, numallpasses)
	fv.allpassR = make([]*fvAllpass, numallpasses)
	fv.allpassL[0] = newAllpass(int(allpasstuningL1*scale), initialfeedback)
	fv.allpassR[0] = newAllpass(int(allpasstuningR1*scale), initialfeedback)
	fv.allpassL[1] = newAllpass(int(allpasstuningL2*scale), initialfeedback)
	fv.allpassR[1] = newAllpass(int(allpasstuningR2*scale), initialfeedback)
	fv.allpassL[2] = newAllpass(int(allpasstuningL3*scale), initialfeedback)
	fv.allpassR[2] = newAllpass(int(allpasstuningR3*scale), initialfeedback)
	fv.allpassL[3] = newAllpass(int(allpasstuningL4*scale), initialfeedback)
	fv.allpassR[3] = newAllpass(int(allpasstuningR4*scale), initialfeedback)

	fv.SetWet(initialwet)
	fv.SetRoomSize(initialroom)
	fv.SetDry(initialdry)
	fv.SetDamp(initialdamp)
	fv.SetWidth(initialwidth)
	fv.SetMode(initialmode)

	fv.SetSelf(fv)

	return fv
}

func (fv *FreeVerb) SetWet(wet float64) {
	fv.wet = wet * scalewet
	fv.update()
}

func (fv *FreeVerb) SetRoomSize(roomsize float64) {
	fv.roomsize = (roomsize * scaleroom) + offsetroom
	fv.update()
}

func (fv *FreeVerb) SetDry(dry float64) {
	fv.dry = dry * scaledry
}

func (fv *FreeVerb) SetDamp(damp float64) {
	fv.damp = damp * scaledamp
	fv.update()
}

func (fv *FreeVerb) SetWidth(width float64) {
	fv.width = width
	fv.update()
}

func (fv *FreeVerb) SetMode(mode float64) {
	fv.mode = mode
	fv.update()
}

func (fv *FreeVerb) update() {
	fv.wet1 = fv.wet * (fv.width/2.0 + 0.5)
	fv.wet2 = fv.wet * ((1.0 - fv.width) / 2.0)

	if fv.mode >= freezemode {
		fv.roomsize1 = 1
		fv.damp1 = 0
		fv.gain = muted
	} else {
		fv.roomsize1 = fv.roomsize
		fv.damp1 = fv.damp
		fv.gain = fixedgain
	}

	for i := 0; i < numcombs; i++ {
		fv.combL[i].feedback = fv.roomsize1
		fv.combR[i].feedback = fv.roomsize1
		fv.combL[i].setDamp(fv.damp1)
		fv.combR[i].setDamp(fv.damp1)
	}
}

func (fv *FreeVerb) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Wet
		fv.SetWet(value.(float64))
	case 1: // Dry
		fv.SetDry(value.(float64))
	case 2: // RoomSize
		fv.SetRoomSize(value.(float64))
	case 3: // Damp
		fv.SetDamp(value.(float64))
	case 4: // Width
		fv.SetWidth(value.(float64))
	case 5: // Mode
		fv.SetMode(value.(float64))
	}
}

func (fv *FreeVerb) ReceiveMessage(msg any) []*muse.Message {
	if valueMap, ok := msg.(map[string]any); ok {
		if wet, ok := valueMap["wet"].(float64); ok {
			fv.SetWet(wet)
		}
		if roomSize, ok := valueMap["roomSize"].(float64); ok {
			fv.SetRoomSize(roomSize)
		}
		if dry, ok := valueMap["dry"].(float64); ok {
			fv.SetDry(dry)
		}
		if damp, ok := valueMap["damp"].(float64); ok {
			fv.SetDamp(damp)
		}
		if width, ok := valueMap["width"].(float64); ok {
			fv.SetWidth(width)
		}
		if mode, ok := valueMap["mode"].(float64); ok {
			fv.SetMode(mode)
		}
	}

	return nil
}

// DSP for free verb
func (fv *FreeVerb) Synthesize() bool {
	if !fv.BaseModule.Synthesize() {
		return false
	}

	buflen := fv.Config.BufferSize
	outBuffer1 := fv.Outputs[0].Buffer
	outBuffer2 := fv.Outputs[1].Buffer

	var inBuffer1 []float64
	var inBuffer2 []float64

	if fv.Inputs[0].IsConnected() {
		inBuffer1 = fv.Inputs[0].Buffer
	}

	if fv.Inputs[1].IsConnected() {
		inBuffer2 = fv.Inputs[1].Buffer
	}

	for i := 0; i < buflen; i++ {
		outL, outR, inputL, inputR, input := 0.0, 0.0, 0.0, 0.0, 0.0

		if inBuffer1 != nil {
			inputL = inBuffer1[i]
		}

		if inBuffer2 != nil {
			inputR = inBuffer2[i]
		} else {
			inputR = inputL
		}

		input = (inputL + inputR) * fv.gain

		for j := 0; j < numcombs; j++ {
			outL += fv.combL[j].process(input)
			outR += fv.combR[j].process(input)
		}

		for j := 0; j < numallpasses; j++ {
			outL = fv.allpassL[j].process(outL)
			outR = fv.allpassR[j].process(outR)
		}

		outBuffer1[i] = outL*fv.wet1 + outR*fv.wet2 + inputL*fv.dry
		outBuffer2[i] = outR*fv.wet1 + outL*fv.wet2 + inputR*fv.dry
	}

	return true
}
