package merpc

import (
	"log"
	"net"
	"net/rpc"
)

const (
	rpcPath = "/_meRPC_"
)

type RpcServer struct {
	rpcSvr    *rpc.Server
	//httpSvr   *http.Server
	addr      string
	l         net.Listener
	exitFuncs []func()
}

func NewServer(addr string) *RpcServer {
	s := &RpcServer{}
	s.addr = addr
	s.rpcSvr = rpc.NewServer()
	//svrMux := http.NewServeMux()
	//svrMux.Handle(rpcPath, s.rpcSvr)
	//
	//s.httpSvr = &http.Server{Handler: svrMux}
	//func() {
	//	var err error
	//	err = httpSvr.Serve(l)
	//	if err == http.ErrServerClosed {
	//	}
	//}()
	//go httpSvr.Serve(l)

	return s
}

func (s *RpcServer) Register(r interface{}) error {
	return s.rpcSvr.Register(r)
}

func (s *RpcServer) Serve() error {
	var err error
	s.l, err = net.Listen("tcp", s.addr)
	if err != nil {
		log.Println("rpc listen error", err)
		return err
	}

	for {
		conn, err := s.l.Accept()
		if err != nil {
			log.Println("rpc accept", err)
			break
		}
		go s.rpcSvr.ServeConn(conn)
	}

	for _, fn := range s.exitFuncs {
		fn()
	}

	return nil //s.httpSvr.Serve(l)
}

func (s *RpcServer) Close() {
	s.l.Close()
}

func (s *RpcServer) AtClose(fn func()) {
	s.exitFuncs = append(s.exitFuncs, fn)
}
