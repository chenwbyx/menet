package menet

import (
	"log"
	"net"
	"runtime/debug"
	"sync/atomic"
	"time"
)

const (
	sessionRunning = 0
	sessionStop    = 1
)

type TcpSession struct {
	server    *TcpServer
	conn      net.Conn
	uid       int32
	writeChan chan []byte
	stop      chan bool
	v         int32
	postMsg   *ProtobufMessage
}

func newTcpSession(conn net.Conn) *TcpSession {
	tcpSession := new(TcpSession)
	tcpSession.conn = conn
	tcpSession.writeChan = make(chan []byte)
	tcpSession.stop = make(chan bool)
	ph := new(ProtobufHandle)
	go func() {
		// write thread
		for {
			select {
			case msg := <-tcpSession.writeChan:
				tcpSession.conn.Write(msg)
			case <-tcpSession.stop:
				goto write_exit
			}
		}
	write_exit:
		close(tcpSession.writeChan)
		log.Println("send thread exit", tcpSession.conn)
	}()
	go func() {
		// read thread
		defer conn.Close()
		var err error
		for {
			err = nil
			if atomic.LoadInt32(&tcpSession.v) == sessionStop {
				log.Println("recv thread exit")
				break
			}
			conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
			var msg Message
			msg, err = ph.Decode(conn)
			if err != nil {
				log.Println("recv thread", err)
				break
			}
			protoMsg, ok := msg.(*ProtobufMessage)
			if !ok {
				log.Println("recv thread proto error")
				break
			}
			fn, ok := tcpSession.server.protoFuncs[protoMsg.MsgNo]
			if ok {
				content := func() []byte {
					defer func() {
						if e := recover(); e != nil {
							log.Println(e)
							log.Printf("stack: %s", debug.Stack())
						}
					}()
					return fn(tcpSession, protoMsg)
				}()
				outMsg := &ProtobufMessage{
					MsgNo: protoMsg.MsgNo,
					Body:  content,
				}
				msgStr, err := ph.Encode(outMsg)
				if err != nil {
					log.Println("proto handle", err)
				} else {
					tcpSession.sendMsg(msgStr)
				}
				if tcpSession.postMsg != nil {
					msgStr, err := ph.Encode(tcpSession.postMsg)
					if err != nil {
						log.Println("post msg", err)
					} else {
						tcpSession.sendMsg(msgStr)
					}
					tcpSession.postMsg = nil
				}
			}
		}
		if err != nil {
			tcpSession.Close()
			log.Println("send stop to write thread")
		}
		log.Println("recv thread exit")
	}()
	return tcpSession
}

func (s *TcpSession) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

func (s *TcpSession) SetUid(uid int32) {
	s.uid = uid
}

func (s *TcpSession) GetUid() int32 {
	return s.uid
}

func (s *TcpSession) sendMsg(buf []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("%#v\n", s)
		}
	} ()
	if atomic.LoadInt32(&s.v) == sessionRunning {
		s.writeChan <- buf
	}
}

func (s *TcpSession) PushMsg(msgNo uint16, content []byte) {
	ph := &ProtobufHandle{}
	pm := &ProtobufMessage{
		MsgNo: msgNo,
		Body:  content,
	}
	msgStr, err := ph.Encode(pm)
	if err != nil {
		return
	}
	//if atomic.LoadInt32(&s.v) == sessionRunning {
	//	s.writeChan <- msgStr
	//}
	s.sendMsg(msgStr)
}

func (s *TcpSession) Post(msgNo uint16, content []byte) {
	s.postMsg = &ProtobufMessage{
		MsgNo: msgNo,
		Body:  content,
	}
}

func (s *TcpSession) Close() {
	if atomic.CompareAndSwapInt32(&s.v, sessionRunning, sessionStop) {
		for _, fn := range s.server.sessionExitFuncs {
			fn(s.GetUid())
		}
		s.stop <- true
		s.server.sessions.Delete(s)
		atomic.AddInt32(&s.server.sessionNum, -1)
		if s.uid != 0 {
			Delete(s.uid)
		}
		s.conn.Close()
	} else {
	}
}
func (s *TcpSession) Stop() {
	s.stop <- true
}
