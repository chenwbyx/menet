package merpc

import (
	"net/rpc"
	"log"
	"sync"
	"net"
	"time"
)

type RpcClient struct {
	c *rpc.Client
	addr string
	timeout time.Duration
	sync.Mutex
}

func NewClient(addr string) *RpcClient{
	client := &RpcClient{addr:addr,timeout:time.Second*3}
	client.getRpcClient()

	return client
}

func NewClientTimeout(addr string, timeout time.Duration) *RpcClient {
	client := &RpcClient{addr:addr,timeout:timeout}
	client.getRpcClient()

	return client
}

func (c *RpcClient)getRpcClient() error {
	c.Lock()
	defer c.Unlock()
	//var err error
	if c.c == nil {
		conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
		if err != nil {
			log.Println("rpc dialing:", err)
			return err
		}
		c.c = rpc.NewClient(conn)
		//c.c, err = rpc.DialHTTPPath("tcp", c.addr, rpcPath)
	}
	return nil
}

func (c *RpcClient)Call(serviceMethod string, args interface{}, reply interface{}) error {
	var err error
	err = c.getRpcClient()
	if err != nil {
		c.Close()
		return err
	}

	err = c.c.Call(serviceMethod, args, reply)
	if err != nil {
		log.Printf("rpc call [method:%s] error:%#v\n", serviceMethod, err)
		if err == rpc.ErrShutdown {
			c.Close()
			err = c.getRpcClient()
			if err != nil {
				c.Close()
				return err
			}
			return c.c.Call(serviceMethod, args, reply)
		}
	}
	return err
}

func (c *RpcClient)Close() {
	c.Lock()
	defer c.Unlock()
	if c.c != nil {
		c.c.Close()
		c.c = nil
	}
}
