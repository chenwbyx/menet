package menet

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"log"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type msgHandle struct {
	f  func(uint16, proto.Message)
	pb proto.Message
}

const (
	clientRunning = 0
	clientStop    = 1
)

type TcpClient struct {
	conn       net.Conn
	msgRecv    *sync.Map // map[uint16]chan[]byte
	msgHandles map[uint16]msgHandle
	writeChan  chan []byte
	stop       int32
	coder      MessageHandle
}

func NewTcpClient(addr string) *TcpClient {
	return NewTcpClientWittCoder(addr, &ProtobufHandle{})
}

func NewTcpClientWittCoder(addr string, coder MessageHandle) *TcpClient {
	var err error
	client := new(TcpClient)
	client.conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.Println("dial fail:", addr)
		return nil
	}
	client.msgRecv = new(sync.Map) // make(map[uint16]chan[]byte)
	client.msgHandles = make(map[uint16]msgHandle)
	client.writeChan = make(chan []byte)
	client.stop = clientRunning
	client.coder = coder
	go func() {
		for {
			var msg []byte
			var ok bool
			select {
			case msg, ok = <-client.writeChan:
				if ok {
					client.conn.Write(msg)
				}
			}
			if ok == false {
				break
			}
		}
		log.Println("send thread exit")
	}()
	go func() {
		//decoder := &ProtobufHandle{}
		var err error
		for {
			err = nil
			if atomic.LoadInt32(&client.stop) == clientStop {
				break
			}
			var msg Message
			msg, err = client.coder.Decode(client.conn)
			if err != nil {
				break
			}
			pbMsg := msg.(*ProtobufMessage)
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
				pb := reflect.ValueOf(h.pb).Interface().(proto.Message)
				err := proto.Unmarshal(pbMsg.Body, pb)
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

func (c *TcpClient) run() {

}

func (c *TcpClient) SendMsg(msgNo uint16, content []byte) {
	msg := &ProtobufMessage{MsgNo: msgNo, Body: content}
	//encoder := &ProtobufHandle{}
	request, _ := c.coder.Encode(msg)
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			log.Printf("stack: %s", debug.Stack())
		}
	}()
	c.writeChan <- request // c.conn.Write(request)
}

func (c *TcpClient) Request(msgNo uint16, req proto.Message, resp proto.Message) error {
	content, _ := proto.Marshal(req)
	ch := make(chan []byte)
	_, has := c.msgRecv.LoadOrStore(msgNo, ch)
	if has {
		return errors.New("has a pending request")
	}
	// c.msgRecv.Store(msgNo, ch) // c.msgRecv[msgNo] = ch
	c.SendMsg(msgNo, content)
	timeout := time.NewTimer(time.Second * 5)
	defer timeout.Stop()
	select {
	case recvMsg := <-ch:
		return proto.Unmarshal(recvMsg, resp)
	case <-timeout.C:
		return errors.New("request timeout")
	}
}

func (c *TcpClient) Handle(msgNo uint16, msgType proto.Message, f func(uint16, proto.Message)) {
	c.msgHandles[msgNo] = msgHandle{f: f, pb: msgType}
}

func (c *TcpClient) AsyncRequest(msgNo uint16, req proto.Message) {
	content, _ := proto.Marshal(req)
	c.SendMsg(msgNo, content)
}

func (c *TcpClient) Close() {
	if atomic.CompareAndSwapInt32(&c.stop, clientRunning, clientStop) {
		close(c.writeChan)
		c.conn.Close()
	}
}

func (c *TcpClient) Dead() bool {
	return atomic.LoadInt32(&c.stop) == clientStop
}
