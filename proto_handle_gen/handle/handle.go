//go:generate proto_handle_gen
package handle

import (
	"pslg/protos"
	"github.com/golang/protobuf/proto"
	"menet"
	"time"
)

// proto: 1

 // proto: 1
 // params: 
func pingPong(session *menet.TcpSession, ping *protos.Req_PING, rsp_ping *protos.Rsp_PING) int {
	rsp_ping.Pong = proto.Int64(ping.GetPing())
 	return 1
}

// proto: 2
func heartBeat(session *menet.TcpSession, _ *protos.Req_PING, resp *protos.Rsp_PING) {
	resp.Pong = proto.Int64(time.Now().Unix())
}
