package pingpong

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/delay"
	"github.com/almerlucke/muse/utils/timing"
	"math"
)

type PingPong struct {
	*muse.BaseModule
	left    *delay.Delay
	right   *delay.Delay
	read    float64
	newRead float64
	dry     float64
	wet     float64
	fb      float64
}

func New(delayLength float64, read float64, feedback float64, dry float64, wet float64, stereo bool) *PingPong {
	numInputs := 1
	if stereo {
		numInputs = 2
	}

	delayLengthSamps := int(math.Ceil(timing.MilliToSampsf(delayLength, muse.CurrentConfiguration().SampleRate)))

	pp := &PingPong{
		BaseModule: muse.NewBaseModule(numInputs, 2),
		left:       delay.New(delayLengthSamps),
		right:      delay.New(delayLengthSamps),
		read:       read,
		newRead:    read,
		dry:        dry,
		wet:        wet,
		fb:         feedback,
	}

	pp.SetSelf(pp)

	return pp
}

func (pp *PingPong) SetRead(read float64) {
	pp.newRead = read
}

func (pp *PingPong) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Read Location
		pp.SetRead(value.(float64))
	}
}

func (pp *PingPong) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if readLocation, ok := content["location"]; ok {
		pp.SetRead(readLocation.(float64))
	}

	return nil
}

func (pp *PingPong) synthesizeStereo() {
	var (
		inLeft      = pp.Inputs[0].Buffer
		inRight     = pp.Inputs[1].Buffer
		outLeft     = pp.Outputs[0].Buffer
		outRight    = pp.Outputs[1].Buffer
		lookup      = timing.MilliToSampsf(pp.read, pp.Config.SampleRate)
		lookupDelta = 0.0
	)

	if pp.newRead != pp.read {
		lookupDelta = (timing.MilliToSampsf(pp.newRead, pp.Config.SampleRate) - lookup) / float64(pp.Config.BufferSize)
	}

	for i := 0; i < pp.Config.BufferSize; i++ {
		left := pp.left.ReadLinear(lookup)
		right := pp.right.ReadLinear(lookup)

		pp.left.Write(inLeft[i] + right*pp.fb)
		pp.right.Write(inRight[i] + left)

		outLeft[i] = inLeft[i]*pp.dry + left*pp.wet
		outRight[i] = inRight[i]*pp.dry + right*pp.wet

		lookup += lookupDelta
	}

	if pp.newRead != pp.read {
		pp.read = pp.newRead
	}
}

func (pp *PingPong) Synthesize() bool {
	if !pp.BaseModule.Synthesize() {
		return false
	}

	if len(pp.Inputs) == 2 {
		pp.synthesizeStereo()
		return true
	}

	var (
		in          = pp.Inputs[0].Buffer
		outLeft     = pp.Outputs[0].Buffer
		outRight    = pp.Outputs[1].Buffer
		lookup      = timing.MilliToSampsf(pp.read, pp.Config.SampleRate)
		lookupDelta = 0.0
	)

	if pp.newRead != pp.read {
		lookupDelta = (timing.MilliToSampsf(pp.newRead, pp.Config.SampleRate) - lookup) / float64(pp.Config.BufferSize)
	}

	for i := 0; i < pp.Config.BufferSize; i++ {
		left := pp.left.ReadLinear(lookup)
		right := pp.right.ReadLinear(lookup)

		pp.left.Write(in[i] + right*pp.fb)
		pp.right.Write(left)

		outLeft[i] = in[i]*pp.dry + left*pp.wet
		outRight[i] = in[i]*pp.dry + right*pp.wet

		lookup += lookupDelta
	}

	if pp.newRead != pp.read {
		pp.read = pp.newRead
	}

	return true
}
