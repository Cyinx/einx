package lua_state

import (
	"fmt"
	"github.com/Cyinx/einx/slog"
	"github.com/yuin/gopher-lua"
	"math"
)

type LValue = lua.LValue
type LTable = lua.LTable
type LNumber = lua.LNumber

type LuaRuntime struct {
	lua *lua.LState
}

func (this *LuaRuntime) GetVm() *lua.LState {
	return this.lua
}
func NewLuaStae() *LuaRuntime {
	vm := lua.NewState(lua.Options{
		CallStackSize:       4096,
		RegistrySize:        4096,
		SkipOpenLibs:        true,
		IncludeGoStackTrace: true,
	})

	std_libs := map[string]lua.LGFunction{
		lua.LoadLibName:   lua.OpenPackage,
		lua.BaseLibName:   lua.OpenBase,
		lua.TabLibName:    lua.OpenTable,
		lua.OsLibName:     OpenOsRuntime,
		lua.StringLibName: lua.OpenString,
		lua.MathLibName:   lua.OpenMath,
	}

	for name, lib := range std_libs {
		vm.Push(vm.NewFunction(lib))
		vm.Push(lua.LString(name))
		vm.Call(1, 0)
	}

	vm.SetGlobal("print", vm.NewFunction(luaPrint))
	vm.SetGlobal("lua_unmarshal", vm.NewFunction(luaUnMarshal))
	vm.SetGlobal("lua_marshal", vm.NewFunction(luaMarshal))

	runtime := &LuaRuntime{
		lua: vm,
	}
	return runtime
}

func (this *LuaRuntime) DoFile(path string) {
	if err := this.lua.DoFile(path); err != nil {
		slog.LogError("lua", "lua dofile error:%v", err)
	}
}

func ConvertMap(l *lua.LState, data map[string]interface{}) *lua.LTable {
	lt := l.NewTable()

	for k, v := range data {
		lt.RawSetString(k, convertValue(l, v))
	}

	return lt
}

func ConvertLuaTable(lv *lua.LTable) map[string]interface{} {
	returnData, _ := ConvertLuaValue(lv).(map[string]interface{})
	return returnData
}

func convertValue(l *lua.LState, val interface{}) lua.LValue {
	if val == nil {
		return lua.LNil
	}
	switch v := val.(type) {
	case lua.LValue:
		return v
	case bool:
		return lua.LBool(v)
	case string:
		return lua.LString(v)
	case []byte:
		ud := l.NewUserData()
		ud.Value = v
		return ud
	case float32:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case int32:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case uint32:
		return lua.LNumber(v)
	case uint64:
		return lua.LNumber(v)
	case map[string]interface{}:
		return ConvertMap(l, v)
	case []interface{}:
		lt := l.NewTable()
		for k, v := range v {
			lt.RawSetInt(k+1, convertValue(l, v))
		}
		return lt
	default:
		return nil
	}
}

func ConvertLuaValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case *lua.LUserData:
		return v.Value
	case lua.LBool:
		return bool(v)
	case lua.LString:
		return string(v)
	case lua.LNumber:
		f64i := float64(v)
		I64i := int64(v)
		if f64i == float64(I64i) {
			return I64i
		}
		return f64i
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			ret := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				keystr := fmt.Sprint(ConvertLuaValue(key))
				ret[keystr] = ConvertLuaValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, ConvertLuaValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		slog.LogError("lua", "error lua type %v", lv)
		panic("error lua type")
	}
}

func (this *LuaRuntime) PCall(f string, args ...interface{}) {
	l := this.lua
	l.Push(l.GetGlobal(f))
	for _, arg := range args {
		val := convertValue(l, arg)
		l.Push(val)
	}
	if err := l.PCall(len(args), -1, nil); err != nil {
		slog.LogError("lua", "lua pcall err:%v", err)
	}
}

func (this *LuaRuntime) PCall2(f string, args ...LValue) {
	l := this.lua
	l.Push(l.GetGlobal(f))
	for _, arg := range args {
		l.Push(arg)
	}
	if err := l.PCall(len(args), -1, nil); err != nil {
		slog.LogError("lua", "lua pcall2 err:%v", err)
	}
}

func (this *LuaRuntime) PCall3(f LValue, args ...LValue) {
	l := this.lua
	l.Push(f)
	for _, arg := range args {
		l.Push(arg)
	}
	if err := l.PCall(len(args), -1, nil); err != nil {
		slog.LogError("lua", "lua pcall3 err:%v", err)
	}
}

func (this *LuaRuntime) GetGlobal(f string) LValue {
	return this.lua.GetGlobal(f)
}

func (this *LuaRuntime) RegisterFunction(s string, f func(*lua.LState) int) {
	l := this.lua
	l.SetGlobal(s, l.NewFunction(f))
}

func Marshal(b []byte, lv lua.LValue) []byte {
	var buffer []byte = nil
	switch v := lv.(type) {
	case *lua.LNilType:
		buffer = append(b, 'z')
	case lua.LBool:
		if v == true {
			buffer = append(b, 't')
		} else {
			buffer = append(b, 'f')
		}
	case lua.LString:
		slen := uint32(len(v))
		buffer = append(b, 's', byte(slen), byte(slen>>8), byte(slen>>16), byte(slen>>24))
		buffer = append(buffer, v...)
	case lua.LNumber:
		f64i := float64(v)
		I64i := int64(v)
		if f64i == float64(I64i) {
			buffer = append(b, 'i')
			ux := uint64(I64i) << 1
			if I64i < 0 {
				ux = ^ux
			}
			for ux >= 0x80 {
				buffer = append(buffer, byte(ux)|0x80)
				ux >>= 7
			}
			buffer = append(buffer, byte(ux))
		} else {
			n := math.Float64bits(float64(v))
			buffer = append(b, 'd', byte(n), byte(n>>8), byte(n>>16), byte(n>>24), byte(n>>32), byte(n>>40), byte(n>>48), byte(n>>56))
		}
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 {
			buffer = append(b, '[')
			v.ForEach(func(key, value lua.LValue) {
				buffer = Marshal(Marshal(buffer, key), value)
			})
			buffer = append(buffer, ']')
		} else {
			buffer = append(b, '{')
			for i := 1; i <= maxn; i++ {
				buffer = Marshal(buffer, v.RawGetInt(i))
			}
			buffer = append(buffer, '}')
		}
	default:
		slog.LogError("lua", "error lua type %v", lv)
		panic("error lua type")
	}
	return buffer
}

func UnMarshal(b []byte, l *lua.LState) (lua.LValue, []byte) {
	if len(b) < 1 {
		return lua.LNil, b
	}
	t := b[0]
	switch t {
	case 'z':
		return lua.LNil, b[1:]
	case 't':
		return lua.LBool(true), b[1:]
	case 'f':
		return lua.LBool(false), b[1:]
	case 's':
		if len(b) < 5 {
			slog.LogWarning("lua", "error:unknow unmarshal string")
			return lua.LNil, b
		}
		slen := uint32(b[1]) | uint32(b[2])<<8 | uint32(b[3])<<16 | uint32(b[4])<<24
		return lua.LString(b[5 : 5+slen]), b[5+slen:]
	case 'd':
		if len(b) < 9 {
			slog.LogWarning("lua", "error:unknow unmarshal number")
			return lua.LNil, b
		}
		n := uint64(b[1]) | uint64(b[2])<<8 | uint64(b[3])<<16 | uint64(b[4])<<24 |
			uint64(b[5])<<32 | uint64(b[6])<<40 | uint64(b[7])<<48 | uint64(b[8])<<56
		return lua.LNumber(math.Float64frombits(n)), b[9:]
	case 'i':
		length := len(b)
		if length < 2 {
			slog.LogWarning("lua", "error:unknow unmarshal varint")
			return lua.LNil, b
		}
		var ux uint64
		var s uint32
		i := 1
		for ; i < 9; i++ {
			if i >= length {
				slog.LogWarning("lua", "error:unknow unmarshal varint:overflow")
				return lua.LNil, b
			}
			m := b[i]
			if m < 0x80 {
				if i == 9 && m > 1 {
					slog.LogWarning("lua", "error:unknow unmarshal varint:overflow")
					return lua.LNil, b
				}
				ux |= uint64(m) << s
				break
			}
			ux |= uint64(m&0x7f) << s
			s += 7
		}
		x := int64(ux >> 1)
		if ux&1 != 0 {
			x = ^x
		}
		return lua.LNumber(x), b[i+1:]
	case '[':
		var key lua.LValue
		var val lua.LValue
		tb := b[1:]
		lt := l.NewTable()
		for tb[0] != ']' {
			key, tb = UnMarshal(tb, l)
			val, tb = UnMarshal(tb, l)
			lt.RawSet(key, val)
		}
		return lt, tb[1:]
	case '{':
		var val lua.LValue
		index := 1
		tb := b[1:]
		lt := l.NewTable()
		for tb[0] != '}' {
			val, tb = UnMarshal(tb, l)
			lt.RawSetInt(index, val)
			index++
		}
		return lt, tb[1:]
	default:
		slog.LogError("lua", "error lua type %v", t)
		panic("error lua type")
	}
	return lua.LNil, b
}
