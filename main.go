package main

import (
	"fmt"
	"log"
)

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

func (s *Socket) IsConnected() bool {
	return len(s.Connections) > 0
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
	SetIdentifier(identifier string)
	Identifier() string
	NumInputs() int
	NumOutputs() int
	SetBuffersFromPool(pool *BufferPool)
	TotalInputBuffersNeeded() int
	TotalOutputBuffersNeeded() int
	InputAtIndex(index int) *Socket
	OutputAtIndex(index int) *Socket
	AddInputConnection(inputIndex int, conn *Connection)
	AddOutputConnection(outputIndex int, conn *Connection)
	HasRun() bool
	PrepareRun()
	Run(config *Configuration)
}

type BaseModule struct {
	inputs     []*Socket
	outputs    []*Socket
	hasRun     bool
	identifier string
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

func (m *BaseModule) Identifier() string {
	return m.identifier
}

func (m *BaseModule) SetIdentifier(identifier string) {
	m.identifier = identifier
}

func (m *BaseModule) NumInputs() int {
	return len(m.inputs)
}

func (m *BaseModule) NumOutputs() int {
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
	return m.NumInputs()
}

func (m *BaseModule) TotalOutputBuffersNeeded() int {
	return m.NumOutputs()
}

func (m *BaseModule) InputAtIndex(index int) *Socket {
	return m.inputs[index]
}

func (m *BaseModule) OutputAtIndex(index int) *Socket {
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

			for bufIndex, sample := range conn.Module.OutputAtIndex(conn.Index).Buffer {
				inputBuffer[bufIndex] += sample
			}
		}
	}
}

func Connect(from Module, outIndex int, to Module, inIndex int) {
	pf, ok := from.(*Patch)
	if ok {
		from = pf.OutputModuleAtIndex(outIndex)
		outIndex = 0
	}

	pt, ok := to.(*Patch)
	if ok {
		to = pt.InputModuleAtIndex(inIndex)
		inIndex = 0
	}

	from.AddOutputConnection(outIndex, &Connection{Module: to, Index: inIndex})
	to.AddInputConnection(inIndex, &Connection{Module: from, Index: outIndex})
}

func DebugConnections(module Module) {
	log.Printf("Module %s", module.Identifier())
	for i := 0; i < module.NumInputs(); i++ {
		input := module.InputAtIndex(i)
		log.Printf("Input %d", i+1)
		log.Printf("Buffer %v", input.Buffer)

		for _, conn := range input.Connections {
			log.Printf("Connection to output %d of module %s", conn.Index+1, conn.Module.Identifier())
		}

		for _, conn := range input.Connections {
			DebugConnections(conn.Module)
		}
	}
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

	for i := 0; i < config.BufferSize; i++ {
		t.outputs[0].Buffer[i] = t.inputs[0].Buffer[i] + 1
	}
}

type Patch struct {
	*BaseModule
	subModules    []Module
	inputModules  []*ThruModule
	outputModules []*ThruModule
}

func NewPatch(numInputs int, numOutputs int, identifier string) *Patch {
	subModules := []Module{}

	inputModules := make([]*ThruModule, numInputs)
	for i := 0; i < numInputs; i++ {
		inputModules[i] = NewThruModule()
		inputModules[i].SetIdentifier(fmt.Sprintf("Patch_%v_input_%v", identifier, i+1))
		subModules = append(subModules, inputModules[i])
	}

	outputModules := make([]*ThruModule, numOutputs)
	for i := 0; i < numOutputs; i++ {
		outputModules[i] = NewThruModule()
		outputModules[i].SetIdentifier(fmt.Sprintf("Patch_%v_output_%v", identifier, i+1))
		subModules = append(subModules, outputModules[i])
	}

	p := &Patch{
		BaseModule:    NewBaseModule(0, 0),
		subModules:    subModules,
		inputModules:  inputModules,
		outputModules: outputModules,
	}

	p.SetIdentifier(identifier)

	return p
}

func (p *Patch) NumInputs() int {
	return len(p.inputModules)
}

func (p *Patch) NumOutputs() int {
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
	log.Printf("patch AddInputConnection %s", p.Identifier())
	p.inputModules[inputIndex].AddInputConnection(0, conn)
}

func (p *Patch) AddOutputConnection(outputIndex int, conn *Connection) {
	log.Printf("patch AddOutputConnection %s", p.Identifier())
	p.outputModules[outputIndex].AddOutputConnection(0, conn)
}

func (p *Patch) InputModuleAtIndex(index int) Module {
	return p.inputModules[index]
}

func (p *Patch) OutputModuleAtIndex(index int) Module {
	return p.outputModules[index]
}

func (p *Patch) InputAtIndex(index int) *Socket {
	return p.inputModules[index].InputAtIndex(0)
}

func (p *Patch) OutputAtIndex(index int) *Socket {
	return p.outputModules[index].OutputAtIndex(0)
}

func main() {
	config := Configuration{
		SampleRate: 44100,
		BufferSize: 12,
	}

	op := NewPatch(0, 0, "op_patch")

	ip := NewPatch(1, 1, "ip_patch")
	it1 := NewThruModule()
	it1.SetIdentifier("it1")

	ip.AddModule(it1)
	Connect(ip.InputModuleAtIndex(0), 0, it1, 0)
	Connect(it1, 0, ip.OutputModuleAtIndex(0), 0)

	t1 := NewThruModule()
	t1.SetIdentifier("t1")
	t2 := NewThruModule()
	t2.SetIdentifier("t2")

	op.AddModule(t1)
	op.AddModule(t2)
	op.AddModule(ip)

	Connect(t1, 0, ip, 0)
	Connect(ip, 0, t2, 0)

	pool := NewBufferPool(op.TotalInputBuffersNeeded(), op.TotalOutputBuffersNeeded(), config.BufferSize)
	op.SetBuffersFromPool(pool)

	op.PrepareRun()
	t2.Run(&config)

	DebugConnections(t2)
	DebugConnections(ip.OutputModuleAtIndex(0))

	// for _, sample := range ip.OutputModuleAtIndex(0).OutputAtIndex(0).Buffer {
	// 	log.Printf("%v", sample)
	// }

	// pool.ClearInputBuffers()
	// ip.PrepareRun()
	// ip.OutputModuleAtIndex(0).Run(&config)

	// for _, sample := range ip.OutputModuleAtIndex(0).OutputAtIndex(0).Buffer {
	// 	log.Printf("%v", sample)
	// }

	// pool.Debug()
}
