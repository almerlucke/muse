package muse

import (
	"fmt"
	"strings"

	"github.com/almerlucke/muse/io"
)

type Buffer []float64

func (b Buffer) Clear() {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

type Connection struct {
	Module Module
	Index  int
}

type Socket struct {
	Buffer      Buffer
	Connections []*Connection
}

func NewSocket(bufferSize int) *Socket {
	return &Socket{
		Buffer:      make(Buffer, bufferSize),
		Connections: []*Connection{},
	}
}

func (s *Socket) AddConnection(c *Connection) {
	s.Connections = append(s.Connections, c)
}

func (s *Socket) IsConnected() bool {
	return len(s.Connections) > 0
}

type Configuration struct {
	SampleRate float64
	BufferSize int
}

type Module interface {
	SetIdentifier(identifier string)
	Identifier() string
	NumInputs() int
	NumOutputs() int
	Configuration() *Configuration
	InputAtIndex(index int) *Socket
	OutputAtIndex(index int) *Socket
	AddInputConnection(inputIndex int, conn *Connection)
	AddOutputConnection(outputIndex int, conn *Connection)
	DidSynthesize() bool
	PrepareSynthesis()
	Synthesize() bool
	ReceiveMessage(msg any)
}

type BaseModule struct {
	Inputs        []*Socket
	Outputs       []*Socket
	Config        *Configuration
	didSynthesize bool
	identifier    string
}

func NewBaseModule(numInputs int, numOutputs int, config *Configuration, identifier string) *BaseModule {
	inputs := make([]*Socket, numInputs)

	for i := 0; i < numInputs; i++ {
		inputs[i] = NewSocket(config.BufferSize)
	}

	outputs := make([]*Socket, numOutputs)

	for i := 0; i < numOutputs; i++ {
		outputs[i] = NewSocket(config.BufferSize)
	}

	return &BaseModule{
		Inputs:     inputs,
		Outputs:    outputs,
		Config:     config,
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
	return len(m.Inputs)
}

func (m *BaseModule) NumOutputs() int {
	return len(m.Outputs)
}

func (m *BaseModule) InputAtIndex(index int) *Socket {
	return m.Inputs[index]
}

func (m *BaseModule) OutputAtIndex(index int) *Socket {
	return m.Outputs[index]
}

func (m *BaseModule) Configuration() *Configuration {
	return m.Config
}

func (m *BaseModule) AddInputConnection(inputIndex int, conn *Connection) {
	m.Inputs[inputIndex].AddConnection(conn)
}

func (m *BaseModule) AddOutputConnection(outputIndex int, conn *Connection) {
	m.Outputs[outputIndex].AddConnection(conn)
}

func (m *BaseModule) DidSynthesize() bool {
	return m.didSynthesize
}

func (m *BaseModule) PrepareSynthesis() {
	m.didSynthesize = false

	// Clear input buffers first as we accumulate multiple inputs in one buffer
	for _, input := range m.Inputs {
		input.Buffer.Clear()
	}
}

func (m *BaseModule) Synthesize() bool {
	if m.didSynthesize {
		return false
	}

	m.didSynthesize = true

	for _, input := range m.Inputs {
		// Accumulate connection outputs in single input
		inputBuffer := input.Buffer

		for _, conn := range input.Connections {
			conn.Module.Synthesize()

			for bufIndex, sample := range conn.Module.OutputAtIndex(conn.Index).Buffer {
				inputBuffer[bufIndex] += sample
			}
		}
	}

	return true
}

func (m *BaseModule) ReceiveMessage(msg any) {
	// STUB
}

type ThruModule struct {
	*BaseModule
}

func NewThruModule(config *Configuration, identifier string) *ThruModule {
	return &ThruModule{
		BaseModule: NewBaseModule(1, 1, config, identifier),
	}
}

func (t *ThruModule) Synthesize() bool {
	if !t.BaseModule.Synthesize() {
		return false
	}

	for i := 0; i < t.Config.BufferSize; i++ {
		t.Outputs[0].Buffer[i] = t.Inputs[0].Buffer[i]
	}

	return true
}

type Message struct {
	Address string `json:"address"`
	Content any    `json:"content"`
}

type Messenger interface {
	Messages(timestamp int64, config *Configuration) []*Message
}

type Patch interface {
	Module
	AddMessenger(msgr Messenger)
	AddModule(m Module)
	Contains(m Module) bool
	Lookup(address string) Module
	InputModuleAtIndex(index int) Module
	OutputModuleAtIndex(index int) Module
	PostMessages()
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

func NewPatch(numInputs int, numOutputs int, config *Configuration, identifier string) *BasePatch {
	subModules := []Module{}

	inputModules := make([]*ThruModule, numInputs)
	for i := 0; i < numInputs; i++ {
		inputModules[i] = NewThruModule(config, fmt.Sprintf("patch_%s_input_%d", identifier, i+1))
		subModules = append(subModules, inputModules[i])
	}

	outputModules := make([]*ThruModule, numOutputs)
	for i := 0; i < numOutputs; i++ {
		outputModules[i] = NewThruModule(config, fmt.Sprintf("patch_%s_output_%d", identifier, i+1))
		subModules = append(subModules, outputModules[i])
	}

	p := &BasePatch{
		BaseModule:    NewBaseModule(0, 0, config, identifier),
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

func (p *BasePatch) Synthesize() bool {
	if !p.BaseModule.Synthesize() {
		return false
	}

	for _, output := range p.outputModules {
		output.Synthesize()
	}

	p.timestamp += int64(p.Config.BufferSize)

	return true
}

func (p *BasePatch) PostMessages() {
	for _, msgr := range p.messengers {
		msgs := msgr.Messages(p.timestamp, p.Config)
		for _, msg := range msgs {
			module := p.Lookup(msg.Address)
			if module != nil {
				module.ReceiveMessage(msg.Content)
			}
		}
	}

	for _, sub := range p.subModules {
		subPatch, ok := sub.(Patch)
		if ok {
			subPatch.PostMessages()
		}
	}
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
	Config *Configuration
}

func NewEnvironment(numOutputs int, sampleRate float64, bufferSize int) *Environment {
	config := &Configuration{SampleRate: sampleRate, BufferSize: bufferSize}

	return &Environment{
		BasePatch: NewPatch(0, numOutputs, config, "environment"),
		Config:    config,
	}
}

func (e *Environment) Synthesize() bool {
	e.PostMessages()
	e.PrepareSynthesis()
	return e.BasePatch.Synthesize()
}

func (e *Environment) SynthesizeToFile(filePath string, numSeconds float64) error {
	numChannels := e.NumOutputs()

	swr, err := io.OpenSoundWriter(filePath, int32(numChannels), int32(e.Config.SampleRate), true)
	if err != nil {
		return err
	}

	interleaveBuffer := make([]float64, e.NumOutputs()*e.Config.BufferSize)

	framesToProduce := int64(e.Config.SampleRate * numSeconds)

	for framesToProduce > 0 {
		e.Synthesize()

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
