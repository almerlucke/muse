package muse

type Module interface {
	ControlSupporter
	MessageReceiver
	Identifiable
	Stater
	NumInputs() int
	NumOutputs() int
	Configuration() *Configuration
	InputAtIndex(index int) *Socket
	OutputAtIndex(index int) *Socket
	AddInputConnection(inputIndex int, conn *Connection)
	AddOutputConnection(outputIndex int, conn *Connection)
	DidSynthesize() bool
	MustSynthesize() bool
	PrepareSynthesis()
	Synthesize() bool
}

type BaseModule struct {
	*BaseControlSupport
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
		BaseControlSupport: NewBaseControlSupport(identifier),
		Inputs:             inputs,
		Outputs:            outputs,
		Config:             config,
	}
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

func (m *BaseModule) MustSynthesize() bool {
	return false
}

func (m *BaseModule) ReceiveMessage(msg any) []*Message {
	// STUB
	return nil
}

func (m *BaseModule) SetState(state map[string]any) {
	// STUB
}

func (m *BaseModule) GetState() map[string]any {
	// STUB
	return map[string]any{}
}
