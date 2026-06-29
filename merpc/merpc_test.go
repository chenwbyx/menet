//go:build integration

package merpc

import (
	"testing"
	"fmt"
	"time"
	"menet/merpc/rpc_testdata"
)

//go:generate genrpc

//rpc: Rpc
type Rpc struct{}

func (*Rpc)RpcTest(req *rpc_testdata.Req, resp *rpc_testdata.Resp) error {
	fmt.Printf("%#v\n", req)
	fmt.Printf("%#v\n", req.PtrSlice)
	fmt.Printf("%#v\n", *req.Nest.PtrEmbed)
	resp.PtrInt = nil
	resp.PtrEmbed = nil
	resp.T = make(map[int32]*rpc_testdata.Embed)
	resp.T[1] = nil

	return nil
}

func pInt32(i int32) *int32 {
	return &i
}

func pInt64(i int64) *int64 {
	return &i
}

func pString(s string) *string {
	return &s
}

func TestRpcClientTimeout(t *testing.T) {
	client := NewClient("172.168.11.96:2300")
	client.Close()
}

func TestRpc(t *testing.T) {
	server := NewServer("127.0.0.1:9999")
	client := NewClient("127.0.0.1:9999")
	if client == nil {
		t.Error("connect fail")
		return
	}
	server.Register(&Rpc{})
	go func() {
		server.Serve()
	} ()
	time.Sleep(time.Millisecond)

	err := client.Call("Rpc.NoThisMethod", &rpc_testdata.Resp{}, &rpc_testdata.Resp{})
	if err == nil {
		t.Error("expect an error: no metherd")
		return
	}

	resp := &rpc_testdata.Resp{}
	req := &rpc_testdata.Req{Uid:0,
		Id:1,
		State:1,
		State1:1,
		Value:1,
		String:"1",
		ValueMap:map[int32]bool{1:true},
		ValueArray: [4]int32{1,1,1,1},
		ValueSliceList: [][2]int32{{1,1},{1,1}},
		//MapMap     map[int32]map[int32]bool
		//MapList    map[int32][4]int32
		//MapSList   map[int32][]int32
		//SListMap   []map[int32]bool
		//SListList  [][4]int32
		//SListSList [][]int32
		//ListMap    [4]map[int32]bool
		//ListList   [4][4]int32
		//ListSList  [4][]int32
		PtrInt32: nil,
		PtrInt64: pInt64(0),
		PtrStr: pString("1"),
		PtrSlice: &[]int32{},
		Nest: &rpc_testdata.Embed{PtrEmbed:&rpc_testdata.Embed{}},
	}
	fmt.Println(pInt32(0))
	err = client.Call("Rpc.RpcTest", req, resp)
	if err != nil {
		t.Error("rpc call fail")
		return
	}
	fmt.Println(resp)

	client.Close()
	server.Close()
}
