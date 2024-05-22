package flanger

import (
	"github.com/almerlucke/genny/float/line"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/delay"
	"github.com/almerlucke/muse/utils/timing"
	"math"
)

const (
	FlangeMinPos = 2.0
	FlangeMaxPos = 20.0
)

type Flanger struct {
	*muse.BaseModule
	delayLeft  *delay.Delay
	delayRight *delay.Delay
	fb         float64
	depth      float64
	newDepth   float64
	dry        float64
	wet        float64
}

func New(depth float64, feedback float64, dry float64, wet float64, stereo bool) *Flanger {
	numInputs := 1
	if stereo {
		numInputs = 2
	}

	delaySize := int(math.Ceil(timing.MilliToSampsf(FlangeMaxPos, muse.CurrentConfiguration().SampleRate)))

	f := &Flanger{
		BaseModule: muse.NewBaseModule(numInputs, numInputs),
		delayLeft:  delay.New(delaySize),
		delayRight: delay.New(delaySize),
		depth:      depth,
		fb:         feedback,
		dry:        dry,
		wet:        wet,
	}

	f.SetSelf(f)

	return f
}

func (f *Flanger) SetDepth(depth float64) {
	f.newDepth = depth
}

func (f *Flanger) ReceiveControlValue(value any, index int) {
	switch index {
	case 0:
		f.SetDepth(value.(float64))
	}
}

func (f *Flanger) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if depth, ok := content["depth"]; ok {
		f.SetDepth(depth.(float64))
	}

	return nil
}

func (f *Flanger) synthesizeStereo() {
	var (
		inBufLeft   = f.Inputs[0].Buffer
		inBufRight  = f.Inputs[1].Buffer
		outBufLeft  = f.Outputs[0].Buffer
		outBufRight = f.Outputs[1].Buffer
		readLine    line.Line
	)

	readLine.From(timing.MilliToSampsf(f.depth*(FlangeMaxPos-FlangeMinPos)+FlangeMinPos, f.Config.SampleRate))
	if f.depth != f.newDepth {
		readLine.To(timing.MilliToSampsf(f.newDepth*(FlangeMaxPos-FlangeMinPos)+FlangeMinPos, f.Config.SampleRate), f.Config.BufferSize)
		f.depth = f.newDepth
	}

	for i := 0; i < f.Config.BufferSize; i++ {
		inLeft := inBufLeft[i]
		inRight := inBufRight[i]
		readPos := readLine.Generate()
		delOutLeft := f.delayLeft.ReadLinear(readPos)
		delOutRight := f.delayRight.ReadLinear(readPos)
		f.delayLeft.Write(inLeft + delOutLeft*f.fb)
		f.delayRight.Write(inRight + delOutRight*f.fb)
		flangOutLeft := inLeft + delOutLeft
		flangOutRight := inRight + delOutRight
		outBufLeft[i] = inLeft*f.dry + flangOutLeft*f.wet
		outBufRight[i] = inRight*f.dry + flangOutRight*f.wet
	}
}

func (f *Flanger) Synthesize() bool {
	if !f.BaseModule.Synthesize() {
		return false
	}

	if len(f.Inputs) == 2 {
		f.synthesizeStereo()
		return true
	}

	var (
		inBuf    = f.Inputs[0].Buffer
		outBuf   = f.Outputs[0].Buffer
		readLine line.Line
	)

	readLine.From(timing.MilliToSampsf(f.depth*(FlangeMaxPos-FlangeMinPos)+FlangeMinPos, f.Config.SampleRate))
	if f.depth != f.newDepth {
		readLine.To(timing.MilliToSampsf(f.newDepth*(FlangeMaxPos-FlangeMinPos)+FlangeMinPos, f.Config.SampleRate), f.Config.BufferSize)
		f.depth = f.newDepth
	}

	for i := range f.Config.BufferSize {
		in := inBuf[i]
		readPos := readLine.Generate()
		delOut := f.delayLeft.ReadLinear(readPos)
		f.delayLeft.Write(in + delOut*f.fb)
		flangOut := in + delOut
		outBuf[i] = in*f.dry + flangOut*f.wet
	}

	return true
}
