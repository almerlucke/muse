package muse

import (
	"fmt"
	"strings"
)

type Buffer []float64

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
	DidSynthesize() bool
	PrepareSynthesis()
	Synthesize(config *Configuration) bool
	ReceiveMessage(msg any)
}

type BaseModule struct {
	inputs        []*Socket
	outputs       []*Socket
	didSynthesize bool
	identifier    string
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

func (m *BaseModule) DidSynthesize() bool {
	return m.didSynthesize
}

func (m *BaseModule) PrepareSynthesis() {
	m.didSynthesize = false
}

func (m *BaseModule) Synthesize(config *Configuration) bool {
	if m.didSynthesize {
		return false
	}

	m.didSynthesize = true

	for _, input := range m.inputs {
		inputBuffer := input.Buffer

		for _, conn := range input.Connections {
			conn.Module.Synthesize(config)

			for bufIndex, sample := range conn.Module.OutputAtIndex(conn.Index).Buffer {
				inputBuffer[bufIndex] += sample
			}
		}
	}

	return true
}

func (m *BaseModule) ReceiveMessage(msg any) {
	// Do nothing
}

type ThruModule struct {
	*BaseModule
}

func NewThruModule(identifier string) *ThruModule {
	return &ThruModule{
		BaseModule: NewBaseModule(1, 1, identifier),
	}
}

func (t *ThruModule) Synthesize(config *Configuration) bool {
	if !t.BaseModule.Synthesize(config) {
		return false
	}

	for i := 0; i < config.BufferSize; i++ {
		t.outputs[0].Buffer[i] = t.inputs[0].Buffer[i]
	}

	return true
}

type Message struct {
	Address string
	Content any
}

type Messenger interface {
	Post(timestamp int64, config *Configuration) []*Message
}

type Patch interface {
	Module
	AddMessenger(msgr Messenger)
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
	messengers    []Messenger
	identifierMap map[string]Module
	timestamp     int64
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
		messengers:    []Messenger{},
		inputModules:  inputModules,
		outputModules: outputModules,
		identifierMap: map[string]Module{},
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

func (p *BasePatch) AddMessenger(msgr Messenger) {
	p.messengers = append(p.messengers, msgr)
}

func (p *BasePatch) AddModule(m Module) {
	p.subModules = append(p.subModules, m)

	id := m.Identifier()
	if id != "" {
		p.identifierMap[id] = m
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
	identifier := ""
	restAddress := ""

	if len(components) > 0 {
		identifier = components[0]
	}

	if len(components) == 2 {
		restAddress = components[1]
	}

	m := p.identifierMap[identifier]
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

func (p *BasePatch) PrepareSynthesis() {
	p.BaseModule.PrepareSynthesis()

	for _, module := range p.subModules {
		module.PrepareSynthesis()
	}
}

func (p *BasePatch) Synthesize(config *Configuration) bool {
	if !p.BaseModule.Synthesize(config) {
		return false
	}

	for _, msgr := range p.messengers {
		msgs := msgr.Post(p.timestamp, config)
		for _, msg := range msgs {
			module := p.Lookup(msg.Address)
			if module != nil {
				module.ReceiveMessage(msg.Content)
			}
		}
	}

	for _, output := range p.outputModules {
		output.Synthesize(config)
	}

	p.timestamp += int64(config.BufferSize)

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
	e.PrepareSynthesis()
	e.Synthesize(e.Config)
}

func (e *Environment) SynthesizeToFile(filePath string, numSeconds float64) error {
	e.PrepareBuffers()

	numChannels := e.NumOutputs()

	swr, err := OpenSoundWriter(filePath, int32(numChannels), int32(e.Config.SampleRate), true)
	if err != nil {
		return err
	}

	interleaveBuffer := make([]float64, e.NumOutputs()*e.Config.BufferSize)

	framesToProduce := int64(e.Config.SampleRate * numSeconds)

	for framesToProduce > 0 {
		e.Produce()

		interleaveIndex := 0

		numFrames := e.Config.BufferSize

		if framesToProduce <= int64(e.Config.BufferSize) {
			numFrames = int(framesToProduce)
		}

		for i := 0; i < numFrames; i++ {
			for c := 0; c < numChannels; c++ {
				interleaveBuffer[interleaveIndex] = e.OutputAtIndex(c).Buffer[i]
				interleaveIndex++
			}
		}

		swr.WriteSamples(interleaveBuffer[:numFrames*numChannels])

		framesToProduce -= int64(numFrames)
	}

	return swr.Close()
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
