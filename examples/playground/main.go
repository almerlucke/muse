package main

import (
	"github.com/almerlucke/muse/components/generator"
	"github.com/almerlucke/muse/components/ops"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/mkb218/gosndfile/sndfile"
)

func main() {
	algo := ops.NewDX7Algo1(waveshaping.NewSineTable(2048), 44100.0)
	generator.WriteToFile(algo, "/Users/almerlucke/Desktop/test.aiff", 10.0, 44100, sndfile.SF_FORMAT_AIFF)
}
