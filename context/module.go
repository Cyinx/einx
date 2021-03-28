package context

import (
	"github.com/Cyinx/einx/timer"
)

type TimerHandler = timer.TimerHandler
type Module interface {
	GetID() AgentID
	GetName() string
	RpcCall(string, ...interface{})
	AwaitRpcCall(string, ...interface{}) []interface{}
	AddTimer(delay uint64, op TimerHandler, args ...interface{}) uint64
	RemoveTimer(timer_id uint64) bool
}
