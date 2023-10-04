package muse

import (
	"bufio"
	"log"
	"math"
	"os"

	"github.com/almerlucke/muse/buffer"
	"github.com/almerlucke/muse/io"
	"github.com/dh1tw/gosamplerate"
	"github.com/gordonklaus/portaudio"
)

var DefaultSamplerate = 44100.0
var DefaultBufferSize = 1024

func init() {
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

func NewWithInputs(numInputs, numOutputs int) *Muse {
	e := &Muse{
		BasePatch: NewPatch(numInputs, numOutputs),
	}

	e.SetSelf(e)

	return e
}

func (m *Muse) Synthesize() bool {
	m.PrepareSynthesis()
	return m.BasePatch.Synthesize()
}

func (m *Muse) RenderToSoundFile(filePath string, numSeconds float64, outputSampleRate float64, normalize bool) error {
	inputSampleRate := m.Config.SampleRate
	numChannels := m.NumOutputs()
	buffers := make([]buffer.Buffer, numChannels)

	for c := 0; c < numChannels; c++ {
		buffers[c] = m.OutputAtIndex(c).Buffer
	}

	wr, err := io.NewWriter(filePath, numChannels, m.Config.BufferSize, int(inputSampleRate), int(outputSampleRate), gosamplerate.SRC_SINC_BEST_QUALITY, normalize)
	if err != nil {
		return err
	}

	numCycles := int64(math.Ceil((m.Config.SampleRate * numSeconds) / float64(m.Config.BufferSize)))

	if numCycles > 0 {
		for numCycles > 1 {
			m.Synthesize()
			wr.WriteBuffers(buffers, false)
			numCycles--
		}
		m.Synthesize()
		wr.WriteBuffers(buffers, true)
	}

	wr.Close()

	return nil
}

func (m *Muse) audioCallback(in, out [][]float32) {
	// Prepare synthesis
	m.PrepareSynthesis()

	// Force did synthesize on input thru modules for muse patch,
	// in this way we skip the normal synthesize call
	numInputs := m.NumInputs()

	for i := 0; i < numInputs; i++ {
		m.InputModuleAtIndex(i).SetDidSynthesize(true)
	}

	// Copy system audio input to thru modules output
	for i := 0; i < m.Config.BufferSize; i++ {
		for j := 0; j < numInputs; j++ {
			m.InputModuleAtIndex(j).OutputAtIndex(0).Buffer[i] = float64(in[j][i])
		}
	}

	// Synthesize rest of the patch like normal
	m.BasePatch.Synthesize()

	// Copy outputs to system audio output
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
		1,
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

func (m *Muse) RenderAudio() error {
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

func (m *Muse) PlotControl(ctrl Control, outIndex int, frames int, w float64, h float64, filePath string) error {
	pm := NewPlotModule(frames)

	ctrl.CtrlConnect(outIndex, pm, 0)

	for i := 0; i < frames; i++ {
		m.Synthesize()
	}

	pm.CtrlDisconnect()

	return pm.Save(w, h, true, filePath)
}

func (m *Muse) PlotModule(module Module, outIndex int, frames int, w float64, h float64, filePath string) error {
	pm := NewPlotModule(frames * m.Config.BufferSize)

	m.AddModule(pm)

	module.Connect(outIndex, pm, 0)

	for i := 0; i < frames; i++ {
		m.Synthesize()
	}

	m.RemoveModule(pm)

	return pm.Save(w, h, false, filePath)
}
