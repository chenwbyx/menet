package menet

import "sync"

var gSessionMap = new(sync.Map)

func Get(uid int32) (*TcpSession, bool) {
	v, has := gSessionMap.Load(uid)
	if has {
		session, ok := v.(*TcpSession)
		if ok {
			return session, true
		}
	}
	return nil, false
}

func Add(uid int32, s *TcpSession) {
	gSessionMap.Store(uid, s)
}

func Delete(uid int32) {
	gSessionMap.Delete(uid)
}

func PushToUser(uid int32, msgNo uint16, content []byte) {
	v, ok := gSessionMap.Load(uid)
	if ok {
		s := v.(*TcpSession)
		s.PushMsg(msgNo, content)
	}
}

func PushToAll(msgNo uint16, content []byte) {
	var all []*TcpSession
	gSessionMap.Range(func(key, value interface{}) bool {
		s := value.(*TcpSession)
		all = append(all, s)
		return true
	})
	for _, s := range all {
		s.PushMsg(msgNo, content)
	}
}
