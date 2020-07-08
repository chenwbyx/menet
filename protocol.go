package menet

import "log"

type ProtoFunc func(*TcpSession, *ProtobufMessage) []byte

var protoFuncs = make(map[uint16]ProtoFunc)

func Register(msgNo uint16, fn ProtoFunc) {
	if fn != nil {
		_, ok := protoFuncs[msgNo]
		if !ok {
			protoFuncs[msgNo] = fn
		} else {
			log.Println("duplicated msg no", msgNo)
		}
	}
}
