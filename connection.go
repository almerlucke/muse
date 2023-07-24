package muse

type Connection struct {
	Module Module
	Index  int
}

// IConn for quick connecting multiple module/control outputs to inputs of module/control
type IConn struct {
	Object   any
	OutIndex int
	InIndex  int
}

// IConns parse args to []*ICon, args need to be in the following order:
// - required module whose output is routed to input index
// - optional outIndex of module (default 0)
// - if outIndex is given follows an optional inIndex (otherwise index of count of IConn's parsed so far)
// - repeat...
func IConns(args ...any) []*IConn {
	n := len(args)

	argIndex := 0
	inIndexCnt := 0
	iConns := []*IConn{}

	var inIndex int
	var outIndex int

	for argIndex < len(args) {
		obj := args[argIndex]
		argIndex++

		outIndex = 0
		inIndex = inIndexCnt

		if argIndex < n {
			if givenOutIndex, ok := args[argIndex].(int); ok {
				outIndex = givenOutIndex
				argIndex++
			}
		}

		if argIndex < n {
			if givenInIndex, ok := args[argIndex].(int); ok {
				inIndex = givenInIndex
				argIndex++
			}
		}

		iConns = append(iConns, &IConn{Object: obj, OutIndex: outIndex, InIndex: inIndex})

		inIndexCnt++
	}

	return iConns
}

// IConn parse input and output to []*ICon
// func IConns(rawIconns ...any) []*IConn {
// 	index := 0
// 	icons := make([]*IConn, len(rawIconns)/3)

// 	for index < len(rawIconns) {
// 		module := rawIconns[index].(Module)
// 		outIndex := rawIconns[index+1].(int)
// 		inIndex := rawIconns[index+2].(int)

// 		icons[index/3] = &IConn{
// 			Module:   module,
// 			OutIndex: outIndex,
// 			InIndex:  inIndex,
// 		}

// 		index += 3
// 	}

// 	return icons
// }
