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

func Perfomance(b bool) Option {
	return func(args ...interface{}) {
		module.PerfomancePrint = b
	}
}

func OnClosing(f func()) Option {
	return func(args ...interface{}) {
		_einxDefault.onClose = f
	}
}

type networkOpt struct {
	Name               func(string) Option
	Module             func(string) Option
	ListenAddr         func(string) Option
	ServeHandler       func(SessionHandler) Option
	TransportMaxCount  func(int) Option
	TransportMaxLength func(int) Option
	TransportKeepAlive func(bool, int64) Option
}

var NetworkOption networkOpt = networkOpt{
	Name: network.Name,
	Module: func(s string) Option {
		m := GetModule(s)
		return network.Module(m.(event.EventReceiver))
	},
	ListenAddr:         network.ListenAddr,
	ServeHandler:       network.ServeHandler,
	TransportMaxCount:  network.TransportMaxCount,
	TransportMaxLength: network.TransportMaxLength,
	TransportKeepAlive: network.TransportKeepAlive,
}
