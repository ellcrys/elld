package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

var (
	// OpHandshake represents a handshake ops code
	OpHandshake OpCode = 0x1
)

// OpCode represents an op code
type OpCode uint8

// Message represents action and data instructions to be sent to a remote peer
type Message struct {
	Op  OpCode `json:"op"`
	Msg []byte `json:"msg"`
}

// NewMessage creates a new message
func NewMessage(op OpCode, m []byte) *Message {
	return &Message{Op: op, Msg: m}
}

// NewMessageFromJSON creates a new message from an encoded hex string created by Hex()
func NewMessageFromJSON(d []byte) (*Message, error) {
	var m Message
	if err := json.Unmarshal(d, &m); err != nil {
		return nil, fmt.Errorf("failed to parse message. Malformed message -> %s", err)
	}
	return &m, nil
}

// Hex converts the message to json and then to hex with `0x` prefix
func (m *Message) Hex() []byte {
	b, _ := json.Marshal(m)
	return []byte(fmt.Sprintf("0x%s", hex.EncodeToString(b)))
}

// HexString is the same as ToHex except it returns string
func (m *Message) HexString() string {
	b, _ := json.Marshal(m)
	return fmt.Sprintf("0x%s", hex.EncodeToString(b))
}

// Bytes returns json encoded representation of instance as raw bytes
func (m *Message) Bytes() []byte {
	b, _ := json.Marshal(m)
	return b
}

// Scan copies the message into a struct or map
func (m *Message) Scan(dest interface{}) error {
	return json.Unmarshal(m.Msg, &dest)
}

// HandshakeMsg represents a handshake message
type HandshakeMsg struct {
	ID string
}
