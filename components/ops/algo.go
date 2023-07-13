package ops

type Algo struct {
	OpTypes     map[string]string   `json:"opTypes"`
	Objects     []map[string]string `json:"objects"`
	Connections [][]string          `json:"connections"`
	Root        string              `json:"root"`
}

var DX7Algo1 = &Algo{
	OpTypes: map[string]string{
		"op1": "modulator",
		"op2": "carrier",
		"op3": "modulator",
		"op4": "modulator",
		"op5": "modulator",
		"op6": "carrier",
	},
	Objects: []map[string]string{{
		"id": "mix1", "type": "mix",
	}},
	Connections: [][]string{
		{"op1", "op2"},
		{"op3", "op3"},
		{"op3", "op4"},
		{"op4", "op5"},
		{"op5", "op6"},
		{"op2", "mix1"},
		{"op6", "mix1"},
	},
	Root: "mix1",
}
