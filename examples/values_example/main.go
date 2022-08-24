package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/almerlucke/muse/values"
)

type Test struct {
	V float64 `json:"v"`
}

func (t *Test) GetState() map[string]any {
	return map[string]any{"v": t.V}
}

func (t *Test) SetState(s map[string]any) {
	t.V = s["v"].(float64)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	c1 := values.NewConst(&Test{V: 1.2})

	c2 := values.NewConst(&Test{V: 1.0})

	b1, err := json.Marshal(c1.GetState())
	if err != nil {
		log.Printf("marshal err %v", err)
	}

	log.Printf("b1 done")

	var jsonState map[string]any

	err = json.Unmarshal(b1, &jsonState)
	if err != nil {
		log.Printf("unmarshal err %v", err)
	}

	log.Printf("unmarshal done %v", jsonState)

	c2.SetState(jsonState)

	log.Printf("state %v", c2.GetState())

	// p := values.MapPrototype{
	// 	"duration":  values.NewSequence([]any{250.0, 500.0, 125.0, 250.0, 250.0, 750.0, 500.0, 375.0, 250.0, 250.0}, true),
	// 	"amplitude": values.NewSequence([]any{1.0, 0.6, 1.0, 0.5, 0.5, 1.0, 0.3, 1.0, 0.7, 1.0}, true),
	// 	"message": values.MapPrototype{
	// 		"osc": values.MapPrototype{
	// 			"frequency": values.NewSequence([]any{400.0, 500.0, 600.0, 100.0, 50.0, 50.0, 100.0, 250.0, 750.0}, true),
	// 			"phase":     values.NewConst[any](0.0),
	// 		},
	// 	},
	// }

	// m := p.Map()
	// log.Printf("d: %v", m.F("duration"))
	// log.Printf("f: %v", m.M("message").M("osc").F("frequency"))
	// log.Printf("p: %v", m.M("message").M("osc").F("phase"))
	// m = p.Map()
	// log.Printf("d: %v", m.F("duration"))
	// log.Printf("f: %v", m.M("message").M("osc").F("frequency"))
	// log.Printf("p: %v", m.M("message").M("osc").F("phase"))
	// m = p.Map()
	// log.Printf("d: %v", m.F("duration"))
	// log.Printf("f: %v", m.M("message").M("osc").F("frequency"))
	// log.Printf("p: %v", m.M("message").M("osc").F("phase"))
}
