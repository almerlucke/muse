package muse

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/almerlucke/muse/io"
	"github.com/gordonklaus/portaudio"
	"github.com/mkb218/gosndfile/sndfile"
)

var DefaultSamplerate = 44100.0
var DefaultBufferSize = 1024

func init() {
	rand.Seed(time.Now().UnixNano())

	configurationInit()
}

type Muse struct {
	*BasePatch
	stream *portaudio.Stream
}

func New(numOutputs int) *Muse {
	e := &Muse{
		BasePatch: NewPatch(0, numOutputs),
	}

	e.SetSelf(e)

	return e
}

func (m *Muse) Synthesize() bool {
	m.PrepareSynthesis()
	return m.BasePatch.Synthesize()
}

func (m *Muse) RenderToSoundFile(filePath string, numSeconds float64, outputSampleRate float64, normalizeOutput bool, format sndfile.Format) error {
	numChannels := m.NumOutputs()

	swr := io.NewSoundWriter(numChannels, int(m.Config.SampleRate), int(outputSampleRate), normalizeOutput)

	interleaveBuffer := make([]float64, m.NumOutputs()*m.Config.BufferSize)

	framesToProduce := int64(m.Config.SampleRate * numSeconds)

	for framesToProduce > 0 {
		m.Synthesize()

		interleaveIndex := 0

		numFrames := m.Config.BufferSize

		if framesToProduce <= int64(m.Config.BufferSize) {
			numFrames = int(framesToProduce)
		}

		for i := 0; i < numFrames; i++ {
			for c := 0; c < numChannels; c++ {
				interleaveBuffer[interleaveIndex] = m.OutputAtIndex(c).Buffer[i]
				interleaveIndex++
			}
		}

		swr.WriteFrames(interleaveBuffer[:numFrames*numChannels])

		framesToProduce -= int64(numFrames)
	}

	return swr.Finish(filePath, format)
}

func (m *Muse) audioCallback(in, out [][]float32) {
	m.Synthesize()

	numOutputs := m.NumOutputs()

	for i := 0; i < m.Config.BufferSize; i++ {
		for j := 0; j < numOutputs; j++ {
			out[j][i] = float32(m.OutputAtIndex(j).Buffer[i])
		}
	}
}

func (m *Muse) InitializeAudio() error {
	portaudio.Initialize()

	stream, err := portaudio.OpenDefaultStream(
		2,
		m.NumOutputs(),
		m.Config.SampleRate,
		m.Config.BufferSize,
		m.audioCallback,
	)

	if err != nil {
		portaudio.Terminate()
		return err
	}

	m.stream = stream

	return nil
}

func (m *Muse) StartAudio() error {
	return m.stream.Start()
}

func (m *Muse) StopAudio() error {
	return m.stream.Stop()
}

func (m *Muse) TerminateAudio() {
	if m.stream != nil {
		m.stream.Close()
		m.stream = nil
	}

	portaudio.Terminate()
}

func (m *Muse) RenderLive() error {
	err := m.InitializeAudio()
	if err != nil {
		return err
	}

	defer m.TerminateAudio()

	m.StartAudio()

	log.Printf("Press enter to quit...")

	reader := bufio.NewReader(os.Stdin)

	reader.ReadString('\n')

	m.StopAudio()

	return nil
}

/*
func (e *Environment) PlotControl(ctrl Control, outIndex int, frames int, w float64, h float64, filePath string) error {
	pm := NewPlotModule(frames, e.Config)

	ctrl.CtrlConnect(outIndex, pm, 0)

	for i := 0; i < frames; i++ {
		e.Synthesize()
	}

	pm.CtrlDisconnect()

	return pm.Save(w, h, true, filePath)
}

func (e *Environment) PlotModule(module Module, outIndex int, frames int, w float64, h float64, filePath string) error {
	pm := NewPlotModule(frames*e.Config.BufferSize, e.Config)

	e.AddModule(pm)

	module.Connect(outIndex, pm, 0)

	for i := 0; i < frames; i++ {
		e.Synthesize()
	}

	e.RemoveModule(pm)

	return pm.Save(w, h, false, filePath)
}
*/
