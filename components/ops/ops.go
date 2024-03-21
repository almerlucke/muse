// Package ops fm ops
package ops

type FrequencyMode int
type OperatorType int

const (
	TrackFrequency FrequencyMode = iota
	FixedFrequency
)

const (
	Modulator OperatorType = iota
	Carrier
)

var DefaultEnvLevels = [4]float64{1.0, 0.35, 0.35, 0.0}
var DefaultEnvRates = [4]float64{0.95, 0.95, 0.8, 0.85}
var DefaultPitchEnvLevels = [4]float64{0.5, 0.5, 0.5, 0.5}
var DefaultPitchEnvRates = [4]float64{0.95, 0.8, 0.8, 0.95}

type Component interface {
	Run() float64
	PrepareRun()
	Connect(Component)
}

type baseComponent struct {
	run    bool
	output float64
}

type Op struct {
	baseComponent
	table    []float64
	phase    float64
	inc      float64
	fr       float64
	fc       float64
	level    float64
	opType   OperatorType
	input    Component
	fcMode   FrequencyMode
	levelEnv *Envelope
}

func NewOp(table []float64, fr float64, fc float64, sr float64, level float64, fcMode FrequencyMode) *Op {
	var inc float64

	if fcMode == TrackFrequency {
		inc = (fc * fr) / sr
	} else {
		inc = fc / sr
	}

	return &Op{
		inc:      inc,
		table:    table,
		fr:       fr,
		fc:       fc,
		level:    level,
		fcMode:   fcMode,
		levelEnv: NewEnvelope(DefaultEnvLevels, DefaultEnvRates, sr, EnvelopeNoteOffRelease),
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

func (op *Op) LevelEnvelope() *Envelope {
	return op.levelEnv
}

func (op *Op) SetOperatorType(opType OperatorType) {
	op.opType = opType
}

func (op *Op) OperatorType() OperatorType {
	return op.opType
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

func (op *Op) SetFrequencyRatio(fr float64, sr float64) {
	op.fr = fr
	op.SetFrequency(op.fc, sr)
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

	op.fc = fc
}

func (op *Op) FrequencyMode() FrequencyMode {
	return op.fcMode
}

func (op *Op) SetFrequencyMode(fcMode FrequencyMode, sr float64) {
	op.fcMode = fcMode
	op.SetFrequency(op.fc, sr)
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

	op.output = (v1 + fr*(v2-v1)) * op.level * op.levelEnv.Tick()

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

func (mix *Mixer) Connect(component Component) {
	mix.inputs = append(mix.inputs, component)
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

func (mix *Mixer) Run() float64 {
	if mix.run {
		return mix.output
	}

	mix.run = true

	m := 0.0

	for _, input := range mix.inputs {
		m += input.Run()
	}

	mix.output = m

	return mix.output
}

type Ops struct {
	baseComponent
	root      Component
	ops       []*Op
	outputVec []float64
	pitchEnv  *Envelope
	level     float64
	fc        float64
	sr        float64
}

func NewOps(table []float64, fc float64, sr float64) *Ops {
	ops := make([]*Op, 6)

	ops[0] = NewOp(table, 2.01, fc, sr, 0.2, TrackFrequency)
	ops[1] = NewOp(table, 1.02, fc, sr, 0.5, TrackFrequency)
	ops[2] = NewOp(table, 4.01, fc, sr, 0.2, TrackFrequency)
	ops[3] = NewOp(table, 3.02, fc, sr, 0.1, TrackFrequency)
	ops[4] = NewOp(table, 1.53, fc, sr, 0.2, TrackFrequency)
	ops[5] = NewOp(table, 3.01, fc, sr, 0.5, TrackFrequency)

	opsObj := &Ops{
		root:      nil,
		ops:       ops,
		outputVec: []float64{0.0},
		pitchEnv:  NewEnvelope(DefaultPitchEnvLevels, DefaultPitchEnvRates, sr, EnvelopeNoteOffRelease),
		sr:        sr,
		fc:        fc,
		level:     1.0,
	}

	opsObj.Apply(DX7Algo1)

	return opsObj
}

func (ops *Ops) ResetAlgo() {
	for _, op := range ops.ops {
		op.Connect(nil)
		op.opType = Modulator
	}
	ops.root = ops.ops[0]
}

func (ops *Ops) Apply(algo *Algo) {
	ops.ResetAlgo()

	objects := map[string]Component{}

	objects["op1"] = ops.ops[0]
	objects["op2"] = ops.ops[1]
	objects["op3"] = ops.ops[2]
	objects["op4"] = ops.ops[3]
	objects["op5"] = ops.ops[4]
	objects["op6"] = ops.ops[5]

	for id, opType := range algo.OpTypes {
		if opType == "carrier" {
			objects[id].(*Op).SetOperatorType(Carrier)
		} else {
			objects[id].(*Op).SetOperatorType(Modulator)
		}
	}

	for _, obj := range algo.Objects {
		if obj["type"] == "mix" {
			objects[obj["id"]] = NewMixer()
		}
	}

	for _, connection := range algo.Connections {
		objects[connection[1]].Connect(objects[connection[0]])
	}

	ops.root = objects[algo.Root]
}

func (ops *Ops) SetReleaseMode(releaseMode EnvelopeReleaseMode) {
	ops.pitchEnv.SetReleaseMode(releaseMode)
	for _, op := range ops.ops {
		op.levelEnv.SetReleaseMode(releaseMode)
	}
}

func (ops *Ops) PitchEnvelope() *Envelope {
	return ops.pitchEnv
}

func (ops *Ops) NoteOn(fc float64, level float64, duration float64) {
	ops.fc = fc
	ops.level = level

	for _, op := range ops.ops {
		if op.FrequencyMode() == TrackFrequency {
			op.SetFrequency(fc, ops.sr)
		}
	}

	ops.pitchEnv.TriggerHard(duration)
	for _, op := range ops.ops {
		op.levelEnv.TriggerHard(duration)
	}
}

func (ops *Ops) NoteOff() {
	ops.pitchEnv.NoteOff()
	for _, op := range ops.ops {
		op.LevelEnvelope().NoteOff()
	}
}

func (ops *Ops) Idle() bool {
	for _, op := range ops.ops {
		if op.opType == Carrier && !op.levelEnv.Idle() {
			return false
		}
	}

	return true
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

	fc := ops.pitchEnv.Tick() * 2.0 * ops.fc

	for _, op := range ops.ops {
		if op.FrequencyMode() == TrackFrequency {
			op.SetFrequency(fc, ops.sr)
		}
	}

	ops.output = ops.root.Run() * ops.level

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
