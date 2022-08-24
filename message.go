package muse

import (
	"encoding/json"
	"os"
)

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

func ReadMessages(file string) ([]*Message, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var messages []*Message

	err = json.Unmarshal(data, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

type Receiver interface {
	ReceiveMessage(msg any) []*Message
}

type Messenger interface {
	Identifiable
	Receiver
	Stater
	Messages(timestamp int64, config *Configuration) []*Message
}

type BaseMessenger struct {
	identifier string
}

func NewBaseMessenger(identifier string) *BaseMessenger {
	return &BaseMessenger{
		identifier: identifier,
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

func (m *BaseMessenger) Identifier() string {
	return m.identifier
}

func (m *BaseMessenger) SetIdentifier(identifier string) {
	m.identifier = identifier
}

func (m *BaseMessenger) SetState(state map[string]any) {
	// STUB
}

func (m *BaseMessenger) GetState() map[string]any {
	// STUB
	return map[string]any{}
}
