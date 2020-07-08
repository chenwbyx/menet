package mewnet

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"github.com/gorilla/websocket"
	"io"
	"log"
)

//const MaxPacketLen uint32 = 1024 * 1024

//type Message interface{}

//type MessageHandle interface {
//	Decode(io.Reader) (Message, error)
//	Encode(Message) ([]byte, error)
//}

const EZlibCompressMinimum = 1 * 1024 * 1024 //1024 / 4

type ProtoMessage struct {
	MsgNo    uint32 //协议号
	Compress uint16 //压缩方式  0: 不压缩   1: zlib
	Tag      uint16 //标志位校验位 PROTOTAG  或 JSONTAG 暂未使用
	Sid      uint32 //协议序列号ID 防止重放攻击
	Body     []byte //协议内容
}

//type ProtobufHandle struct {
//}
//解包
func Decode(conn *websocket.Conn) (*ProtoMessage, error) {
	mType, p, err := conn.ReadMessage()
	if err != nil {
		log.Printf("mType = %v p = %v err = %v conn = %v", mType, p, err, conn.LocalAddr())
		return nil, err
	}
	pm := new(ProtoMessage)
	//length = uint32(p[0]&0xff) + uint32(p[1]&0xff)<<8 + uint32(p[2]&0xff)<<16 + uint32(p[3]&0xff)<<24   //使用littleEndian的方式转换
	pm.MsgNo = uint32(p[4]&0xff) + uint32(p[5]&0xff)<<8 + uint32(p[6]&0xff)<<16 + uint32(p[7]&0xff)<<24   //使用littleEndian的方式转换
	pm.Compress = uint16(p[8]&0xff) + uint16(p[9]&0xff)<<8                                                //使用littleEndian的方式转换
	pm.Tag = uint16(p[10]&0xff) + uint16(p[11]&0xff)<<8                                                   //使用littleEndian的方式转换
	pm.Sid = uint32(p[12]&0xff) + uint32(p[13]&0xff)<<8 + uint32(p[14]&0xff)<<16 + uint32(p[15]&0xff)<<24 //使用littleEndian的方式转换
	if pm.Compress == 1 {
		var buf bytes.Buffer
		b := bytes.NewReader(p[16:])
		r, err := zlib.NewReader(b)
		if err != nil {
			log.Println("zlib.NewReader error", err.Error())
			return nil, err
		}
		_, err = io.Copy(&buf, r)
		if err != nil {
			log.Println("zlib io.Copy error", err.Error())
			return nil, err
		}
		pm.Body = buf.Bytes()
		err = r.Close()
		if err != nil {
			log.Println("zlib.NewReader.Close error", err.Error())
			return nil, err
		}
	} else {
		pm.Body = p[16:]
	}
	return pm, nil
}

//序列化包
func Encode(msg *ProtoMessage) []byte {
	var body []byte
	// 大数据开启压缩
	if len(msg.Body) > EZlibCompressMinimum && msg.Compress != 1 {
		msg.Compress = 1
	}
	if msg.Compress == 1 {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		_, err := w.Write(msg.Body)
		if err != nil {
			log.Println("zlib.NewWriter.Write error", err.Error())
			return nil
		}
		err = w.Close()
		if err != nil {
			log.Println("zlib.NewWriter.Close error", err.Error())
			return nil
		}
		body = buf.Bytes()
	} else {
		body = msg.Body
	}

	var l = len(body) //消息长度
	var buf [4]byte   //msgNo和Tag缓冲区
	var msgOut []byte
	binary.LittleEndian.PutUint32(buf[:], uint32(l)) //消息包长度 4字节
	msgOut = append(msgOut, buf[:]...)

	binary.LittleEndian.PutUint32(buf[:], msg.MsgNo) //协议号 4字节
	msgOut = append(msgOut, buf[:]...)

	binary.LittleEndian.PutUint16(buf[:2], msg.Compress) //压缩位 扩展为2字节
	msgOut = append(msgOut, buf[:2]...)

	binary.LittleEndian.PutUint16(buf[:2], msg.Tag) //校验位 扩展为2字节
	msgOut = append(msgOut, buf[:2]...)

	binary.LittleEndian.PutUint32(buf[:], msg.Sid) //协议序列号ID 防止重放攻击
	msgOut = append(msgOut, buf[:]...)

	msgOut = append(msgOut, body...)
	//msgOut = append(msgOut, msg.Body[:]...)
	//log.Println("### Encode msg", l, msg.MsgNo, msg.Tag, msg.Sid, body)
	return msgOut
}
