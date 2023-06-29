// ops fm ops
package ops

type ComponentType int

type Component interface {
	Run() float64
	PrepareRun()
}

type baseComponent struct {
	run    bool
	output float64
}

func (bc *baseComponent) PrepareRun() {
	bc.run = false
}

type Connection struct {
	Component Component
	Level     float64
}

func NewConnection(component Component, level float64) *Connection {
	return &Connection{
		Component: component,
		Level:     level,
	}
}

type Op struct {
	baseComponent
	table []float64
	phase float64
	inc   float64
	input *Connection
}

func NewOp(table []float64, fc float64, sr float64) *Op {
	return &Op{
		inc:   fc / sr,
		table: table,
	}
}

func (op *Op) Connect(input *Connection) {
	op.input = input
}

func (op *Op) SetTable(table []float64) {
	op.table = table
}

func (op *Op) SetFrequency(fc, sr float64) {
	op.inc = fc / sr
}

func (op *Op) PrepareRun() {
	if !op.run {
		return
	}

	op.baseComponent.PrepareRun()

	if op.input != nil {
		op.input.Component.PrepareRun()
	}
}

func (op *Op) Run() float64 {
	if op.run {
		return op.output
	}

	op.run = true

	phaseOffset := 0.0

	if op.input != nil {
		phaseOffset = op.input.Component.Run() * op.input.Level
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

	op.output = v1 + fr*(v2-v1)

	return op.output
}

type Mixer struct {
	baseComponent
	inputs []*Connection
}

func NewMixer(inputs ...*Connection) *Mixer {
	return &Mixer{
		inputs: inputs,
	}
}

func (mix *Mixer) PrepareRun() {
	if !mix.run {
		return
	}

	mix.baseComponent.PrepareRun()

	for _, input := range mix.inputs {
		input.Component.PrepareRun()
	}
}

func (m *Mixer) Run() float64 {
	if m.run {
		return m.output
	}

	m.run = true

	mix := 0.0

	for _, input := range m.inputs {
		mix += input.Component.Run() * input.Level
	}

	m.output = mix

	return m.output
}

type Algorithm struct {
	baseComponent
	Root        Component
	Connections []*Connection
	ParamMap    map[string]*Connection
	Ops         []*Op
	OutputOps   []*Op
	outputVec   []float64
}

func (algo *Algorithm) PrepareRun() {
	if !algo.run {
		return
	}

	algo.baseComponent.PrepareRun()
	algo.Root.PrepareRun()
}

func (algo *Algorithm) Run() float64 {
	if algo.run {
		return algo.output
	}

	algo.run = true

	algo.output = algo.Root.Run()

	return algo.output
}

func (algo *Algorithm) NumDimensions() int {
	return 1
}

func (algo *Algorithm) Generate() []float64 {
	algo.PrepareRun()
	algo.outputVec[0] = algo.Run()
	return algo.outputVec
}

func NewDX7Algo1(table []float64, sr float64) *Algorithm {
	ops := make([]*Op, 6)
	ops[0] = NewOp(table, 300.0, sr)
	ops[1] = NewOp(table, 160.0, sr)
	ops[2] = NewOp(table, 300.2, sr)
	ops[3] = NewOp(table, 152.0, sr)
	ops[4] = NewOp(table, 153.0, sr)
	ops[5] = NewOp(table, 140.0, sr)

	outputOps := make([]*Op, 2)
	outputOps[0] = ops[1]
	outputOps[1] = ops[5]

	connections := make([]*Connection, 7)
	connections[0] = NewConnection(ops[0], 0.3)
	connections[1] = NewConnection(ops[2], 0.1)
	connections[2] = NewConnection(ops[2], 0.2)
	connections[3] = NewConnection(ops[3], 0.3)
	connections[4] = NewConnection(ops[4], 0.3)
	connections[5] = NewConnection(ops[1], 0.5)
	connections[6] = NewConnection(ops[5], 0.5)

	ops[1].Connect(connections[0]) // ops 1 to ops 2
	ops[2].Connect(connections[1]) // ops 3 feedback
	ops[3].Connect(connections[2]) // ops 3 to ops 4
	ops[4].Connect(connections[3]) // ops 4 to ops 5
	ops[5].Connect(connections[4]) // ops 5 to ops 6

	mix := NewMixer(connections[5], connections[6])

	paramMap := map[string]*Connection{
		"op1-2-level":        connections[0],
		"op3-feedback-level": connections[1],
		"op3-4-level":        connections[2],
		"op4-5-level":        connections[3],
		"op5-6-level":        connections[4],
		"op2-mix-level":      connections[5],
		"op6-mix-level":      connections[6],
	}

	ops[1].SetFrequency(160, sr)
	ops[5].SetFrequency(150, sr)

	return &Algorithm{
		Root:        mix,
		Connections: connections,
		ParamMap:    paramMap,
		Ops:         ops,
		OutputOps:   outputOps,
		outputVec:   []float64{0.0},
	}
}
