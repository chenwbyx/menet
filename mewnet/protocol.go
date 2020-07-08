package mewnet

type ProtoFunc func(*WsSession, *ProtoMessage) []byte

//var protoFunctions = make(map[uint16]ProtoFunc)

//func Register(msgNo uint16, fn ProtoFunc) {
//	if fn != nil {
//		_, ok := protoFunctions[msgNo]
//		if !ok {
//			protoFunctions[msgNo] = fn
//		} else {
//			log.Println("duplicated msg no", msgNo)
//		}
//	}
//}
