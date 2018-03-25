package lua_state

import (
	"github.com/Cyinx/einx/slog"
	"github.com/yuin/gopher-lua"
	"strings"
	"time"
)

var start_tick time.Time

func init() {
	start_tick = time.Now()
}

func getIntField(L *lua.LState, tb *lua.LTable, key string, v int) int {
	ret := tb.RawGetString(key)
	if ln, ok := ret.(lua.LNumber); ok {
		return int(ln)
	}
	return v
}

func getBoolField(L *lua.LState, tb *lua.LTable, key string, v bool) bool {
	ret := tb.RawGetString(key)
	if lb, ok := ret.(lua.LBool); ok {
		return bool(lb)
	}
	return v
}

func OpenOsRuntime(L *lua.LState) int {
	osmod := L.RegisterModule(lua.OsLibName, osFuncs)
	L.Push(osmod)
	return 1
}

var osFuncs = map[string]lua.LGFunction{
	"clock":    osClock,
	"difftime": osDiffTime,
	"date":     osDate,
	"time":     osTime,
}

func osClock(L *lua.LState) int {
	L.Push(lua.LNumber(float64(time.Now().Sub(start_tick)) / float64(time.Second)))
	return 1
}

func osDiffTime(L *lua.LState) int {
	L.Push(lua.LNumber(L.CheckInt64(1) - L.CheckInt64(2)))
	return 1
}

func osDate(L *lua.LState) int {
	t := time.Now().UTC()
	cfmt := "%c"
	if L.GetTop() >= 1 {
		cfmt = L.CheckString(1)
		if strings.HasPrefix(cfmt, "!") {
			t = time.Now().UTC()
			cfmt = strings.TrimLeft(cfmt, "!")
		}
		if L.GetTop() >= 2 {
			t = time.Unix(L.CheckInt64(2), 0)
		}
		if strings.HasPrefix(cfmt, "*t") {
			ret := L.NewTable()
			ret.RawSetString("year", lua.LNumber(t.Year()))
			ret.RawSetString("month", lua.LNumber(t.Month()))
			ret.RawSetString("day", lua.LNumber(t.Day()))
			ret.RawSetString("hour", lua.LNumber(t.Hour()))
			ret.RawSetString("min", lua.LNumber(t.Minute()))
			ret.RawSetString("sec", lua.LNumber(t.Second()))
			ret.RawSetString("wday", lua.LNumber(t.Weekday()))
			ret.RawSetString("yday", lua.LNumber(t.YearDay()))
			// TODO dst
			ret.RawSetString("isdst", lua.LFalse)
			L.Push(ret)
			return 1
		}
	}
	L.Push(lua.LString(strftime(t, cfmt)))
	return 1
}

func osTime(L *lua.LState) int {
	if L.GetTop() == 0 {
		L.Push(lua.LNumber(time.Now().UTC().Unix()))
	} else {
		tbl := L.CheckTable(1)
		sec := getIntField(L, tbl, "sec", 0)
		min := getIntField(L, tbl, "min", 0)
		hour := getIntField(L, tbl, "hour", 12)
		day := getIntField(L, tbl, "day", -1)
		month := getIntField(L, tbl, "month", -1)
		year := getIntField(L, tbl, "year", -1)
		isdst := getBoolField(L, tbl, "isdst", false)
		t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
		// TODO dst
		if false {
			print(isdst)
		}
		L.Push(lua.LNumber(t.UTC().Unix()))
	}
	return 1
}

func luaPrint(L *lua.LState) int {
	if L.GetTop() < 1 {
		return 0
	}

	s := L.CheckString(1)
	slog.LogInfo("lua_print", s)
	return 1
}
