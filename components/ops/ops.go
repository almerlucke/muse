// ops fm ops
package ops

type ComponentType int
type FrequencyMode int

const (
	TrackFrequency FrequencyMode = iota
	FixedFrequency
)

type Component interface {
	Run() float64
	PrepareRun()
}

type baseComponent struct {
	run    bool
	output float64
}

type Op struct {
	baseComponent
	table  []float64
	phase  float64
	inc    float64
	fr     float64
	fc     float64
	level  float64
	input  Component
	fcMode FrequencyMode
}

func NewOp(table []float64, fr float64, fc float64, sr float64, level float64, fcMode FrequencyMode) *Op {
	var inc float64

	if fcMode == TrackFrequency {
		inc = (fc * fr) / sr
	} else {
		inc = fc / sr
	}

	return &Op{
		inc:    inc,
		table:  table,
		fr:     fr,
		fc:     fc,
		level:  level,
		fcMode: fcMode,
	}
}

func (op *Op) Connect(input Component) {
	op.input = input
}

func (op *Op) Table() []float64 {
	return op.table
}

func (op *Op) SetTable(table []float64) {
	op.table = table
}

func (op *Op) Level() float64 {
	return op.level
}

func (op *Op) SetLevel(level float64) {
	op.level = level
}

func (op *Op) FrequencyRatio() float64 {
	return op.fr
}

func (op *Op) SetFrequencyRatio(fr float64) {
	op.fr = fr
}

func (op *Op) Frequency() float64 {
	return op.fc
}

func (op *Op) SetFrequency(fc, sr float64) {
	switch op.fcMode {
	case TrackFrequency:
		op.inc = (fc * op.fr) / sr
	case FixedFrequency:
		op.inc = fc / sr
	}
}

func (op *Op) FrequencyMode() FrequencyMode {
	return op.fcMode
}

func (op *Op) SetFrequencyMode(fcMode FrequencyMode) {
	op.fcMode = fcMode
}

func (op *Op) PrepareRun() {
	if !op.run {
		return
	}

	op.run = false

	if op.input != nil {
		op.input.PrepareRun()
	}
}

func (op *Op) Run() float64 {
	if op.run {
		return op.output
	}

	op.run = true

	phaseOffset := 0.0

	if op.input != nil {
		phaseOffset = op.input.Run()
	}

	phase := op.phase + phaseOffset

	for phase >= 1.0 {
		phase -= 1.0
	}

	for phase < 0.0 {
		phase += 1.0
	}

	op.phase += op.inc

	for op.phase >= 1.0 {
		op.phase -= 1.0
	}

	for op.phase < 0.0 {
		op.phase += 1.0
	}

	nf := phase * float64(len(op.table)-1)
	n1 := int(nf)
	fr := nf - float64(n1)
	v1 := op.table[n1]
	v2 := op.table[n1+1]

	op.output = (v1 + fr*(v2-v1)) * op.level

	return op.output
}

type Mixer struct {
	baseComponent
	inputs []Component
}

func NewMixer(inputs ...Component) *Mixer {
	return &Mixer{
		inputs: inputs,
	}
}

func (mix *Mixer) PrepareRun() {
	if !mix.run {
		return
	}

	mix.run = false

	for _, input := range mix.inputs {
		input.PrepareRun()
	}
}

func (m *Mixer) Run() float64 {
	if m.run {
		return m.output
	}

	m.run = true

	mix := 0.0

	for _, input := range m.inputs {
		mix += input.Run()
	}

	m.output = mix

	return m.output
}

type Ops struct {
	baseComponent
	root      Component
	ops       []*Op
	outputVec []float64
}

func NewDX7Algo1(table []float64, fc float64, sr float64) *Ops {
	ops := make([]*Op, 6)

	ops[0] = NewOp(table, 2.01, fc, sr, 0.2, TrackFrequency)
	ops[1] = NewOp(table, 1.02, fc, sr, 0.5, TrackFrequency)
	ops[2] = NewOp(table, 4.01, fc, sr, 0.2, TrackFrequency)
	ops[3] = NewOp(table, 3.02, fc, sr, 0.1, TrackFrequency)
	ops[4] = NewOp(table, 1.53, fc, sr, 0.2, TrackFrequency)
	ops[5] = NewOp(table, 3.01, fc, sr, 0.5, TrackFrequency)

	ops[1].Connect(ops[0]) // ops 1 to ops 2
	ops[2].Connect(ops[2]) // ops 3 to ops 3 feedback
	ops[3].Connect(ops[2]) // ops 3 to ops 4
	ops[4].Connect(ops[3]) // ops 4 to ops 5
	ops[5].Connect(ops[4]) // ops 5 to ops 6

	mix := NewMixer(ops[1], ops[5])

	return &Ops{
		root:      mix,
		ops:       ops,
		outputVec: []float64{0.0},
	}
}

func (ops *Ops) SetFrequency(fc float64, sr float64) {
	for _, op := range ops.ops {
		if op.FrequencyMode() == TrackFrequency {
			op.SetFrequency(fc, sr)
		}
	}
}

func (ops *Ops) Operator(index int) *Op {
	return ops.ops[index]
}

func (ops *Ops) PrepareRun() {
	if !ops.run {
		return
	}

	ops.run = false

	ops.root.PrepareRun()
}

func (ops *Ops) Run() float64 {
	if ops.run {
		return ops.output
	}

	ops.run = true

	ops.output = ops.root.Run()

	return ops.output
}

/*
// Generator interface
*/

func (ops *Ops) NumDimensions() int {
	return 1
}

func (ops *Ops) Generate() []float64 {
	ops.PrepareRun()
	ops.outputVec[0] = ops.Run()
	return ops.outputVec
}
