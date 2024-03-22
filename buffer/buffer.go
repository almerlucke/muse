package buffer

import (
	"github.com/almerlucke/sndfile/writer"
)

type Buffer []float64

func (b Buffer) Clear() {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

func (b Buffer) Lookup(pos float64, wrap bool) float64 {
	l := len(b)
	i1 := int(pos)
	i2 := i1 + 1
	fr := pos - float64(i1)

	if wrap {
		i2 = i2 % l
	} else {
		if i2 >= l {
			i2 = l - 1
		}
	}

	s1 := b[i1]

	return s1 + (b[i2]-s1)*fr
}

type WriterConverter struct {
	conv    *writer.ChannelConverter[float64]
	buffers [][]float64
}

func NewWriterConverter(frameSize int, numChannels int) *WriterConverter {
	return &WriterConverter{
		conv:    writer.NewChannelConverter[float64](frameSize, numChannels),
		buffers: make([][]float64, numChannels),
	}
}

func (c *WriterConverter) Convert(input any) []float32 {
	for i, b := range input.([]Buffer) {
		c.buffers[i] = b
	}

	return c.conv.Convert(c.buffers)
}

func (c *WriterConverter) FrameSize() int {
	return c.conv.FrameSize()
}
