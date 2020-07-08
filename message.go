package menet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

const MaxPacketLen uint32 = 1024 * 1024

type Message interface{}

type MessageHandle interface {
	Decode(io.Reader) (Message, error)
	Encode(Message) ([]byte, error)
}

type ProtobufMessage struct {
	MsgNo uint16
	Body  []byte
}

type ProtobufHandle struct {
}

func (ph *ProtobufHandle) Decode(r io.Reader) (Message, error) {
	var buf [8]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return nil, err
	}
	l := *(*uint32)((unsafe.Pointer)(&buf[0]))
	if l > MaxPacketLen {
		return nil, errors.New(fmt.Sprintf("large packet %d", l))
	}
	body := make([]byte, l)
	_, err = io.ReadFull(r, body)
	if err != nil {
		return nil, err
	}
	pm := new(ProtobufMessage)
	pm.MsgNo = *(*uint16)((unsafe.Pointer)(&buf[4]))
	pm.Body = body

	//log.Println("recv msg len", l, "id", pm.MsgNo, "content", body)

	return pm, nil
}

func (ph *ProtobufHandle) Encode(msg Message) ([]byte, error) {
	protoMsg, ok := msg.(*ProtobufMessage)
	if !ok {
		return nil, errors.New("type error")
	}
	l := len(protoMsg.Body)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(l))
	var msgBuf []byte
	msgBuf = append(msgBuf, buf[:]...)
	binary.LittleEndian.PutUint32(buf[:], uint32(protoMsg.MsgNo))
	msgBuf = append(msgBuf, buf[:]...)
	msgBuf = append(msgBuf, protoMsg.Body[:]...)
	return msgBuf, nil
}
