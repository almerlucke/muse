package muse

// type Environment struct {
// 	*BasePatch
// 	stream *portaudio.Stream
// }

// func NewEnvironment(numOutputs int) *Environment {
// 	e := &Environment{
// 		BasePatch: NewPatch(0, numOutputs),
// 	}

// 	e.SetSelf(e)

// 	return e
// }

// func (e *Environment) Synthesize() bool {
// 	e.PrepareSynthesis()
// 	return e.BasePatch.Synthesize()
// }

// func (e *Environment) SynthesizeToFile(filePath string, numSeconds float64, outputSampleRate float64, normalizeOutput bool, format sndfile.Format) error {
// 	numChannels := e.NumOutputs()

// 	swr := io.NewSoundWriter(numChannels, int(e.Config.SampleRate), int(outputSampleRate), normalizeOutput)

// 	interleaveBuffer := make([]float64, e.NumOutputs()*e.Config.BufferSize)

// 	framesToProduce := int64(e.Config.SampleRate * numSeconds)

// 	for framesToProduce > 0 {
// 		e.Synthesize()

// 		interleaveIndex := 0

// 		numFrames := e.Config.BufferSize

// 		if framesToProduce <= int64(e.Config.BufferSize) {
// 			numFrames = int(framesToProduce)
// 		}

// 		for i := 0; i < numFrames; i++ {
// 			for c := 0; c < numChannels; c++ {
// 				interleaveBuffer[interleaveIndex] = e.OutputAtIndex(c).Buffer[i]
// 				interleaveIndex++
// 			}
// 		}

// 		swr.WriteFrames(interleaveBuffer[:numFrames*numChannels])

// 		framesToProduce -= int64(numFrames)
// 	}

// 	return swr.Finish(filePath, format)
// }

// func (e *Environment) PortaudioStream() (*portaudio.Stream, error) {
// 	return portaudio.OpenDefaultStream(
// 		1,
// 		e.NumOutputs(),
// 		e.Config.SampleRate,
// 		e.Config.BufferSize,
// 		e.portaudioCallback,
// 	)
// }

// func (e *Environment) portaudioCallback(in, out [][]float32) {
// 	e.Synthesize()

// 	numOutputs := e.NumOutputs()

// 	for i := 0; i < e.Config.BufferSize; i++ {
// 		for j := 0; j < numOutputs; j++ {
// 			out[j][i] = float32(e.OutputAtIndex(j).Buffer[i])
// 		}
// 	}
// }

// func (e *Environment) InitializeAudio() (*portaudio.Stream, error) {
// 	portaudio.Initialize()

// 	stream, err := e.PortaudioStream()
// 	if err != nil {
// 		portaudio.Terminate()
// 		return nil, err
// 	}

// 	e.stream = stream

// 	return e.stream, nil
// }

// func (e *Environment) TerminateAudio() {
// 	if e.stream != nil {
// 		e.stream.Close()
// 		e.stream = nil
// 	}

// 	portaudio.Terminate()
// }

// func (e *Environment) QuickPlayAudio() error {
// 	stream, err := e.InitializeAudio()
// 	if err != nil {
// 		return err
// 	}

// 	defer e.TerminateAudio()

// 	stream.Start()

// 	log.Printf("Press enter to quit...")

// 	reader := bufio.NewReader(os.Stdin)

// 	reader.ReadString('\n')

// 	return nil
// }
