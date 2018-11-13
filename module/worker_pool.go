package module

import (
	"fmt"
	"github.com/Cyinx/einx/slog"
	"hash/crc32"
	"sync"
)

var worker_pools_map sync.Map

type WorkerPool interface {
	ForEachModule(func(m Module))
	RegisterRpcHandler(string, RpcHandler)
	RegisterHandler(ProtoTypeID, MsgHandler)
	RpcCall(string, ...interface{})
}

type ModuleWorkerPool struct {
	modules    []Module
	name       string
	balance_id uint32
	size       uint32
}

func CreateWorkers(name string, size int) WorkerPool {
	w := GetWorkerPool(name).(*ModuleWorkerPool)
	w.size = uint32(size)
	if w.modules == nil {
		w.modules = make([]Module, size)
	}
	for i := 0; i < size; i++ {
		m := NewModule(fmt.Sprintf("%s_worker_%d", name, i+1))
		w.modules[i] = m
	}
	return w
}

func GetWorkerPool(name string) WorkerPool {
	v, ok := worker_pools_map.Load(name)
	if ok == true {
		return v.(WorkerPool)
	} else {
		w := &ModuleWorkerPool{
			name:       name,
			balance_id: 0,
			size:       0,
		}
		worker_pools_map.Store(name, w)
		return w
	}
}

func (this *ModuleWorkerPool) Start() {
	for _, m := range this.modules {
		go func(m Module) { m.(ModuleWoker).Run(&wait_close) }(m)
	}
}

func (this *ModuleWorkerPool) ForEachModule(f func(m Module)) {
	for _, v := range this.modules {
		f(v)
	}
}

func (this *ModuleWorkerPool) Close() {
	slog.LogInfo("worker_pool", "worker_pool [%v] will close.", this.name)
	for _, v := range this.modules {
		v.(ModuleWoker).Close()
	}
}

func (this *ModuleWorkerPool) RegisterRpcHandler(name string, f RpcHandler) {
	for _, v := range this.modules {
		v.(ModuleRouter).RegisterRpcHandler(name, f)
	}
}

func (this *ModuleWorkerPool) RegisterHandler(type_id ProtoTypeID, f MsgHandler) {
	for _, v := range this.modules {
		v.(ModuleRouter).RegisterHandler(type_id, f)
	}
}

func (this *ModuleWorkerPool) RpcCall(name string, args ...interface{}) {
	crcValue := crc32.ChecksumIEEE([]byte(name))
	m := this.modules[crcValue&this.size]
	m.RpcCall(name, args...)
}
