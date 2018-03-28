einx
------
a framework in golang for game server or app server.

a example server for einx (https://github.com/Cyinx/game_server_einx)

----------------------------------------------------
einx 是一个由 golang 编写的用于游戏服务器或者应用服务器的开源手脚架。

设计目的:

采用较为传统的以业务划分并发模块的方式降低逻辑开发难度，在逻辑开发与执行效率间取得平衡

设计核心：

* 模块与组件的组合机制,模块是逻辑核心。
* 引入lua脚本
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
