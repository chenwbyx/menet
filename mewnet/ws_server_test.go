package mewnet

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

/**C-request */
type Req_Ping struct {
}

/*S-response*/
type Resp_PING struct {
	Time int64
}

func init() {
	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate | log.Lmicroseconds)
}

func TestPushMsgConcurrence(t *testing.T) {
	wg := &sync.WaitGroup{}
	loopNum := 10

	clientPingHandle := func(no uint32, req interface{}) {
		log.Printf("clientPingHandle--->outMsg--->%v", req)
		wg.Done()
	}

	serverPingHandle := func(w *WsSession, p *ProtoMessage) (outMsg []byte) {
		w.SetUid(1)
		GetServer().AddSession(1, w)
		resp := &Resp_PING{}
		outMsg, _ = json.Marshal(resp)
		log.Printf("serverPingHandle--->outMsg--->%v", string(outMsg))
		resp.Time = time.Now().Unix()
		for i := 0; i < loopNum; i++ {
			wg.Add(1)
			go GetServer().PushMsg(1, 100, outMsg)
		}
		return
	}

	defer func() {
		time.Sleep(time.Second * 1)
	}()

	addr := fmt.Sprintf(":%v", 6617)
	server := NewWsServer(addr)
	server.Register(100, serverPingHandle)
	go func() {
		defer server.Stop()
		server.Start()
	}()

	req := Req_Ping{}
	resp := Resp_PING{}
	client := NewWsClient(addr)
	client.Handle(100, 0, clientPingHandle)
	wg.Add(1)
	client.Request(100, &req, &resp)
	defer client.Close()
	wg.Wait()

}

func TestServerTimer(t *testing.T) {
	uid := int32(1)
	wg := &sync.WaitGroup{}

	loopNum := 1000
	uidMap := map[int]int{}
	clientPingHandle := func(no uint32, req interface{}) {
		//log.Printf("clientPingHandle--->outMsg--->%v", req)
		wg.Done()
	}
	serverPingHandle := func(w *WsSession, p *ProtoMessage) (outMsg []byte) {
		w.SetUid(uid)
		GetServer().AddSession(uid, w)
		resp := &Resp_PING{}
		outMsg, _ = json.Marshal(resp)
		//log.Printf("serverPingHandle--->outMsg--->%v", string(outMsg))
		resp.Time = time.Now().Unix()
		for i := 0; i < loopNum; i++ {
			uidMap[i] = 1
		}
		for i := 0; i < loopNum; i++ {
			_ = uidMap[i]
		}
		return
	}
	timerStop := make(chan bool)

	timerFunc := func() {
		for i := 0; i < loopNum; i++ {
			uidMap[i] = i + i
		}
	}

	defer func() {
		time.Sleep(time.Second * 1)
	}()

	addr := fmt.Sprintf(":%v", 6617)
	server := NewWsServer(addr)
	server.Register(100, serverPingHandle)

	go func() {
		defer server.Stop()
		server.Start()
	}()

	req := Req_Ping{}
	resp := Resp_PING{}
	client := NewWsClient(addr)
	client.Handle(100, 0, clientPingHandle)

	wg.Add(1)
	client.Request(100, &req, &resp)

	go func() {
		log.Println("ticker handler begin")
		for {
			select {
			case <-timerStop:
				break
			default:
				//timerFunc()
				server.PushWork(uid, timerFunc)
			}
			time.Sleep(1)
		}
	}()

	for i := 0; i < loopNum; i++ {
		wg.Add(1)
		client.Request(100, &req, &resp)
	}

	defer client.Close()
	wg.Wait()
	timerStop <- true

}
