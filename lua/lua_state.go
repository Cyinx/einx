package lua_state

import (
	"fmt"
	"github.com/Cyinx/einx/slog"
	"github.com/yuin/gopher-lua"
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

func (this *LuaRuntime) Marshal(b []byte, lv lua.LValue) []byte {
	var buffer []byte = nil
	switch v := lv.(type) {
	case *lua.LNilType:
		buffer = append(b, '0')
	case lua.LBool:
		buffer = append(b, 'b', byte(v))
	case lua.LString:
		buffer = append(b, 's', string(v))
	case lua.LNumber:
		buffer = append(b, 'n', float64(v))
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			buffer = append(b, 't')
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
		break
	}
	return buffer
}
