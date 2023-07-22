package generator

import (
	"github.com/almerlucke/muse/io"
	"github.com/mkb218/gosndfile/sndfile"
)

type Generator interface {
	NumDimensions() int
	Generate() []float64
}

func WriteToSndFile(gen Generator, filePath string, seconds float64, sr int, format sndfile.Format) error {
	wr := io.NewSoundWriter(gen.NumDimensions(), sr, sr, false)

	numFrames := int64(seconds * float64(sr))
	for numFrames > 0 {
		frame := gen.Generate()
		wr.WriteFrames(frame)
		numFrames--
	}

	return wr.Finish(filePath, format)
}
