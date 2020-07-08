package mewnet

import (
	"log"
	"menet/util"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	sessionRunning = 0
	sessionStop    = 1
)

type WsSession struct {
	server     *WsServer
	conn       *websocket.Conn
	uid        int32
	writeQueue *MessageQueue
	state      int32
	postMsg    *ProtoMessage
	workQueue  *WorkQueue
	wg         sync.WaitGroup
}

func newWsSession(server *WsServer, conn *websocket.Conn) *WsSession {
	wsSession := new(WsSession)
	wsSession.server = server
	wsSession.conn = conn
	wsSession.writeQueue = NewMessageQueue()
	wsSession.workQueue = NewWorkQueue()
	wsSession.wg.Add(2)
	go func() {
		// write thread
		defer func() {
			wsSession.wg.Done()
			wsSession.Close()
			log.Println("send thread exit", wsSession.conn.LocalAddr())
		}()
		var writeList [][]byte
		var err error
	LabelWriteThread:
		for {
			writeList = writeList[0:0]
			exit := wsSession.writeQueue.Pick(&writeList)
			// 遍历要发送的数据
			for _, msg := range writeList {
				err = wsSession.conn.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					break LabelWriteThread
				}
			}
			if exit {
				break
			}
		}
	}()
	go func() {
		// read thread
		defer func() {
			wsSession.wg.Done()
			wsSession.Close()
			log.Println("recv thread exit")
		}()
		var err error
		for {
			err = nil
			if atomic.LoadInt32(&wsSession.state) == sessionStop {
				log.Println("recv thread session close")
				break
			}
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
			var protoMsg *ProtoMessage
			protoMsg, err = Decode(conn)
			if err != nil {
				log.Println("recv thread", err)
				break
			}
			fn, ok := wsSession.server.protoFunctions[protoMsg.MsgNo]
			if ok {
				// 处理接受消息
				func() {
					defer func() {
						if e := recover(); e != nil {
							log.Println(e)
							log.Printf("stack: %s", debug.Stack())
						} else {
						}
					}()
					outMsg := &ProtoMessage{
						MsgNo:    protoMsg.MsgNo,
						Compress: 0,
						Tag:      0,
						Sid:      protoMsg.Sid,
						Body:     fn(wsSession, protoMsg),
					}
					msgStr := Encode(outMsg)
					if err != nil {
						log.Println("proto handle", err)
					} else {
						wsSession.sendMsg(msgStr)
					}
					if wsSession.postMsg != nil {
						msgStr := Encode(wsSession.postMsg)
						if err != nil {
							log.Println("post msg", err)
						} else {
							wsSession.sendMsg(msgStr)
						}
						wsSession.postMsg = nil
					}
				}()
			} else {
				// 未知包
				break
			}
			// 处理其他工作
			funcList := wsSession.workQueue.Dump()
			for _, work := range funcList {
				work()
			}
		}
	}()
	return wsSession
}

func (s *WsSession) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

func (s *WsSession) SetUid(uid int32) {
	s.uid = uid
}

func (s *WsSession) GetUid() int32 {
	return s.uid
}

func (s *WsSession) sendMsg(buf []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("%#v\n", s)
		}
	}()
	if atomic.LoadInt32(&s.state) == sessionRunning {
		s.writeQueue.Add(buf)
	} else {
	}
}

func (s *WsSession) PushMsg(msgNo uint32, content []byte) {
	pm := &ProtoMessage{
		MsgNo:    msgNo,
		Compress: 0,
		Tag:      0,
		Sid:      0,
		Body:     content,
	}
	msgStr := Encode(pm)
	s.sendMsg(msgStr)
}

func (s *WsSession) PushWork(fn func()) {
	if atomic.LoadInt32(&s.state) == sessionRunning {
		s.workQueue.Add(util.RecoverWrapFunc(fn))
	}
}

func (s *WsSession) Post(msgNo uint32, content []byte) {
	s.postMsg = &ProtoMessage{
		MsgNo:    msgNo,
		Compress: 0,
		Tag:      0,
		Sid:      0,
		Body:     content,
	}
}

func (s *WsSession) close(wait bool) {
	if atomic.CompareAndSwapInt32(&s.state, sessionRunning, sessionStop) {
		s.writeQueue.Add(nil)
		s.server.sessions.Delete(s)
		if s.uid != 0 {
			s.server.RemoveSession(s.uid)
		}
		_ = s.conn.SetReadDeadline(time.Now().Add(s.server.timeoutCloseRead))
		if wait {
			s.wg.Wait()
			s.workQueue.Reset()
			s.writeQueue.Reset()
			for _, fn := range s.server.sessionExitFunctions {
				fn(s.GetUid())
			}
			_ = s.conn.Close()
		} else {
			go func() {
				s.wg.Wait()
				s.workQueue.Reset()
				s.writeQueue.Reset()
				for _, fn := range s.server.sessionExitFunctions {
					fn(s.GetUid())
				}
				_ = s.conn.Close()
			}()
		}
	} else {
	}
}

func (s *WsSession) Close() {
	s.close(true)
}
