package proxy

import (
	"encoding/binary"
	"fmt"
)

// 消息类型
const (
	MsgTypeAuth         uint16 = 0x0001
	MsgTypeAuthResponse uint16 = 0x0002
	MsgTypeHeartbeat    uint16 = 0x0003
	MsgTypeHeartbeatAck uint16 = 0x0004

	MsgTypeTunnelCreate uint16 = 0x0010
	MsgTypeTunnelDelete uint16 = 0x0011
	MsgTypeTunnelStart  uint16 = 0x0012
	MsgTypeTunnelStop   uint16 = 0x0013
	MsgTypeTunnelStatus uint16 = 0x0014
	MsgTypeTunnelError  uint16 = 0x001F

	MsgTypeDataOpen    uint16 = 0x0020
	MsgTypeDataClose   uint16 = 0x0021
	MsgTypeDataTransfer uint16 = 0x0022

	MsgTypeConfigPush uint16 = 0x0030
	MsgTypeConfigPull uint16 = 0x0031
)

// 消息标志
const (
	MsgFlagNone      uint16 = 0x0000
	MsgFlagCompressed uint16 = 0x0001
	MsgFlagEncrypted  uint16 = 0x0002
)

// Message 消息结构
type Message struct {
	Type    uint16
	Flag    uint16
	MsgID   uint32
	Payload []byte
}

// 消息头大小: Type(2) + Flag(2) + MsgID(4) = 8字节
const HeaderSize = 8

// Encode 编码消息
func (m *Message) Encode() []byte {
	data := make([]byte, HeaderSize+len(m.Payload))
	binary.BigEndian.PutUint16(data[0:2], m.Type)
	binary.BigEndian.PutUint16(data[2:4], m.Flag)
	binary.BigEndian.PutUint32(data[4:8], m.MsgID)
	copy(data[HeaderSize:], m.Payload)
	return data
}

// DecodeMessage 解码消息
func DecodeMessage(data []byte) (*Message, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("消息太短")
	}

	msg := &Message{
		Type:  binary.BigEndian.Uint16(data[0:2]),
		Flag:  binary.BigEndian.Uint16(data[2:4]),
		MsgID: binary.BigEndian.Uint32(data[4:8]),
	}
	if len(data) > HeaderSize {
		msg.Payload = data[HeaderSize:]
	}

	return msg, nil
}

// AuthPayload 认证载荷
type AuthPayload struct {
	Token      string `json:"token"`
	DeviceID   string `json:"device_id"`
	ClientVersion string `json:"client_version"`
	Platform   string `json:"platform"`
}

// AuthResponsePayload 认证响应载荷
type AuthResponsePayload struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Heartbeat int      `json:"heartbeat_interval"`
	Tunnels   []TunnelInfo `json:"tunnels"`
}

// TunnelInfo 隧道信息
type TunnelInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	LocalHost  string `json:"local_host"`
	LocalPort  int    `json:"local_port"`
	RemotePort int    `json:"remote_port"`
}

// DataOpenPayload 数据通道打开载荷
type DataOpenPayload struct {
	TunnelID     string `json:"tunnel_id"`
	ConnectionID string `json:"connection_id"`
	Protocol     string `json:"protocol"`
	SourceAddr   string `json:"source_addr"`
}
