# menet

[English](README.md)

Go 语言实时多人游戏服务器中间件集合。提供网络层（TCP/WebSocket）、数据持久化、社交平台登录、排行榜、RPC、并发数据结构等模块，专为游戏服务器场景设计。

[![CI](https://github.com/chenwbyx/menet/actions/workflows/ci.yml/badge.svg)](https://github.com/chenwbyx/menet/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/chenwbyx/menet.svg)](https://pkg.go.dev/github.com/chenwbyx/menet)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/chenwbyx/menet)](go.mod)

## 特性

- **TCP & WebSocket 网络层** — 每连接独立 Session，支持 protobuf/JSON 消息分发、服务器推送、优雅关闭
- **数据持久化框架** — 基于代码生成的 ORM，内存缓存 + 异步回写 MySQL，支持按用户加载/卸载生命周期，Bomb 文件崩溃恢复
- **可插拔社交登录** — 驱动模式适配器，内置 QQ、微信、阿里、游客等平台支持
- **排行榜系统** — 内存跳表 + Redis Sorted Set 双实现
- **RPC 框架** — 基于 `net/rpc` 封装，自动重连，AST 代码生成客户端桩
- **无锁队列** — Michael-Scott 非阻塞队列算法，附带 7 种并发队列策略的性能对比
- **代码生成器** — `proto_handle_gen` 自动生成协议处理器注册，`persist` 自动生成持久化管理器

## 架构

```
Client ←→ TCP / WebSocket ←→ Server (TcpServer / WsServer)
                                     │
                               Session (读协程 + 写协程)
                                     │
                               协议解码 (长度前缀 protobuf / JSON)
                                     │
                               消息分发 (msgNo → ProtoFunc 处理函数)
                                     │
                               Handler → 响应 → 编码 → 写入
                                     │
                               PushToUser / PushToAll / Broadcast
```

## 模块说明

| 模块 | 说明 | 核心类型 |
|------|------|----------|
| `menet`（根包） | TCP 服务器/客户端框架，protobuf 二进制协议 | `TcpServer`, `TcpSession`, `TcpClient`, `ProtoFunc` |
| `mewnet/` | WebSocket 网络层（v2），支持 TLS、zlib 压缩、防重放 | `WsServer`, `WsSession`, `WsClient`, `WsSessionHub` |
| `persist/` | 代码生成持久化框架，内存缓存 + 异步 MySQL 回写 | `IPersist`, `IPersistUser`, `Load()`, `Unload()` |
| `merpc/` | RPC 框架，自动重连 + AST 代码生成 | `RpcServer`, `RpcClient` |
| `login/` | 可插拔社交登录验证（QQ、微信、阿里、游客等） | `ILogin`, `RegisterDriver()`, `Validate()` |
| `lockfree_queue/` | 7 种并发队列实现 + benchmark 对比 | `IntQueue`（无锁）, `LockIntQueue`, `ChanIntQueue` |
| `ranking/` | 内存排行榜，基于跳表（O(log n) 排名查询） | `RankList`, `RankItem` |
| `ranklist/` | Redis 排行榜，基于 Sorted Set | `RankList`, `GetRankList()` |
| `crypto/` | AES-ECB 加密 + PKCS5/PKCS7 填充 | `EncryptECB()`, `DecryptECB()` |
| `pubsub/` | Redis 发布/订阅封装 | `PubSub`, `Subscribe()`, `Publish()` |
| `proto_handle_gen/` | 代码生成器：扫描 `// proto: N` 注释，生成处理器注册代码 | `go generate` 工具 |

## 快速开始

### TCP 服务器

```go
package main

import (
    "fmt"
    "net"
    "menet"
)

func main() {
    // 注册消息处理函数，消息号 1001
    menet.Register(1001, func(session *menet.TcpSession, msg *menet.ProtobufMessage) []byte {
        fmt.Printf("收到来自 %s 的消息: %v\n", session.RemoteAddr(), msg.Body)
        return nil // 返回 nil 表示不回复，或返回编码后的响应
    })

    // 启动 TCP 服务器
    listener, _ := net.Listen("tcp", ":9000")
    server := menet.NewTcpServer(listener)

    server.AtSessionClose(func(sid int32) {
        fmt.Printf("会话 %d 断开连接\n", sid)
    })

    server.Start() // 阻塞运行
}
```

### TCP 客户端

```go
client := menet.NewTcpClient("localhost:9000")

// 同步请求-响应（5 秒超时）
err := client.Request(1001, reqProto, respProto)

// 异步：注册回调处理服务器推送消息
client.Handle(2001, &pb.ServerNotify{}, func(msgNo uint16, msg proto.Message) {
    fmt.Println("服务器推送:", msg)
})
```

### WebSocket 服务器

```go
wsServer := mewnet.NewWsServer(":8080")
wsServer.Register(1001, func(session *mewnet.WsSession, msg *mewnet.ProtoMessage) []byte {
    // 处理消息
    return nil
})

// 可选：启用 TLS
wsServer.SetCert("cert.pem", "key.pem")

wsServer.Start()
```

## 代码生成器

### proto_handle_gen

在处理函数上添加 `// proto: <消息号>` 注释：

```go
// proto: 1001
func HandleLogin(session *menet.TcpSession, req *pb.LoginReq, resp *pb.LoginResp) {
    resp.Token = "abc123"
}
```

运行 `go generate` 自动生成注册代码：

```go
//go:generate proto_handle_gen
```

生成结果：一个 `RegisterProtoHandles(server)` 函数，自动注册所有标注的处理函数。

### persist

定义数据结构并添加 struct tag：

```go
type PlayerItem struct {
    Uid    int32 `xorm:"pk"`
    ItemId int32 `xorm:"pk"`
    Count  int32
    Group  int32 `hash:"group=1;unique=0"`
}
```

运行 persist 生成器，自动生成完整的持久化管理器：
- 内存哈希索引
- 异步回写队列
- 按用户加载/卸载生命周期
- 操作合并（Insert+Update → Insert，Insert+Delete → 抵消）
- Bomb 文件崩溃恢复

## 性能测试

无锁队列 vs 其他实现（5000 协程 × 每个 10 条数据）：

| 实现 | 策略 | ns/op |
|------|------|-------|
| `IntQueue` | 无锁（Michael-Scott CAS） | ~9.3M |
| `IntQueueSignal` | 无锁 + sync.Cond | ~9.0M |
| `IntQueueChan` | 无锁 + channel | ~9.1M |
| `LockIntQueue` | 互斥锁 + slice | ~12.5M |
| `LockIntRingBuf` | 互斥锁 + 环形缓冲区 | ~12.3M |
| `ChanIntQueue` | Channel 双缓冲交换 | ~31.4M |
| `ChanIntRingBuf` | Channel 双缓冲 + 环形缓冲区 | ~31.6M |

**无锁队列在高并发场景下比 channel 方案快约 3 倍。**

运行 benchmark：

```bash
cd lockfree_queue
go test -bench=. -benchmem
```

## 安装

```bash
go get github.com/chenwbyx/menet
```

## 参与贡献

欢迎贡献代码！流程：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 开源协议

本项目基于 MIT 协议开源 — 详见 [LICENSE](LICENSE) 文件。
