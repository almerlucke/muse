package main

import "log"

type Sample float64

type Buffer []Sample

type Connection struct {
	Module Module
	Index  int
}

type Socket struct {
	Buffer      Buffer
	Connections []*Connection
}

type Configuration struct {
	SampleRate float64
	BufferSize int
}

type BufferPool struct {
	mainBuffer        Buffer
	inputBuffers      []Buffer
	outputBuffers     []Buffer
	inputBufferIndex  int
	outputBufferIndex int
	bufferSize        int
}

func NewBufferPool(numInputBuffers int, numOutputBuffers int, bufferSize int) *BufferPool {
	pool := BufferPool{
		mainBuffer:    make(Buffer, (numInputBuffers+numOutputBuffers)*bufferSize),
		bufferSize:    bufferSize,
		inputBuffers:  make([]Buffer, numInputBuffers),
		outputBuffers: make([]Buffer, numOutputBuffers),
	}

	offset := 0

	for i := 0; i < numInputBuffers; i++ {
		pool.inputBuffers[i] = pool.mainBuffer[offset : offset+bufferSize]
		offset += bufferSize
	}

	for i := 0; i < numOutputBuffers; i++ {
		pool.outputBuffers[i] = pool.mainBuffer[offset : offset+bufferSize]
		offset += bufferSize
	}

	return &pool
}

func (p *BufferPool) GetInputBuffer() Buffer {
	buffer := p.inputBuffers[p.inputBufferIndex]
	p.inputBufferIndex++
	return buffer
}

func (p *BufferPool) GetOutputBuffer() Buffer {
	buffer := p.outputBuffers[p.outputBufferIndex]
	p.outputBufferIndex++
	return buffer
}

func (p *BufferPool) ClearInputBuffers() {
	for i := 0; i < len(p.inputBuffers)*p.bufferSize; i++ {
		p.mainBuffer[i] = 0
	}
}

func (p *BufferPool) Debug() {
	log.Printf("input buffers %v", len(p.inputBuffers))
	log.Printf("output buffers %v", len(p.outputBuffers))
	log.Printf("main buffer %v", p.mainBuffer)
}

type Module interface {
	GetNumInputs() int
	GetNumOutputs() int
	SetBuffersFromPool(pool *BufferPool)
	TotalInputBuffersNeeded() int
	TotalOutputBuffersNeeded() int
	GetInputAtIndex(index int) *Socket
	GetOutputAtIndex(index int) *Socket
	AddInputConnection(inputIndex int, conn *Connection)
	AddOutputConnection(outputIndex int, conn *Connection)
	HasRun() bool
	PrepareRun()
	Run(config *Configuration)
}

type BaseModule struct {
	inputs  []*Socket
	outputs []*Socket
	hasRun  bool
}

func NewBaseModule(numInputs int, numOutputs int) *BaseModule {
	inputs := make([]*Socket, numInputs)

	for i := 0; i < numInputs; i++ {
		inputs[i] = &Socket{
			Buffer:      nil,
			Connections: []*Connection{},
		}
	}

	outputs := make([]*Socket, numOutputs)

	for i := 0; i < numOutputs; i++ {
		outputs[i] = &Socket{
			Buffer:      nil,
			Connections: []*Connection{},
		}
	}

	return &BaseModule{
		inputs:  inputs,
		outputs: outputs,
	}
}

func (m *BaseModule) GetNumInputs() int {
	return len(m.inputs)
}

func (m *BaseModule) GetNumOutputs() int {
	return len(m.outputs)
}

func (m *BaseModule) SetBuffersFromPool(pool *BufferPool) {
	for _, input := range m.inputs {
		input.Buffer = pool.GetInputBuffer()
	}

	for _, output := range m.outputs {
		output.Buffer = pool.GetOutputBuffer()
	}
}

func (m *BaseModule) TotalInputBuffersNeeded() int {
	return m.GetNumInputs()
}

func (m *BaseModule) TotalOutputBuffersNeeded() int {
	return m.GetNumOutputs()
}

func (m *BaseModule) GetInputAtIndex(index int) *Socket {
	return m.inputs[index]
}

func (m *BaseModule) GetOutputAtIndex(index int) *Socket {
	return m.outputs[index]
}

func (m *BaseModule) AddInputConnection(inputIndex int, conn *Connection) {
	m.inputs[inputIndex].Connections = append(m.inputs[inputIndex].Connections, conn)
}

func (m *BaseModule) AddOutputConnection(outputIndex int, conn *Connection) {
	m.outputs[outputIndex].Connections = append(m.outputs[outputIndex].Connections, conn)
}

func (m *BaseModule) HasRun() bool {
	return m.hasRun
}

func (m *BaseModule) PrepareRun() {
	m.hasRun = false
}

func (m *BaseModule) Run(config *Configuration) {
	if m.hasRun {
		return
	}

	m.hasRun = true

	for _, input := range m.inputs {
		inputBuffer := input.Buffer

		for _, conn := range input.Connections {
			conn.Module.Run(config)

			for bufIndex, sample := range conn.Module.GetOutputAtIndex(conn.Index).Buffer {
				inputBuffer[bufIndex] += sample
			}
		}
	}
}

func Connect(from Module, outIndex int, to Module, inIndex int) {
	from.AddOutputConnection(outIndex, &Connection{Module: to, Index: inIndex})
	to.AddInputConnection(inIndex, &Connection{Module: from, Index: outIndex})
}

type ThruModule struct {
	*BaseModule
}

func NewThruModule() *ThruModule {
	return &ThruModule{
		BaseModule: NewBaseModule(1, 1),
	}
}

func (t *ThruModule) Run(config *Configuration) {
	t.BaseModule.Run(config)

	for index, sample := range t.inputs[0].Buffer {
		t.outputs[0].Buffer[index] = sample + 1.0
	}
}

type Patch struct {
	*BaseModule
	subModules    []Module
	inputModules  []*ThruModule
	outputModules []*ThruModule
}

func NewPatch(numInputs int, numOutputs int) *Patch {
	subModules := []Module{}

	inputModules := make([]*ThruModule, numInputs)
	for i := 0; i < numInputs; i++ {
		inputModules[i] = NewThruModule()
		subModules = append(subModules, inputModules[i])
	}

	outputModules := make([]*ThruModule, numOutputs)
	for i := 0; i < numOutputs; i++ {
		outputModules[i] = NewThruModule()
		subModules = append(subModules, outputModules[i])
	}

	return &Patch{
		BaseModule:    NewBaseModule(0, 0),
		subModules:    subModules,
		inputModules:  inputModules,
		outputModules: outputModules,
	}
}

func (p *Patch) GetNumInputs() int {
	return len(p.inputModules)
}

func (p *Patch) GetNumOutputs() int {
	return len(p.outputModules)
}

func (p *Patch) SetBuffersFromPool(pool *BufferPool) {
	for _, module := range p.subModules {
		module.SetBuffersFromPool(pool)
	}
}

func (p *Patch) TotalInputBuffersNeeded() int {
	totalInputBuffersNeeded := 0

	for _, module := range p.subModules {
		totalInputBuffersNeeded += module.TotalInputBuffersNeeded()
	}

	return totalInputBuffersNeeded
}

func (p *Patch) TotalOutputBuffersNeeded() int {
	totalOutputBuffersNeeded := 0

	for _, module := range p.subModules {
		totalOutputBuffersNeeded += module.TotalOutputBuffersNeeded()
	}

	return totalOutputBuffersNeeded
}

func (p *Patch) AddModule(m Module) {
	p.subModules = append(p.subModules, m)
}

func (p *Patch) PrepareRun() {
	p.BaseModule.PrepareRun()

	for _, module := range p.subModules {
		module.PrepareRun()
	}
}

func (p *Patch) AddInputConnection(inputIndex int, conn *Connection) {
	p.inputModules[inputIndex].AddInputConnection(0, conn)
}

func (p *Patch) AddOutputConnection(outputIndex int, conn *Connection) {
	p.outputModules[outputIndex].AddOutputConnection(0, conn)
}

func (p *Patch) GetInputAtIndex(index int) *Socket {
	return p.inputModules[index].GetInputAtIndex(0)
}

func (p *Patch) GetOutputAtIndex(index int) *Socket {
	return p.outputModules[index].GetOutputAtIndex(0)
}

func main() {
	config := Configuration{
		SampleRate: 44100,
		BufferSize: 12,
	}

	p := NewPatch(0, 0)
	m1 := NewThruModule()
	m2 := NewThruModule()
	m3 := NewThruModule()

	p.AddModule(m1)
	p.AddModule(m2)
	p.AddModule(m3)

	pool := NewBufferPool(p.TotalInputBuffersNeeded(), p.TotalOutputBuffersNeeded(), config.BufferSize)
	p.SetBuffersFromPool(pool)

	Connect(m1, 0, m2, 0)
	Connect(m2, 0, m3, 0)

	p.PrepareRun()
	m3.Run(&config)

	for _, sample := range m3.outputs[0].Buffer {
		log.Printf("%v", sample)
	}

	pool.ClearInputBuffers()
	p.PrepareRun()
	m3.Run(&config)

	for _, sample := range m3.outputs[0].Buffer {
		log.Printf("%v", sample)
	}

	pool.Debug()
}
