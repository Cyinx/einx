package einx

import (
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/network"
	"github.com/Cyinx/einx/slog"
)

type Option = func(...interface{})

func LogPath(p string) Option {
	return func(args ...interface{}) {
		slog.SetLogPath(p)
	}
}

func KeepAlive(open bool, pingTime int64) Option {
	return func(args ...interface{}) {
		network.SetKeepAlive(open, pingTime)
	}
}

func Perfomance(b bool) Option {
	return func(args ...interface{}) {
		module.PerfomancePrint = b
	}
}

type networkOpt struct {
	Name              func(string) Option
	Module            func(string) Option
	ListenAddr        func(string) Option
	ServeHandler      func(SessionHandler) Option
	TransportMaxCount func(int) Option
}

var NetworkOption networkOpt = networkOpt{
	Name: network.Name,
	Module: func(s string) Option {
		m := GetModule(s)
		return network.Module(m.(event.EventReceiver))
	},
	ListenAddr:        network.ListenAddr,
	ServeHandler:      network.ServeHandler,
	TransportMaxCount: network.TransportMaxCount,
}
