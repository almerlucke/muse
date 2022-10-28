package muse

import (
	"fmt"
	"strings"
)

type Patch interface {
	Module
	AddMessenger(Messenger) Messenger
	RemoveMessenger(Messenger)
	RemoveMessengerByID(string)
	AddMessageReceiver(MessageReceiver, string)
	RemoveMessageReceiverByID(string)
	AddModule(Module) Module
	AddControl(Control) Control
	Contains(Module) bool
	Lookup(string) MessageReceiver
	InputModuleAtIndex(index int) Module
	OutputModuleAtIndex(index int) Module
	SendMessage(*Message)
	SendMessages([]*Message)
	InternalInputControl() Control
	InternalOutputControl() Control
}

type BasePatch struct {
	*BaseModule
	internalInputControl  *ControlThru
	internalOutputControl *ControlThru
	subModules            []Module
	inputModules          []*ThruModule
	outputModules         []*ThruModule
	messengers            []Messenger
	controls              []Control
	receivers             map[string]MessageReceiver
	timestamp             int64
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
		BaseModule:            NewBaseModule(0, 0, config, identifier),
		internalInputControl:  NewControlThru(),
		internalOutputControl: NewControlThru(),
		subModules:            subModules,
		messengers:            []Messenger{},
		controls:              []Control{},
		inputModules:          inputModules,
		outputModules:         outputModules,
		receivers:             map[string]MessageReceiver{},
	}

	return p
}

func (p *BasePatch) NumInputs() int {
	return len(p.inputModules)
}

func (p *BasePatch) NumOutputs() int {
	return len(p.outputModules)
}

func (p *BasePatch) AddMessageReceiver(rcvr MessageReceiver, identifier string) {
	if identifier != "" {
		p.receivers[identifier] = rcvr
	}
}

func (p *BasePatch) RemoveMessageReceiverByID(id string) {
	if _, ok := p.receivers[id]; ok {
		delete(p.receivers, id)
	}
}

func (p *BasePatch) AddMessenger(msgr Messenger) Messenger {
	p.messengers = append(p.messengers, msgr)

	p.AddMessageReceiver(msgr, msgr.Identifier())

	return msgr
}

func (p *BasePatch) RemoveMessenger(msgr Messenger) {
	removeIndex := -1
	for index, otherMsgr := range p.messengers {
		if otherMsgr == msgr {
			removeIndex = index
			break
		}
	}

	if removeIndex > -1 {
		msgr := p.messengers[removeIndex]
		p.messengers = append(p.messengers[:removeIndex], p.messengers[removeIndex+1:]...)
		if receiver, ok := p.receivers[msgr.Identifier()]; ok {
			if receiver == msgr {
				delete(p.receivers, msgr.Identifier())
			}
		}
	}
}

func (p *BasePatch) RemoveMessengerByID(id string) {
	removeIndex := -1
	for index, msgr := range p.messengers {
		if msgr.Identifier() == id {
			removeIndex = index
			break
		}
	}

	if removeIndex > -1 {
		msgr := p.messengers[removeIndex]
		p.messengers = append(p.messengers[:removeIndex], p.messengers[removeIndex+1:]...)
		if receiver, ok := p.receivers[id]; ok {
			if receiver == msgr {
				delete(p.receivers, id)
			}
		}
	}
}

func (p *BasePatch) AddModule(m Module) Module {
	p.subModules = append(p.subModules, m)

	p.AddMessageReceiver(m, m.Identifier())

	return m
}

func (p *BasePatch) AddControl(ct Control) Control {
	p.controls = append(p.controls, ct)

	p.AddMessageReceiver(ct, ct.Identifier())

	return ct
}

func (p *BasePatch) Contains(m Module) bool {
	for _, sub := range p.subModules {
		if sub == m {
			return true
		}
	}

	return false
}

func (p *BasePatch) Lookup(address string) MessageReceiver {
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

	// Send messages for each messenger
	for _, msgr := range p.messengers {
		p.SendMessages(msgr.Messages(p.timestamp, p.Config))
	}

	// Tick for each control rate object
	for _, ticker := range p.controls {
		ticker.Tick(p.timestamp, p.Config)
	}

	// Must synthesize some modules outside normal pull mechanism
	for _, module := range p.subModules {
		if module.MustSynthesize() {
			module.Synthesize()
		}
	}

	// Output modules pull and request synthesize from the connected input modules
	for _, output := range p.outputModules {
		output.Synthesize()
	}

	// Update timestamp
	p.timestamp += int64(p.Config.BufferSize)

	return true
}

func (p *BasePatch) ReceiveMessage(msg any) []*Message {
	content := msg.(map[string]any)

	if command, ok := content["command"]; ok {
		switch command.(string) {
		case "AddMessenger":
			if messenger, ok := content["messenger"]; ok {
				p.AddMessenger(messenger.(Messenger))
			}
		case "RemoveMessenger":
			if messenger, ok := content["messenger"]; ok {
				p.RemoveMessenger(messenger.(Messenger))
			}
		case "RemoveMessengerByID":
			if messenger, ok := content["messenger"]; ok {
				p.RemoveMessengerByID(messenger.(string))
			}
		}
	}

	return nil
}

func (p *BasePatch) SendMessage(msg *Message) {
	rcvr := p.Lookup(msg.Address)
	if rcvr != nil {
		p.SendMessages(rcvr.ReceiveMessage(msg.Content))
	}
}

func (p *BasePatch) SendMessages(msgs []*Message) {
	for _, msg := range msgs {
		p.SendMessage(msg)
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

func (p *BasePatch) InternalInputControl() Control {
	return p.internalInputControl
}

func (p *BasePatch) InternalOutputControl() Control {
	return p.internalOutputControl
}

func (p *BasePatch) ReceiveControlValue(value any, index int) {
	p.internalInputControl.ReceiveControlValue(value, index)
}

func (p *BasePatch) AddControlInputConnection(inputIndex int, sender Control, outputIndex int) {
	p.internalInputControl.AddControlInputConnection(inputIndex, sender, outputIndex)
}

func (p *BasePatch) AddControlOutputConnection(outputIndex int, receiver Control, inputIndex int) {
	p.internalOutputControl.AddControlOutputConnection(outputIndex, receiver, inputIndex)
}
