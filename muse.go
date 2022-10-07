package muse

func Connect(from Module, outIndex int, to Module, inIndex int) {
	p, ok := from.(Patch)
	if ok {
		if p.Contains(to) {
			from = p.InputModuleAtIndex(outIndex)
			outIndex = 0
		}
	}

	p, ok = to.(Patch)
	if ok {
		if p.Contains(from) {
			to = p.OutputModuleAtIndex(inIndex)
			inIndex = 0
		} else {
			to = p.InputModuleAtIndex(inIndex)
			inIndex = 0
		}
	}

	from.AddOutputConnection(outIndex, &Connection{Module: to, Index: inIndex})
	to.AddInputConnection(inIndex, &Connection{Module: from, Index: outIndex})
}
