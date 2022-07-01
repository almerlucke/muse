package muse

type Socket struct {
	Buffer      Buffer
	Connections []*Connection
}

func NewSocket(bufferSize int) *Socket {
	return &Socket{
		Buffer:      make(Buffer, bufferSize),
		Connections: []*Connection{},
	}
}

func (s *Socket) AddConnection(c *Connection) {
	s.Connections = append(s.Connections, c)
}

func (s *Socket) IsConnected() bool {
	return len(s.Connections) > 0
}
