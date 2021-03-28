package module

type ModuleContext struct {
	m Module
	s Agent
	c Component
	t interface{}
	v map[int]interface{}
	u chan []interface{}
}

func (c *ModuleContext) Reset() {
	c.s = nil
	c.c = nil
	c.t = nil
	c.u = nil
}

func (c *ModuleContext) GetModule() Module {
	return c.m
}

func (c *ModuleContext) GetSender() Agent {
	return c.s
}

func (c *ModuleContext) GetComponent() Component {
	return c.c
}

func (c *ModuleContext) GetAttach() interface{} {
	return c.t
}

func (c *ModuleContext) Store(k int, v interface{}) {
	if c.v == nil {
		c.v = make(map[int]interface{})
	}
	c.v[k] = v
}

func (c *ModuleContext) Get(k int) interface{} {
	v, ok := c.v[k]
	if ok == true {
		return v
	}
	return nil
}

func (c *ModuleContext) Done(args ...interface{}) {
	if c.u == nil {
		return
	}
	c.u <- args
}

type ArgsVar struct {
	args []interface{}
}

func (s *ArgsVar) init() {
	s.args = make([]interface{}, 0, 4)
}

func (s *ArgsVar) ref(args []interface{}) {
	s.args = args
}

func (s *ArgsVar) clear() {
	s.args = nil
}

func (s *ArgsVar) addParam(i interface{}) {
	s.args = append(s.args, i)
}

func (s *ArgsVar) Length() int {
	return len(s.args)
}

func (s *ArgsVar) ReadBool(i int) bool {
	if len(s.args) <= i {
		return false
	}
	return s.args[i].(bool)
}

func (s *ArgsVar) ReadInt(i int) int {
	if len(s.args) <= i {
		return 0
	}
	return s.args[i].(int)
}

func (s *ArgsVar) ReadInt32(i int) int32 {
	if len(s.args) <= i {
		return 0
	}
	return s.args[i].(int32)
}

func (s *ArgsVar) ReadInt64(i int) int64 {
	if len(s.args) <= i {
		return 0
	}
	return s.args[i].(int64)
}

func (s *ArgsVar) ReadUInt64(i int) uint64 {
	if len(s.args) <= i {
		return 0
	}
	return s.args[i].(uint64)
}

func (s *ArgsVar) ReadDouble(i int) float64 {
	if len(s.args) <= i {
		return 0
	}
	return s.args[i].(float64)
}

func (s *ArgsVar) ReadString(i int) string {
	if len(s.args) <= i {
		return ""
	}
	return s.args[i].(string)
}

func (s *ArgsVar) Read(i int) interface{} {
	if len(s.args) <= i {
		return nil
	}
	return s.args[i]
}
