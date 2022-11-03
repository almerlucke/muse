package muse

type bang struct{ Bang string }

var Bang = &bang{Bang: "bang"}

func IsBang(msg any) bool {
	if msg == Bang {
		return true
	}

	if m, ok := msg.(map[string]any); ok {
		if rb, ok := m["bang"]; ok {
			if b, ok := rb.(bool); ok && b {
				return true
			}
		}
	}

	return false
}
