einx
------
a framework in golang for game server or app server.

a example server for einx (https://github.com/Cyinx/game_server_einx)

----------------------------------------------------
einx 是一个由 golang 编写的用于游戏服务器或者应用服务器的开源框架。

设计核心：

* 模块与组件的组合机制,模块是逻辑核心。
* lua脚本
* 按业务分离逻辑

----------------------------------------------------

* einx/db 组件化数据库相关操作
* einx/network 组件化网络IO，目前只支持TCP
* einx/log 异步日志库
* einx/timer 时间轮定时器
* einx/module 模块
* einx/component 组件
* einx/lua 脚本相关操作

模块与组件
---------------
  每个模块有且仅有一个goroutine用于处理被投递到本模块中的消息，在模块中的逻辑不需要考虑同步问题，简化了逻辑开发难度，模块与模块之间可以通过RPC交互

使用 einx 搭建一个简单的服务器
----------------------------------
首先安装 einx
```
go get github.com/go-sql-driver/mysql
go get github.com/yuin/gopher-lua
go get github.com/Cyinx/einx
```

创建一个简单的einx例子:

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

einx的核心是module，module中可以添加各种component作为组件:
```
Cyinx/einx/network	网络相关的component
Cyinx/einx/db		数据库相关的component
```

创建一个TCPServer的component管理器:

```go
package clientmgr

import (
	"github.com/Cyinx/einx"
	"github.com/Cyinx/einx/slog"
	"msg_def"
)

type Agent = einx.Agent
type AgentID = einx.AgentID
type EventType = einx.EventType
type Component = einx.Component
type ComponentID = einx.ComponentID

type ClientMgr struct {
	client_map map[AgentID]Agent
	tcp_link   Component
}

var Instance = &ClientMgr{
	client_map: make(map[AgentID]Agent),
}

func (this *ClientMgr) GetClient(agent_id AgentID) (Agent, bool) {
	client, ok := this.client_map[agent_id]
	return client, ok
}

func (this *ClientMgr) OnLinkerConneted(id AgentID, agent Agent) {
	this.client_map[id] = agent //新连接连入服务器
}

func (this *ClientMgr) OnLinkerClosed(id AgentID, agent Agent) {
	delete(this.client_map, id) //连接断开
}

func (this *ClientMgr) OnComponentError(c Component, err error) {

}

func (this *ClientMgr) OnComponentCreate(id ComponentID, component Component) {
	this.tcp_link = component
	component.Start()
	slog.LogInfo("tcp", "Tcp sever start success")
}
```


创建一个逻辑module，并将TcpServer管理器加入到module之中，服务器就可以启动，并监听2345端口的请求
```go
package main

import (
	"clientmgr"
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

注册消息handler与Rpc：
注册消息handler需要事先注册一个Message：
```go
package msg_def

import (
	"github.com/Cyinx/einx/network"
	"protobuf_gen"
)

type VersionCheck = pbgen.VersionCheck

var VersionCheckMsgID = network.RegisterMsgProto(uint16(pbgen.MainMsgID_GENERAL_MSG),
	uint16(pbgen.HandlerMsgID_VERSION_CHECK),
	(*VersionCheck)(nil))
```

在注册RPC时，使用字符串作为RPC名，注册handler时，需要使用之前注册的MsgID
```go
import (
	"msg_def"
)
var logic = einx.GetModule("logic")
func InitDBHandler() {
	logic.RegisterRpcHandler("testRpc", testRpc)
	logic.RegisterHandler(msg_def.VersionCheckMsgID, CheckVersion)
}

func testRpc(sender interface{}, args []interface{}) {

}

func CheckVersion(agent Agent, args interface{}) {
	version_check_msg := args.(*msg_def.VersionCheck)	
}

```

注册定时器使用module.AddTimer函数，返回值为timerID，如果要提前停止timer，可以执行module.RemoveTimer(timerid):

```go
import (
	"msg_def"
)
var logic = einx.GetModule("logic")
var testTimerID uint64 = 0

func InitDBHandler() {
	logic.RegisterRpcHandler("testRpc", testRpc)
	logic.RegisterHandler(msg_def.VersionCheckMsgID, CheckVersion)
}

func testRpc(sender interface{}, args []interface{}) {
	if testTimerID != 0 {
	  logic.RemoveTimer(testTimerID)
	}
}

func TestTimer(args []interface{}) {
	testTimerID = 0
}

func CheckVersion(agent Agent, args interface{}) {
	version_check_msg := args.(*msg_def.VersionCheck)	
	testTimerID = logic.AddTimer(1000,TestTimer,1,2,"测试")
}

```