package mewnet

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var gServer *WsServer

func GetServer() *WsServer {
	return gServer
}

type WsServer struct {
	httpServer           *http.Server
	protoFunctions       map[uint32]ProtoFunc
	sessions             *sync.Map //map[*WsSession]bool
	sessionNum           int32
	exitFunctions        []func()
	sessionExitFunctions []func(int32)
	wg                   sync.WaitGroup
	timeoutCloseRead     time.Duration
	certFile             string
	keyFile              string
	WsSessionHub
}

type WSHandler struct {
	server  *WsServer
	upgrade websocket.Upgrader
}

func (h *WSHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if conn, err := h.upgrade.Upgrade(resp, req, nil); err != nil {
		return
	} else {
		log.Println("incoming connection", req.RemoteAddr)
		h.server.wg.Add(1)
		wsSession := newWsSession(h.server, conn)
		h.server.sessions.Store(wsSession, true)
		atomic.AddInt32(&h.server.sessionNum, 1)
	}
}

func NewWsServer(addr string) *WsServer {
	server := new(WsServer)
	server.httpServer = &http.Server{
		Addr: addr,
		Handler: &WSHandler{
			upgrade: websocket.Upgrader{
				HandshakeTimeout: 10 * time.Second,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
			server: server,
		},
	}
	server.protoFunctions = make(map[uint32]ProtoFunc)
	server.sessions = new(sync.Map) // make(map[*WsSession]bool)

	gServer = server
	server.AtSessionClose(func(i int32) {
		atomic.AddInt32(&server.sessionNum, -1)
		server.wg.Done()
	})
	return server
}

func (s *WsServer) Register(msgNo uint32, fn ProtoFunc) {
	if fn != nil {
		_, ok := s.protoFunctions[msgNo]
		if !ok {
			s.protoFunctions[msgNo] = fn
		} else {
			log.Println("duplicated msg no", msgNo)
		}
	}
}

func (s *WsServer) Start() {
	var err error
	if s.certFile != "" && s.keyFile != "" {
		if err = s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile); err != nil {
			log.Printf("https server close, err = %v", err)
		}
	} else {
		if err = s.httpServer.ListenAndServe(); err != nil {
			log.Printf("http server close, err = %v", err)
		}
	}

	sessions := make([]*WsSession, 1)
	s.sessions.Range(func(k, v interface{}) bool {
		wsSession, ok := k.(*WsSession)
		if ok {
			sessions = append(sessions, wsSession)
		}
		return true
	})
	for _, wsSession := range sessions {
		if wsSession != nil {
			wsSession.close(true)
		}
	}

	for _, fn := range s.exitFunctions {
		fn()
	}
}

func (s *WsServer) SetTimeoutCloseRead(timeout time.Duration) {
	s.timeoutCloseRead = timeout
}
func (s *WsServer) SetCert(certFile, keyFile string) {
	s.certFile = certFile
	s.keyFile = keyFile
}

func (s *WsServer) SessionNum() int32 {
	return atomic.LoadInt32(&s.sessionNum)
}

func (s *WsServer) Stop() {
	err := s.httpServer.Shutdown(context.TODO())
	if err != nil {
		log.Println("server stop error ", err.Error())
	}
}

func (s *WsServer) AtClose(fn func()) {
	s.exitFunctions = append(s.exitFunctions, fn)
}

func (s *WsServer) AtSessionClose(fn func(int32)) {
	s.sessionExitFunctions = append(s.sessionExitFunctions, fn)
}
