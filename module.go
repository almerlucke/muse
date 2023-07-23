package muse

type Module interface {
	Control
	Stater
	Named(string) Module
	Add(Patch) Module
	Configuration() *Configuration
	NumInputs() int
	NumOutputs() int
	InputAtIndex(int) *Socket
	OutputAtIndex(int) *Socket
	AddInputConnection(int, *Connection)
	AddOutputConnection(int, *Connection)
	RemoveInputConnection(int, Module, int)
	RemoveOutputConnection(int, Module, int)
	DidSynthesize() bool
	MustSynthesize() bool
	PrepareSynthesis()
	Synthesize() bool
	In(...any) Module
	IConns([]*IConn) Module
	Connect(int, Module, int)
	Disconnect()
}

type BaseModule struct {
	*BaseControl
	Inputs        []*Socket
	Outputs       []*Socket
	Config        *Configuration
	didSynthesize bool
}

func NewBaseModule(numInputs int, numOutputs int) *BaseModule {
	config := CurrentConfiguration()

	inputs := make([]*Socket, numInputs)

	for i := 0; i < numInputs; i++ {
		inputs[i] = NewSocket(config.BufferSize)
	}

	outputs := make([]*Socket, numOutputs)

	for i := 0; i < numOutputs; i++ {
		outputs[i] = NewSocket(config.BufferSize)
	}

	return &BaseModule{
		BaseControl: NewBaseControl(),
		Inputs:      inputs,
		Outputs:     outputs,
		Config:      config,
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

func (m *BaseModule) RemoveInputConnection(inputIndex int, sender Module, outputIndex int) {
	removeIndex := -1
	conns := m.Inputs[inputIndex].Connections
	for index, conn := range conns {
		if conn.Index == outputIndex && conn.Module == sender {
			removeIndex = index
			break
		}
	}
	if removeIndex > -1 {
		m.Inputs[inputIndex].Connections = append(conns[:removeIndex], conns[removeIndex+1:]...)
	}
}

func (m *BaseModule) RemoveOutputConnection(outputIndex int, receiver Module, inputIndex int) {
	removeIndex := -1
	conns := m.Outputs[outputIndex].Connections
	for index, conn := range conns {
		if conn.Index == inputIndex && conn.Module == receiver {
			removeIndex = index
			break
		}
	}
	if removeIndex > -1 {
		m.Outputs[outputIndex].Connections = append(conns[:removeIndex], conns[removeIndex+1:]...)
	}
}

func (m *BaseModule) Named(name string) Module {
	self := m.Self().(Module)
	self.SetIdentifier(name)
	return self
}

func (m *BaseModule) Add(p Patch) Module {
	return p.AddModule(m.Self().(Module))
}

func (m *BaseModule) IConns(iConns []*IConn) Module {
	self := m.Self().(Module)
	for _, iConn := range iConns {
		iConn.Object.(Module).Connect(iConn.OutIndex, self, iConn.InIndex)
	}
	return self
}

func (m *BaseModule) In(rawIconns ...any) Module {
	return m.Self().(Module).IConns(IConns(rawIconns...))
}

func (m *BaseModule) Connect(outIndex int, to Module, inIndex int) {
	from := m.Self().(Module)

	p, ok := from.(Patch)
	if ok {
		if p.Contains(to) {
			from = p.InputModuleAtIndex(outIndex)
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

func (m *BaseModule) Disconnect() {
	self := m.Self().(Module)

	for inIndex, socket := range m.Inputs {
		for _, conn := range socket.Connections {
			conn.Module.RemoveOutputConnection(conn.Index, self, inIndex)
		}
	}

	for outIndex, socket := range m.Outputs {
		for _, conn := range socket.Connections {
			conn.Module.RemoveInputConnection(conn.Index, self, outIndex)
		}
	}
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
