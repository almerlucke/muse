package muse

import "github.com/almerlucke/muse/buffer"

type Socket struct {
	Buffer      buffer.Buffer
	Connections []*Connection
}

func NewSocket(bufferSize int) *Socket {
	return &Socket{
		Buffer:      make(buffer.Buffer, bufferSize),
		Connections: []*Connection{},
	}
}

func (s *Socket) AddConnection(c *Connection) {
	s.Connections = append(s.Connections, c)
}

func (s *Socket) IsConnected() bool {
	return len(s.Connections) > 0
}
