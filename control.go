package muse

type ControlReceiver interface {
	ReceiveControlValue(any, int)
}

type ControlSender interface {
	SendControlValue(any, int)
}

type Control interface {
	Selfie
	Identifiable
	MessageReceiver
	ControlReceiver
	ControlSender
	AddControlInputConnection(int, Control, int)
	AddControlOutputConnection(int, Control, int)
	RemoveControlInputConnection(int, Control, int)
	RemoveControlOutputConnection(int, Control, int)
	CtrlConnect(int, Control, int)
	CtrlDisconnect()
	Tick(int64, *Configuration)
}

type ControlConnection struct {
	Control Control
	Index   int
}

// Control acts at control rate (once every audio frame) instead of audio rate
type BaseControl struct {
	identifier     string
	self           any
	inConnections  map[int][]*ControlConnection
	outConnections map[int][]*ControlConnection
}

func NewBaseControl(id string) *BaseControl {
	bc := &BaseControl{
		inConnections:  map[int][]*ControlConnection{},
		outConnections: map[int][]*ControlConnection{},
		identifier:     id,
	}

	bc.self = bc

	return bc
}

func (c *BaseControl) Identifier() string {
	return c.identifier
}

func (c *BaseControl) SetIdentifier(id string) {
	c.identifier = id
}

func (c *BaseControl) Self() any {
	return c.self
}

func (c *BaseControl) SetSelf(self any) {
	c.self = self
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

func (c *BaseControl) RemoveControlInputConnection(inputIndex int, sender Control, outputIndex int) {
	conns := c.inConnections[inputIndex]

	removeIndex := -1
	for index, conn := range conns {
		if conn.Control == sender && conn.Index == outputIndex {
			removeIndex = index
			break
		}
	}

	if removeIndex > -1 {
		c.inConnections[inputIndex] = append(conns[:removeIndex], conns[removeIndex+1:]...)
	}
}

func (c *BaseControl) RemoveControlOutputConnection(outputIndex int, receiver Control, inputIndex int) {
	conns := c.outConnections[outputIndex]

	removeIndex := -1
	for index, conn := range conns {
		if conn.Control == receiver && conn.Index == inputIndex {
			removeIndex = index
			break
		}
	}

	if removeIndex > -1 {
		c.outConnections[outputIndex] = append(conns[:removeIndex], conns[removeIndex+1:]...)
	}
}

func (c *BaseControl) CtrlConnect(outIndex int, receiver Control, inIndex int) {
	self := c.Self().(Control)
	self.AddControlOutputConnection(outIndex, receiver, inIndex)
	receiver.AddControlInputConnection(inIndex, self, outIndex)
}

func (c *BaseControl) CtrlDisconnect() {
	self := c.Self().(Control)

	for inIndex, inConns := range c.inConnections {
		for _, inConn := range inConns {
			inConn.Control.RemoveControlOutputConnection(inConn.Index, self, inIndex)
		}
	}

	for outIndex, outConns := range c.outConnections {
		for _, outConn := range outConns {
			outConn.Control.RemoveControlInputConnection(outConn.Index, self, outIndex)
		}
	}
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
