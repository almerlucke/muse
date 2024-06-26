package muse

type Message struct {
	Address string `json:"address"`
	Content any    `json:"content"`
}

func NewMessage(address string, content any) *Message {
	return &Message{
		Address: address,
		Content: content,
	}
}

type MessageReceiver interface {
	ReceiveMessage(msg any) []*Message
}

type Messenger interface {
	Control
	MsgrNamed(string) Messenger
	MsgrAddTo(Patch) Messenger
	Messages(timestamp int64, config *Configuration) []*Message
}

type BaseMessenger struct {
	*BaseControl
}

func NewBaseMessenger() *BaseMessenger {
	return &BaseMessenger{
		BaseControl: NewBaseControl(),
	}
}

func (m *BaseMessenger) ReceiveMessage(msg any) []*Message {
	// STUB
	return nil
}

func (m *BaseMessenger) Messages(timestamp int64, config *Configuration) []*Message {
	// STUB
	return nil
}

func (m *BaseMessenger) MsgrNamed(name string) Messenger {
	self := m.Self().(Messenger)
	self.SetIdentifier(name)
	return self
}

func (m *BaseMessenger) MsgrAddTo(p Patch) Messenger {
	return p.AddMessenger(m.Self().(Messenger))
}
