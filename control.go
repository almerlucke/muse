package muse

type ControlReceiver interface {
	ReceiveControlValue(any, int)
}

type ControlSender interface {
	SendControlValue(any, int)
}

type ControlConnector interface {
	ConnectToControl(int, ControlReceiver, int)
}

type ControlTicker interface {
	Identifiable
	MessageReceiver
	Tick(int64, *Configuration)
}

type Control interface {
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
type BaseControl struct {
	identifier  string
	connections map[int][]*ControlConnection
}

func NewBaseControl(id string) *BaseControl {
	return &BaseControl{connections: map[int][]*ControlConnection{}, identifier: id}
}

func (c *BaseControl) Identifier() string {
	return c.identifier
}

func (c *BaseControl) SetIdentifier(id string) {
	c.identifier = id
}

func (c *BaseControl) Tick(timestamp int64, config *Configuration) {
	// STUB: do anything with Tick in embedding struct
}

func (c *BaseControl) ReceiveControlValue(value any, index int) {
	// STUB: do anything with the value received in embedding struct
}

func (c *BaseControl) ReceiveMessage(msg any) []*Message {
	// STUB: do anything with the message received in embedding struct
	return nil
}

func (c *BaseControl) SendControlValue(value any, index int) {
	connections := c.connections[index]
	if connections != nil {
		for _, connection := range connections {
			connection.Receiver.ReceiveControlValue(value, connection.Index)
		}
	}
}

func (c *BaseControl) ConnectToControl(outputIndex int, receiver ControlReceiver, inputIndex int) {
	connections := c.connections[outputIndex]
	if connections == nil {
		connections = []*ControlConnection{}
	}

	connections = append(connections, &ControlConnection{Receiver: receiver, Index: inputIndex})

	c.connections[outputIndex] = connections
}

type ControlThru struct {
	*BaseControl
}

func NewControlThru() *ControlThru {
	return &ControlThru{BaseControl: NewBaseControl("")}
}

func (ct *ControlThru) ReceiveControlValue(value any, index int) {
	ct.SendControlValue(value, index)
}
