package muse

type Connection struct {
	Module Module
	Index  int
}

// IConn for quick connecting multiple module outputs to inputs of module
type IConn struct {
	InIndex  int
	Module   Module
	OutIndex int
}

// IConn parse input and output to []*ICon
func IConns(rawIconns ...any) []*IConn {
	index := 0
	icons := make([]*IConn, len(rawIconns)/3)

	for index < len(rawIconns) {
		inIndex := rawIconns[index].(int)
		module := rawIconns[index+1].(Module)
		outIndex := rawIconns[index+2].(int)

		icons[index/3] = &IConn{
			InIndex:  inIndex,
			Module:   module,
			OutIndex: outIndex,
		}

		index += 3
	}

	return icons
}
