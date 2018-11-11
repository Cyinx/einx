## einx
------
a framework in golang for game server or app server.

a example server for einx (https://github.com/Cyinx/game_server_einx)

### Getting Started
---------
1.Install.

   go get github.com/go-sql-driver/mysql

   go get github.com/yuin/gopher-lua

   go get github.com/Cyinx/einx


2.hello world

```go
package main

import (
	"github.com/Cyinx/einx"
	"github.com/Cyinx/einx/slog"
)
func main() {
	slog.SetLogPath("log/game_server/")
	slog.LogInfo("game_server", "start server...")
	slog.LogInfo("game_server", "hello world...")
	einx.Run()
	einx.Close()
}
```
3.start a tcp server component.

```go
package clientmgr

import (
	"github.com/Cyinx/einx"
	"github.com/Cyinx/einx/slog"
	"msg_def" //this is a package for serialization
)

type Agent = einx.Agent
type AgentID = einx.AgentID
type NetLinker = einx.NetLinker
type EventType = einx.EventType
type Component = einx.Component
type ModuleRouter = einx.ModuleRouter
type ComponentID = einx.ComponentID
type Context = einx.Context
type ProtoTypeID = uint32

var logic = einx.GetModule("logic")
var logic_router = logic.(ModuleRouter)

type ClientMgr struct {
	client_map map[AgentID]*Client
	tcp_link   Component
}

var Instance = &ClientMgr{
	client_map: make(map[AgentID]*Client),
}

func GetClient(agent_id uint64) *Client {
	client, _ := Instance.client_map[AgentID(agent_id)]
	return client
}

func (this *ClientMgr) OnLinkerConneted(id AgentID, agent Agent) {
	this.client_map[id] = &Client{linker: agent.(NetLinker)}
}

func (this *ClientMgr) OnLinkerClosed(id AgentID, agent Agent) {
	delete(this.client_map, id)
}

func (this *ClientMgr) OnComponentError(ctx Context, err error) {

}

func (this *ClientMgr) OnComponentCreate(ctx Context, id ComponentID) {
	component := ctx.GetComponent()
	this.tcp_link = component
	component.Start()
	slog.LogInfo("gate_client", "Tcp sever start success")
}

func (this *ClientMgr) ServeHandler(agent Agent, id ProtoTypeID, b []byte) {
	msg := msg_def.UnmarshalMsg(id, b) //deserialize the msg and send it to the module you want.
	if msg != nil {
		logic_router.RouterMsg(agent, id, msg)
	}

}

func (this *ClientMgr) ServeRpc(agent Agent, id ProtoTypeID, b []byte) {
	msg := msg_def.UnmarshalRpc(id, b) //deserialize the rpc msg and send it to the module you want.
	if msg != nil {
		logic_router.RouterMsg(agent, id, msg)
	}
}

```

4.create a module and start a tcp server component.
```go
package main

import (
	"clientmgr" //clientmgr is the package in step 3.
	"github.com/Cyinx/einx"
	"github.com/Cyinx/einx/slog"
)
var logic = einx.GetModule("logic")
func main() {
	slog.SetLogPath("log/game_server/")
	logic.AddTcpServer(":2345", clientmgr.Instance)
	slog.LogInfo("game_server", "start server...")
	einx.Run()
	einx.Close()
}
```
5.register msg handlers or rpc.
```go
import (
	"msg_def"
)
var logic = einx.GetModule("logic")
func InitDBHandler() {
	logic.RegisterRpcHandler("testRpc", testRpc)
	logic.RegisterHandler(msg_def.VersionCheckMsgID, CheckVersion)
}

func testRpc(ctx Context, args []interface{}) {
    rpcData := args[0].([]byte)
}

func CheckVersion(ctx Context, args interface{}) {
     msg := args.(*msg_def.VersionCheckMsg)
}

```

6.register timer.
```go
import (
	"msg_def"
)
var logic = einx.GetModule("logic")
func InitDBHandler() {
	logic.AddTimer(40 testTimer,1,"test")
}

func testTimer(ctx Context, args []interface{}) {
    dataInt := args[0].(int)
    dataString := args[1].(string)
}
```



## Licensing
---------

Apache License.