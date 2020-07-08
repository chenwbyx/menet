package mewnet

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type msgHandle struct {
	f  func(uint32, interface{})
	pb interface{}
}

const (
	clientRunning = 0
	clientStop    = 1
)

type WsClient struct {
	conn       *websocket.Conn
	msgRecv    *sync.Map // map[uint16]chan[]byte
	msgHandles map[uint32]msgHandle
	writeChan  chan []byte
	stop       int32
}

func NewWsClient(addr string) *WsClient {
	var err error
	client := new(WsClient)
	u := url.URL{Scheme: "ws", Host: addr}
	log.Println(u.String(), addr)
	client.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("dial fail:", addr)
		return nil
	}
	//client.conn.PayloadType = websocket.BinaryFrame
	client.msgRecv = new(sync.Map) // make(map[uint16]chan[]byte)
	client.msgHandles = make(map[uint32]msgHandle)
	client.writeChan = make(chan []byte)
	client.stop = clientRunning
	go func() {
		for {
			var msg []byte
			var ok bool
			select {
			case msg, ok = <-client.writeChan:
				if ok {
					//client.conn.Write(msg)
					client.conn.WriteMessage(websocket.BinaryMessage, msg)
				}
			}
			if ok == false {
				break
			}
		}
		log.Println("send thread exit")
	}()
	go func() {
		//decoder := &ProtoMessage{}
		var err error
		for {
			err = nil
			if atomic.LoadInt32(&client.stop) == clientStop {
				break
			}
			var msg *ProtoMessage
			//msg, err = decoder.Decode(client.conn)
			msg, err = Decode(client.conn)
			if err != nil {
				break
			}
			//pbMsg := msg.(*ProtoMessage)
			pbMsg := msg
			v, ok := client.msgRecv.Load(pbMsg.MsgNo) // client.msgRecv[pbMsg.MsgNo]
			if ok {
				client.msgRecv.Delete(pbMsg.MsgNo)
				go func() {
					ch, _ := v.(chan []byte)
					ch <- pbMsg.Body
					close(ch)
				}()
			}
			h, ok := client.msgHandles[pbMsg.MsgNo]
			if ok {
				pb := reflect.ValueOf(h.pb).Interface().(interface{})
				err := json.Unmarshal(pbMsg.Body, &pb)
				if err == nil {
					go h.f(pbMsg.MsgNo, pb)
				}
			}
		}
		client.Close()
		log.Println("recv thread exit")
	}()
	return client
}

func (c *WsClient) run() {

}

func (c *WsClient) SendMsg(msgNo uint32, content []byte) {
	msg := &ProtoMessage{MsgNo: msgNo, Compress: 0, Tag: 0, Sid: 0, Body: content}
	//encoder := &ProtoMessage{}
	//request, _ := encoder.Encode(msg)
	request := Encode(msg)
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			log.Printf("stack: %s", debug.Stack())
		}
	}()
	c.writeChan <- request // c.conn.Write(request)
}

func (c *WsClient) Request(msgNo uint32, req interface{}, resp interface{}) error {
	content, err := json.Marshal(req)
	ch := make(chan []byte)
	_, has := c.msgRecv.LoadOrStore(msgNo, ch)
	if has {
		return errors.New("has a pending request")
	}
	if err != nil {
		return err
	}
	// c.msgRecv.Store(msgNo, ch) // c.msgRecv[msgNo] = ch
	c.SendMsg(msgNo, content)
	timeout := time.NewTimer(time.Second * 5)
	defer timeout.Stop()
	select {
	case recvMsg := <-ch:
		return json.Unmarshal(recvMsg, resp)
	case <-timeout.C:
		return errors.New("request timeout")
	}
}

func (c *WsClient) Handle(msgNo uint32, msgType interface{}, f func(uint32, interface{})) {
	c.msgHandles[msgNo] = msgHandle{f: f, pb: msgType}
}

func (c *WsClient) AsyncRequest(msgNo uint32, req interface{}) {
	content, _ := json.Marshal(req)
	c.SendMsg(msgNo, content)
}

func (c *WsClient) Close() {
	if atomic.CompareAndSwapInt32(&c.stop, clientRunning, clientStop) {
		close(c.writeChan)
		c.conn.Close()
	}
}

func (c *WsClient) Dead() bool {
	return atomic.LoadInt32(&c.stop) == clientStop
}
