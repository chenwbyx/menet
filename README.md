# menet

[中文文档](README_zh.md)

A collection of server-side middleware for building real-time multiplayer game servers in Go. Provides networking (TCP/WebSocket), data persistence, social login, ranking, RPC, and concurrent data structures — all designed for the demands of game server workloads.

[![CI](https://github.com/chenwbyx/menet/actions/workflows/ci.yml/badge.svg)](https://github.com/chenwbyx/menet/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/chenwbyx/menet.svg)](https://pkg.go.dev/github.com/chenwbyx/menet)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/chenwbyx/menet)](go.mod)

## Features

- **TCP & WebSocket networking** — session-per-connection with protobuf/JSON message dispatch, server push, and clean shutdown
- **Data persistence framework** — code-generating ORM with in-memory caching, async write-back to MySQL, per-user load/unload lifecycle, and crash recovery via bomb files
- **Pluggable social login** — driver-based adapter system for QQ, WeChat, Alibaba, guest, and custom platforms
- **Ranking systems** — in-memory skip list and Redis sorted set implementations
- **RPC framework** — thin wrapper over `net/rpc` with auto-reconnect and code-generated client stubs
- **Lock-free queue** — Michael-Scott non-blocking queue algorithm with benchmarks comparing 7 concurrent queue strategies
- **Code generators** — `proto_handle_gen` for protocol handler registration, `persist` for persistence manager generation

## Architecture

```
Client ←→ TCP / WebSocket ←→ Server (TcpServer / WsServer)
                                     │
                               Session (read + write goroutines)
                                     │
                               Protocol Decode (length-prefixed protobuf / JSON)
                                     │
                               Message Dispatch (msgNo → ProtoFunc handler)
                                     │
                               Handler → Response → Encode → Write
                                     │
                               PushToUser / PushToAll / Broadcast
```

## Modules

| Module | Description | Core Types |
|--------|-------------|------------|
| `menet` (root) | TCP server/client framework with protobuf binary protocol | `TcpServer`, `TcpSession`, `TcpClient`, `ProtoFunc` |
| `mewnet/` | WebSocket networking (v2) with TLS, zlib compression, anti-replay | `WsServer`, `WsSession`, `WsClient`, `WsSessionHub` |
| `persist/` | Code-generating persistence framework with in-memory cache + async MySQL write-back | `IPersist`, `IPersistUser`, `Load()`, `Unload()` |
| `merpc/` | RPC framework with auto-reconnect and AST-based code generation | `RpcServer`, `RpcClient` |
| `login/` | Pluggable social login validation (QQ, WeChat, Alibaba, guest, etc.) | `ILogin`, `RegisterDriver()`, `Validate()` |
| `lockfree_queue/` | 7 concurrent queue implementations with benchmarks | `IntQueue` (lock-free), `LockIntQueue`, `ChanIntQueue` |
| `ranking/` | In-memory ranking with skip list (O(log n) rank queries) | `RankList`, `RankItem` |
| `ranklist/` | Redis-backed ranking via sorted sets | `RankList`, `GetRankList()` |
| `crypto/` | AES-ECB encryption with PKCS5/PKCS7 padding | `EncryptECB()`, `DecryptECB()` |
| `pubsub/` | Redis pub/sub wrapper for inter-process messaging | `PubSub`, `Subscribe()`, `Publish()` |
| `proto_handle_gen/` | Code generator: scans `// proto: N` comments, generates handler registration | `go generate` tool |

## Quick Start

### TCP Server

```go
package main

import (
    "fmt"
    "net"
    "menet"
)

func main() {
    // Register message handler for msgNo 1001
    menet.Register(1001, func(session *menet.TcpSession, msg *menet.ProtobufMessage) []byte {
        fmt.Printf("Received from %s: %v\n", session.RemoteAddr(), msg.Body)
        return nil // return nil for no response, or return encoded response bytes
    })

    // Start TCP server
    listener, _ := net.Listen("tcp", ":9000")
    server := menet.NewTcpServer(listener)

    server.AtSessionClose(func(sid int32) {
        fmt.Printf("Session %d disconnected\n", sid)
    })

    server.Start() // blocks
}
```

### TCP Client

```go
client := menet.NewTcpClient("localhost:9000")

// Synchronous request-response (5s timeout)
err := client.Request(1001, reqProto, respProto)

// Async: register callback for server-push messages
client.Handle(2001, &pb.ServerNotify{}, func(msgNo uint16, msg proto.Message) {
    fmt.Println("Server push:", msg)
})
```

### WebSocket Server

```go
wsServer := mewnet.NewWsServer(":8080")
wsServer.Register(1001, func(session *mewnet.WsSession, msg *mewnet.ProtoMessage) []byte {
    // Handle message
    return nil
})

// Optional: enable TLS
wsServer.SetCert("cert.pem", "key.pem")

wsServer.Start()
```

## Code Generators

### proto_handle_gen

Annotate handler functions with `// proto: <msgNo>` comments:

```go
// proto: 1001
func HandleLogin(session *menet.TcpSession, req *pb.LoginReq, resp *pb.LoginResp) {
    resp.Token = "abc123"
}
```

Run `go generate` to produce registration code:

```go
//go:generate proto_handle_gen
```

Generated output: a `RegisterProtoHandles(server)` function that auto-registers all annotated handlers.

### persist

Define your data struct with struct tags:

```go
type PlayerItem struct {
    Uid   int32  `xorm:"pk"`
    ItemId int32 `xorm:"pk"`
    Count int32
    Group int32  `hash:"group=1;unique=0"`
}
```

Run the persist generator to produce a complete persistence manager with:
- In-memory hash indexes
- Async write-back queue
- Load/unload lifecycle per user
- Operation merging (Insert+Update → Insert, Insert+Delete → noop)
- Bomb file crash recovery

## Benchmarks

Lock-free queue vs alternatives (5000 goroutines × 10 items each):

| Implementation | Strategy | ns/op |
|---------------|----------|-------|
| `IntQueue` | Lock-free (Michael-Scott CAS) | ~9.3M |
| `IntQueueSignal` | Lock-free + sync.Cond | ~9.0M |
| `IntQueueChan` | Lock-free + channel | ~9.1M |
| `LockIntQueue` | Mutex + slice | ~12.5M |
| `LockIntRingBuf` | Mutex + ring buffer | ~12.3M |
| `ChanIntQueue` | Channel double-buffer swap | ~31.4M |
| `ChanIntRingBuf` | Channel double-buffer + ring buffer | ~31.6M |

**Lock-free queues are ~3x faster** than channel-based approaches under high concurrency.

Run benchmarks:

```bash
cd lockfree_queue
go test -bench=. -benchmem
```

## Installation

```bash
go get github.com/chenwbyx/menet
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
