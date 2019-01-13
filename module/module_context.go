package module

type ModuleContext struct {
	m Module
	s Agent
	c Component
	t interface{}
	v map[int]interface{}
}

func (this *ModuleContext) Reset() {
	this.s = nil
	this.c = nil
	this.t = nil
}

func (this *ModuleContext) GetModule() Module {
	return this.m
}

func (this *ModuleContext) GetSender() Agent {
	return this.s
}

func (this *ModuleContext) GetComponent() Component {
	return this.c
}

func (this *ModuleContext) GetAttach() interface{} {
	return this.t
}

func (this *ModuleContext) Store(k int, v interface{}) {
	if this.v == nil {
		this.v = make(map[int]interface{})
	}
	this.v[k] = v
}

func (this *ModuleContext) Get(k int) interface{} {
	v, ok := this.v[k]
	if ok == true {
		return v
	}
	return nil
}
