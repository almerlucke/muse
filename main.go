package main

import (
	"fmt"
	"log"
	"strings"
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
	Run(config *Configuration) bool
}

type BaseModule struct {
	inputs     []*Socket
	outputs    []*Socket
	hasRun     bool
	identifier string
}

func NewBaseModule(numInputs int, numOutputs int, identifier string) *BaseModule {
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
		inputs:     inputs,
		outputs:    outputs,
		identifier: identifier,
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

func (m *BaseModule) Run(config *Configuration) bool {
	if m.hasRun {
		return false
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

	return true
}

type ThruModule struct {
	*BaseModule
}

func NewThruModule(identifier string) *ThruModule {
	return &ThruModule{
		BaseModule: NewBaseModule(1, 1, identifier),
	}
}

func (t *ThruModule) Run(config *Configuration) bool {
	if !t.BaseModule.Run(config) {
		return false
	}

	for i := 0; i < config.BufferSize; i++ {
		t.outputs[0].Buffer[i] = t.inputs[0].Buffer[i]
	}

	return true
}

type Patch interface {
	Module
	AddModule(m Module)
	Contains(m Module) bool
	Lookup(address string) Module
	InputModuleAtIndex(index int) Module
	OutputModuleAtIndex(index int) Module
}

type BasePatch struct {
	*BaseModule
	subModules    []Module
	inputModules  []*ThruModule
	outputModules []*ThruModule
	addressMap    map[string]Module
}

func NewPatch(numInputs int, numOutputs int, identifier string) *BasePatch {
	subModules := []Module{}

	inputModules := make([]*ThruModule, numInputs)
	for i := 0; i < numInputs; i++ {
		inputModules[i] = NewThruModule(fmt.Sprintf("patch_%s_input_%d", identifier, i+1))
		subModules = append(subModules, inputModules[i])
	}

	outputModules := make([]*ThruModule, numOutputs)
	for i := 0; i < numOutputs; i++ {
		outputModules[i] = NewThruModule(fmt.Sprintf("patch_%s_output_%d", identifier, i+1))
		subModules = append(subModules, outputModules[i])
	}

	p := &BasePatch{
		BaseModule:    NewBaseModule(0, 0, identifier),
		subModules:    subModules,
		inputModules:  inputModules,
		outputModules: outputModules,
		addressMap:    map[string]Module{},
	}

	return p
}

func (p *BasePatch) NumInputs() int {
	return len(p.inputModules)
}

func (p *BasePatch) NumOutputs() int {
	return len(p.outputModules)
}

func (p *BasePatch) SetBuffersFromPool(pool *BufferPool) {
	for _, module := range p.subModules {
		module.SetBuffersFromPool(pool)
	}
}

func (p *BasePatch) TotalInputBuffersNeeded() int {
	totalInputBuffersNeeded := 0

	for _, module := range p.subModules {
		totalInputBuffersNeeded += module.TotalInputBuffersNeeded()
	}

	return totalInputBuffersNeeded
}

func (p *BasePatch) TotalOutputBuffersNeeded() int {
	totalOutputBuffersNeeded := 0

	for _, module := range p.subModules {
		totalOutputBuffersNeeded += module.TotalOutputBuffersNeeded()
	}

	return totalOutputBuffersNeeded
}

func (p *BasePatch) AddModule(m Module) {
	p.subModules = append(p.subModules, m)

	id := m.Identifier()
	if id != "" {
		p.addressMap[id] = m
	}
}

func (p *BasePatch) Contains(m Module) bool {
	for _, sub := range p.subModules {
		if sub == m {
			return true
		}
	}

	return false
}

func (p *BasePatch) Lookup(address string) Module {
	components := strings.SplitN(address, ".", 2)
	modId := ""
	restAddress := ""

	if len(components) > 0 {
		modId = components[0]
	}

	if len(components) == 2 {
		restAddress = components[1]
	}

	m := p.addressMap[modId]
	if m == nil {
		return nil
	}

	if restAddress != "" {
		sub, ok := m.(Patch)
		if ok {
			return sub.Lookup(restAddress)
		} else {
			return nil
		}
	}

	return m
}

func (p *BasePatch) PrepareRun() {
	p.BaseModule.PrepareRun()

	for _, module := range p.subModules {
		module.PrepareRun()
	}
}

func (p *BasePatch) Run(config *Configuration) bool {
	if !p.BaseModule.Run(config) {
		return false
	}

	for _, output := range p.outputModules {
		if output.InputAtIndex(0).IsConnected() {
			output.Run(config)
		}
	}

	return true
}

func (p *BasePatch) AddInputConnection(inputIndex int, conn *Connection) {
	p.inputModules[inputIndex].AddInputConnection(0, conn)
}

func (p *BasePatch) AddOutputConnection(outputIndex int, conn *Connection) {
	p.outputModules[outputIndex].AddOutputConnection(0, conn)
}

func (p *BasePatch) InputModuleAtIndex(index int) Module {
	return p.inputModules[index]
}

func (p *BasePatch) OutputModuleAtIndex(index int) Module {
	return p.outputModules[index]
}

func (p *BasePatch) InputAtIndex(index int) *Socket {
	return p.inputModules[index].InputAtIndex(0)
}

func (p *BasePatch) OutputAtIndex(index int) *Socket {
	return p.outputModules[index].OutputAtIndex(0)
}

type Environment struct {
	*BasePatch
	pool   *BufferPool
	Config *Configuration
}

func NewEnvironment(numOutputs int, sampleRate float64, bufferSize int) *Environment {
	return &Environment{
		BasePatch: NewPatch(0, numOutputs, "environment"),
		Config:    &Configuration{SampleRate: sampleRate, BufferSize: bufferSize},
	}
}

func (e *Environment) PrepareBuffers() {
	e.pool = NewBufferPool(e.TotalInputBuffersNeeded(), e.TotalOutputBuffersNeeded(), e.Config.BufferSize)
	e.SetBuffersFromPool(e.pool)
}

func (e *Environment) Produce() {
	e.pool.ClearInputBuffers()
	e.PrepareRun()
	e.Run(e.Config)
}

func Connect(from Module, outIndex int, to Module, inIndex int) {
	p, ok := from.(Patch)
	if ok {
		if p.Contains(to) {
			from = p.InputModuleAtIndex(outIndex)
			outIndex = 0
		} else {
			from = p.OutputModuleAtIndex(outIndex)
			outIndex = 0
		}
	}

	p, ok = to.(Patch)
	if ok {
		if p.Contains(from) {
			to = p.OutputModuleAtIndex(inIndex)
			inIndex = 0
		} else {
			to = p.InputModuleAtIndex(inIndex)
			inIndex = 0
		}
	}

	from.AddOutputConnection(outIndex, &Connection{Module: to, Index: inIndex})
	to.AddInputConnection(inIndex, &Connection{Module: from, Index: outIndex})
}

type TestModule struct {
	*BaseModule
	Value Sample
}

func NewTestModule(value Sample, identifier string) *TestModule {
	return &TestModule{
		BaseModule: NewBaseModule(1, 1, identifier),
		Value:      value,
	}
}

func (t *TestModule) Run(config *Configuration) bool {
	if !t.BaseModule.Run(config) {
		return false
	}

	if t.inputs[0].IsConnected() {
		for i := 0; i < config.BufferSize; i++ {
			t.outputs[0].Buffer[i] = t.inputs[0].Buffer[i] + t.Value
		}
	} else {
		for i := 0; i < config.BufferSize; i++ {
			t.outputs[0].Buffer[i] = t.Value
		}
	}

	return true
}

func main() {
	env := NewEnvironment(1, 44100, 12)

	ip := NewPatch(1, 1, "ip_patch")
	it1 := NewTestModule(0.25, "it1")

	ip.AddModule(it1)
	Connect(ip, 0, it1, 0)
	Connect(it1, 0, ip, 0)

	t1 := NewTestModule(1.25, "t1")
	t11 := NewTestModule(0.123, "t11")
	t2 := NewTestModule(3.4, "t2")

	env.AddModule(t1)
	env.AddModule(t11)
	env.AddModule(ip)
	env.AddModule(t2)

	log.Printf("lookup %v", env.Lookup("ip_patch.it1"))

	Connect(t1, 0, ip, 0)
	Connect(t11, 0, ip, 0)
	Connect(ip, 0, t2, 0)
	Connect(t2, 0, env, 0)

	env.PrepareBuffers()

	for i := 0; i < 12; i++ {
		env.Produce()
		for _, sample := range env.OutputAtIndex(0).Buffer {
			log.Printf("%v", sample)
		}
	}
}
