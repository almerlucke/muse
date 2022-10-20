package muse

type ControlReceiver interface {
	ReceiveControlValue(any, int)
}

type ControlSender interface {
	SendControlValue(any, int)
}

type ControlConnector interface {
	ConnectControlOutput(int, ControlReceiver, int)
}

type ControlTicker interface {
	Identifiable
	MessageReceiver
	Tick(int64, *Configuration)
}

type ControlSupporter interface {
	ControlReceiver
	ControlSender
	ControlConnector
	ControlTicker
}

type ControlConnection struct {
	Receiver ControlReceiver
	Index    int
}

// Control acts at control rate (once every audio frame) instead of audio rate
type BaseControlSupport struct {
	identifier  string
	connections map[int][]*ControlConnection
}

func NewBaseControlSupport(id string) *BaseControlSupport {
	return &BaseControlSupport{connections: map[int][]*ControlConnection{}, identifier: id}
}

func (cs *BaseControlSupport) Identifier() string {
	return cs.identifier
}

func (cs *BaseControlSupport) SetIdentifier(id string) {
	cs.identifier = id
}

func (cs *BaseControlSupport) Tick(timestamp int64, config *Configuration) {
	// STUB: do anything with Tick in embedding struct
}

func (cs *BaseControlSupport) ReceiveControlValue(value any, index int) {
	// STUB: do anything with the value received in embedding struct
}

func (cs *BaseControlSupport) ReceiveMessage(msg any) []*Message {
	// STUB: do anything with the message received in embedding struct
	return nil
}

func (cs *BaseControlSupport) SendControlValue(value any, index int) {
	connections := cs.connections[index]
	if connections != nil {
		for _, connection := range connections {
			connection.Receiver.ReceiveControlValue(value, connection.Index)
		}
	}
}

func (cs *BaseControlSupport) ConnectControlOutput(outputIndex int, receiver ControlReceiver, inputIndex int) {
	connections := cs.connections[outputIndex]
	if connections == nil {
		connections = []*ControlConnection{}
	}

	connections = append(connections, &ControlConnection{Receiver: receiver, Index: inputIndex})

	cs.connections[outputIndex] = connections
}

type ControlThru struct {
	*BaseControlSupport
}

func NewControlThru() *ControlThru {
	return &ControlThru{BaseControlSupport: NewBaseControlSupport("")}
}

func (ct *ControlThru) ReceiveControlValue(value any, index int) {
	ct.SendControlValue(value, index)
}
