package muse

import (
	"fmt"
	"strings"
)

type Patch interface {
	Module
	AddMessenger(msgr Messenger) Messenger
	AddModule(m Module) Module
	Contains(m Module) bool
	Lookup(address string) Receiver
	InputModuleAtIndex(index int) Module
	OutputModuleAtIndex(index int) Module
	PostMessages()
}

type BasePatch struct {
	*BaseModule
	subModules    []Module
	inputModules  []*ThruModule
	outputModules []*ThruModule
	messengers    []Messenger
	receivers     map[string]Receiver
	timestamp     int64
}

func NewPatch(numInputs int, numOutputs int, config *Configuration, identifier string) *BasePatch {
	subModules := []Module{}

	inputModules := make([]*ThruModule, numInputs)
	for i := 0; i < numInputs; i++ {
		inputModules[i] = NewThruModule(config, fmt.Sprintf("patch_%s_input_%d", identifier, i+1))
		subModules = append(subModules, inputModules[i])
	}

	outputModules := make([]*ThruModule, numOutputs)
	for i := 0; i < numOutputs; i++ {
		outputModules[i] = NewThruModule(config, fmt.Sprintf("patch_%s_output_%d", identifier, i+1))
		subModules = append(subModules, outputModules[i])
	}

	p := &BasePatch{
		BaseModule:    NewBaseModule(0, 0, config, identifier),
		subModules:    subModules,
		messengers:    []Messenger{},
		inputModules:  inputModules,
		outputModules: outputModules,
		receivers:     map[string]Receiver{},
	}

	return p
}

func (p *BasePatch) NumInputs() int {
	return len(p.inputModules)
}

func (p *BasePatch) NumOutputs() int {
	return len(p.outputModules)
}

func (p *BasePatch) AddMessenger(msgr Messenger) Messenger {
	p.messengers = append(p.messengers, msgr)

	id := msgr.Identifier()
	if id != "" {
		p.receivers[id] = msgr
	}

	return msgr
}

func (p *BasePatch) AddModule(m Module) Module {
	p.subModules = append(p.subModules, m)

	id := m.Identifier()
	if id != "" {
		p.receivers[id] = m
	}

	return m
}

func (p *BasePatch) Contains(m Module) bool {
	for _, sub := range p.subModules {
		if sub == m {
			return true
		}
	}

	return false
}

func (p *BasePatch) Lookup(address string) Receiver {
	components := strings.SplitN(address, ".", 2)
	identifier := ""
	restAddress := ""

	if len(components) > 0 {
		identifier = components[0]
	}

	if len(components) == 2 {
		restAddress = components[1]
	}

	m := p.receivers[identifier]
	if m == nil {
		return nil
	}

	if restAddress != "" {
		sub, ok := m.(Patch)
		if ok {
			return sub.Lookup(restAddress)
		} else {
			return nil
		}
	}

	return m
}

func (p *BasePatch) PrepareSynthesis() {
	p.BaseModule.PrepareSynthesis()

	for _, module := range p.subModules {
		module.PrepareSynthesis()
	}
}

func (p *BasePatch) Synthesize() bool {
	if !p.BaseModule.Synthesize() {
		return false
	}

	for _, output := range p.outputModules {
		output.Synthesize()
	}

	p.timestamp += int64(p.Config.BufferSize)

	return true
}

func (p *BasePatch) sendMessages(msgs []*Message) {
	for _, msg := range msgs {
		rcvr := p.Lookup(msg.Address)
		if rcvr != nil {
			p.sendMessages(rcvr.ReceiveMessage(msg.Content))
		}
	}
}

func (p *BasePatch) PostMessages() {
	for _, msgr := range p.messengers {
		p.sendMessages(msgr.Messages(p.timestamp, p.Config))
	}

	for _, sub := range p.subModules {
		subPatch, ok := sub.(Patch)
		if ok {
			subPatch.PostMessages()
		}
	}
}

func (p *BasePatch) AddInputConnection(inputIndex int, conn *Connection) {
	p.inputModules[inputIndex].AddInputConnection(0, conn)
}

func (p *BasePatch) AddOutputConnection(outputIndex int, conn *Connection) {
	p.outputModules[outputIndex].AddOutputConnection(0, conn)
}

func (p *BasePatch) InputModuleAtIndex(index int) Module {
	return p.inputModules[index]
}

func (p *BasePatch) OutputModuleAtIndex(index int) Module {
	return p.outputModules[index]
}

func (p *BasePatch) InputAtIndex(index int) *Socket {
	return p.inputModules[index].InputAtIndex(0)
}

func (p *BasePatch) OutputAtIndex(index int) *Socket {
	return p.outputModules[index].OutputAtIndex(0)
}
