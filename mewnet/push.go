package mewnet

import (
	"sync"
)

type WsSessionHub struct {
	sessionMap sync.Map
}

func (hub *WsSessionHub) Get(uid int32) (*WsSession, bool) {
	v, has := hub.sessionMap.Load(uid)
	if has {
		session, ok := v.(*WsSession)
		if ok {
			return session, true
		}
	}
	return nil, false
}

func (hub *WsSessionHub) AddSession(uid int32, s *WsSession) {
	hub.sessionMap.Store(uid, s)
}

func (hub *WsSessionHub) RemoveSession(uid int32) {
	hub.sessionMap.Delete(uid)
}

func (hub *WsSessionHub) PushMsg(uid int32, msgNo uint32, content []byte) {
	v, ok := hub.sessionMap.Load(uid)
	if ok {
		s := v.(*WsSession)
		s.PushMsg(msgNo, content)
	}
}

// 放入UID协程中处理，不保证一定执行
func (hub *WsSessionHub) PushWork(uid int32, fn func()) {
	v, ok := hub.sessionMap.Load(uid)
	if ok {
		s := v.(*WsSession)
		s.PushWork(fn)
	}
}

func (hub *WsSessionHub) Broadcast(msgNo uint32, content []byte) {
	var all []*WsSession
	hub.sessionMap.Range(func(key, value interface{}) bool {
		s := value.(*WsSession)
		all = append(all, s)
		return true
	})
	for _, s := range all {
		s.PushMsg(msgNo, content)
	}
}

/**
 * 这个函数及其特殊，一定不要在里面写 非线程安全的代码
 */
func (hub *WsSessionHub) DoSomething(fn func(uid int32, session *WsSession)) {
	hub.sessionMap.Range(func(key, value interface{}) bool {
		uid := key.(int32)
		session := value.(*WsSession)
		fn(uid, session)
		return true
	})
}
