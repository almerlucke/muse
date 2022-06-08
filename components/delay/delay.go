package delay

// Delay structure
type Delay struct {
	Buffer    []float64
	WriteHead int
}

// NewDelay create a new delay
func NewDelay(length int) *Delay {
	return &Delay{
		Buffer: make([]float64, length),
	}
}

// Write to delay
func (delay *Delay) Write(sample float64) {
	delay.Buffer[delay.WriteHead] = sample
	delay.WriteHead++
	if delay.WriteHead >= len(delay.Buffer) {
		delay.WriteHead = 0
	}
}

// Read from delay, location in samples
func (delay *Delay) Read(location float64) float64 {
	buffer := delay.Buffer
	buflen := len(buffer)

	if location >= float64(buflen) {
		location = float64(buflen - 1)
	}

	sampleLocation := float64(delay.WriteHead) - location

	for sampleLocation < 0.0 {
		sampleLocation += float64(buflen)
	}

	firstIndex := int(sampleLocation)
	fraction := sampleLocation - float64(firstIndex)
	secondIndex := firstIndex + 1

	if firstIndex >= buflen {
		firstIndex -= buflen
	}

	if secondIndex >= buflen {
		secondIndex -= buflen
	}

	v1 := buffer[firstIndex]
	v2 := buffer[secondIndex]

	return v1 + (v2-v1)*fraction
}
