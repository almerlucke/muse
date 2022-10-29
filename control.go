package muse

type ControlReceiver interface {
	ReceiveControlValue(any, int)
}

type ControlSender interface {
	SendControlValue(any, int)
}

type ControlConnector interface {
	AddControlInputConnection(int, Control, int)
	AddControlOutputConnection(int, Control, int)
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
	Control Control
	Index   int
}

// Control acts at control rate (once every audio frame) instead of audio rate
type BaseControl struct {
	identifier     string
	inConnections  map[int][]*ControlConnection
	outConnections map[int][]*ControlConnection
}

func NewBaseControl(id string) *BaseControl {
	return &BaseControl{
		inConnections:  map[int][]*ControlConnection{},
		outConnections: map[int][]*ControlConnection{},
		identifier:     id,
	}
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
	connections := c.outConnections[index]
	if connections != nil {
		for _, connection := range connections {
			connection.Control.ReceiveControlValue(value, connection.Index)
		}
	}
}

func (c *BaseControl) AddControlInputConnection(inputIndex int, sender Control, outputIndex int) {
	connections := c.inConnections[inputIndex]
	if connections == nil {
		connections = []*ControlConnection{}
	}

	connections = append(connections, &ControlConnection{Control: sender, Index: outputIndex})

	c.inConnections[inputIndex] = connections
}

func (c *BaseControl) AddControlOutputConnection(outputIndex int, receiver Control, inputIndex int) {
	connections := c.outConnections[outputIndex]
	if connections == nil {
		connections = []*ControlConnection{}
	}

	connections = append(connections, &ControlConnection{Control: receiver, Index: inputIndex})

	c.outConnections[outputIndex] = connections
}

func ConnectControl(sender Control, outIndex int, receiver Control, inIndex int) {
	sender.AddControlOutputConnection(outIndex, receiver, inIndex)
	receiver.AddControlInputConnection(inIndex, sender, outIndex)
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
