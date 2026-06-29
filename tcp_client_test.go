//go:build integration

package menet

import (
	"testing"
	"pslg/protos"
	"github.com/golang/protobuf/proto"
	"time"
	"sync"
)

func TestPingPong(t *testing.T) {
	client := NewTcpClient("172.168.11.116:2333")
	if client == nil {
		t.Error("connect fail")
		return
	}
	wait := new(sync.WaitGroup)
	fn := func() {
		pong := &protos.Rsp_PING{}
		client.Request(100, &protos.Req_PING{
			Ping: proto.Int64(time.Now().Unix()),
		}, pong)
		t.Log(pong)
		wait.Done()
	}

	wait.Add(1)
	go fn()
	wait.Add(1)
	go fn()

	wait.Wait()
	client.Close()
}

func TestPingPongHandle(t *testing.T) {
	client := NewTcpClient("172.168.11.116:2333")
	if client == nil {
		t.Error("connect fail")
		return
	}
	done := make(chan bool)
	client.Handle(100, &protos.Rsp_PING{}, func(msgNo uint16, pb proto.Message) {
		resp, ok := pb.(*protos.Rsp_PING)
		if !ok {
			t.Fail()
		}
		t.Log(resp)
		done<-true
	})
	client.AsyncRequest(100, &protos.Req_PING{
		Ping: proto.Int64(time.Now().Unix()),
	})
	<-done
	client.Close()
}
