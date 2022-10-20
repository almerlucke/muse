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
	ControlSupporter
	Identifiable
	MessageReceiver
	Stater
	Messages(timestamp int64, config *Configuration) []*Message
}

type BaseMessenger struct {
	*BaseControlSupport
}

func NewBaseMessenger(identifier string) *BaseMessenger {
	return &BaseMessenger{
		BaseControlSupport: NewBaseControlSupport(identifier),
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

func (m *BaseMessenger) SetState(state map[string]any) {
	// STUB
}

func (m *BaseMessenger) GetState() map[string]any {
	// STUB
	return map[string]any{}
}
