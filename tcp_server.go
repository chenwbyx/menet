package menet

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type TcpServer struct {
	l          net.Listener
	protoFuncs map[uint16]ProtoFunc
	sessions   *sync.Map//map[*TcpSession]bool
	sessionNum int32
	stopSignal bool
	exitFuncs  []func()
	sessionExitFuncs []func(int32)
}

func NewTcpServer(l net.Listener) *TcpServer {
	server := new(TcpServer)
	server.l = l
	server.protoFuncs = make(map[uint16]ProtoFunc)
	server.sessions = new(sync.Map) // make(map[*TcpSession]bool)
	server.stopSignal = false
	return server
}

func (s *TcpServer) Register(msgNo uint16, fn ProtoFunc) {
	if fn != nil {
		_, ok := s.protoFuncs[msgNo]
		if !ok {
			s.protoFuncs[msgNo] = fn
		} else {
			log.Println("duplicated msg no", msgNo)
		}
	}
}

func (s *TcpServer) Start() {
	for {
		if s.stopSignal {
			break
		}
		conn, err := s.l.Accept()
		if err != nil {
			log.Println("error: ", err)
			continue
		}
		log.Println("incoming connection", conn.RemoteAddr())

		tcpSession := newTcpSession(conn)
		tcpSession.server = s
		s.sessions.Store(tcpSession, true)
		atomic.AddInt32(&s.sessionNum, 1)
		//s.sessions[tcpSession] = true
	}


	sessions := make([]*TcpSession, 1)
	s.sessions.Range(func(k, v interface{}) bool {
		tcpSession, ok := k.(*TcpSession)
		if ok {
			sessions = append(sessions, tcpSession)
		}
		return true
	})
	for _, tcpSession := range sessions {
		if tcpSession != nil {
			tcpSession.Close()
		}
	}

	for _, fn := range s.exitFuncs {
		fn()
	}
	//for tcpSession := range s.sessions {
	//	tcpSession.Close()
	//}
}

func (s *TcpServer) SessionNum() int32 {
	return atomic.LoadInt32(&s.sessionNum)
}

func (s *TcpServer) Stop() {
	s.stopSignal = true
	s.l.Close()
}

func (s *TcpServer) AtClose(fn func()) {
	s.exitFuncs = append(s.exitFuncs, fn)
}

func (s *TcpServer) AtSessionClose(fn func(int32)) {
	s.sessionExitFuncs = append(s.sessionExitFuncs, fn)
}
