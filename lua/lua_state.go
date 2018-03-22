package lua_state

import (
	"fmt"
	"github.com/Cyinx/einx/slog"
	"github.com/yuin/gopher-lua"
	"math"
)

type LuaRuntime struct {
	lua *lua.LState
}

type LuaTable struct {
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

	runtime := &LuaRuntime{
		lua: vm,
	}
	return runtime
}

func (this *LuaRuntime) LoadFile(path string) {
	if err := this.lua.DoFile(path); err != nil {
		slog.LogError("lua", "lua load file err:%v", err)
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
	returnData, _ := convertLuaValue(lv).(map[string]interface{})
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
		return lua.LString(v)
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

func convertLuaValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			ret := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				keystr := fmt.Sprint(convertLuaValue(key))
				ret[keystr] = convertLuaValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, convertLuaValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		return v
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

func (this *LuaRuntime) DoFile(f string) {
	l := this.lua
	if err := l.DoFile(f); err != nil {
		slog.LogError("lua", "lua pcall err:%v", err)
	}
}

func (this *LuaRuntime) RegisterFunction(s string, f func(*lua.LState) int) {
	l := this.lua
	l.SetGlobal(s, l.NewFunction(f))
}

func (this *LuaRuntime) Marshal(b []byte, lv lua.LValue) []byte {
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
		buffer = append(buffer, 's', byte(slen), byte(slen>>8), byte(slen>>16), byte(slen>>24))
		buffer = append(buffer, v...)
	case lua.LNumber:
		n := math.Float64bits(float64(v))
		buffer = append(b, 'n', byte(n), byte(n>>8), byte(n>>16), byte(n>>24), byte(n>>32), byte(n>>40), byte(n>>48), byte(n>>56))
	case *lua.LTable:
		buffer = append(b, '[')
		v.ForEach(func(key, value lua.LValue) {
			buffer = this.Marshal(buffer, key)
			buffer = this.Marshal(buffer, value)
		})
		buffer = append(buffer, ']')
	default:
		break
	}
	return buffer
}

func (this *LuaRuntime) UnMarshal(b []byte) (lua.LValue, []byte) {
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
		return lua.LString(b[5:]), b[5+slen:]
	case 'n':
		if len(b) < 9 {
			slog.LogWarning("lua", "error:unknow unmarshal number")
			return lua.LNil, b
		}
		n := uint64(b[1]) | uint64(b[2])<<8 | uint64(b[3])<<16 | uint64(b[4])<<24 |
			uint64(b[5])<<32 | uint64(b[6])<<40 | uint64(b[7])<<48 | uint64(b[8])<<56
		return lua.LNumber(math.Float64frombits(n)), b[9:]
	case '[':
		var key lua.LValue
		var val lua.LValue
		tb := b[1:]
		lt := this.lua.NewTable()
		for tb[0] != ']' {
			key, tb = this.UnMarshal(tb)
			val, tb = this.UnMarshal(tb)
			lt.RawSet(key, val)
		}
		return lt, tb[1:]
	}
	return lua.LNil, b
}
