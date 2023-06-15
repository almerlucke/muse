package main

import "log"

type Classy interface {
	Self() any
	Super() any
}

type Class struct {
	self any
}

func NewClass(self any) *Class {
	return &Class{self: self}
}

func (c *Class) Self() any {
	return c.self
}

func (c *Class) Super() any {
	return nil
}

type Valuer interface {
	Value() int
}

type Test1 struct {
	*Class
	value int
}

func NewTest1(v int) *Test1 {
	t1 := &Test1{
		value: v,
	}

	t1.Class = NewClass(t1)

	return t1
}

func (t *Test1) Value() int {
	return t.value
}

func (t *Test1) Valuer() Valuer {
	return t.Self().(Valuer)
}

type Test2 struct {
	*Test1
}

func NewTest2(v int) *Test2 {
	t2 := &Test2{
		Test1: NewTest1(v),
	}

	t2.Class = NewClass(t2)

	return t2
}

func (t *Test2) Value() int {
	return 12 + t.Test1.value
}

func (t *Test2) Super() any {
	return t.Test1
}

func main() {
	t2 := NewTest2(34)

	log.Printf("value: %d", t2.Valuer().Value())
	log.Printf("super value: %d", t2.Super().(Valuer).Value())
	log.Printf("self value: %d", t2.Super().(Classy).Self().(Valuer).Value())

	/*
		filter :=  filter.Filter(osc.Osc(100.0), sweep.Sweep(40.0, 1500.0, 1.0), 0.5).Connect(osc.Osc).Connect
	*/
}
