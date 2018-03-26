package util

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ellcrys/gcoin/modules/types"
	net "github.com/libp2p/go-libp2p-net"
)

var (
	// WaitTimeBeforeRead is the amount of time to wait before reading a stream
	WaitTimeBeforeRead = 10 * time.Millisecond
)

// ReadHexStream reads and decodes the content of a stream if the context is hex encoded
func ReadHexStream(s net.Stream) ([]byte, error) {
	c, err := ReadStream(s)
	if err != nil {
		return nil, err
	}

	cStr := string(c)
	if !strings.HasPrefix(cStr, "0x") {
		return nil, fmt.Errorf("content is not hex encoded")
	}

	cBytes, err := hex.DecodeString(strings.TrimLeft(cStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid hex content in stream")
	}

	return cBytes, nil
}

// ReadStream reads bytes from a stream
func ReadStream(s net.Stream) ([]byte, error) {
	p := make([]byte, 4)
	var c []byte

	for {
		n, err := s.Read(p)
		if err != nil {
			if err == io.EOF {
				c = append(c, p[:n]...)
				break
			}
			return nil, fmt.Errorf("failed to read stream -> %s", err)
		}
		c = append(c, p[:n]...)
	}

	return c, nil
}

// ReadMessageFromStream reads a stream and attempts to decode it to a type.Message
func ReadMessageFromStream(s net.Stream) (*types.Message, error) {
	v, err := ReadHexStream(s)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream -> %s", err)
	}

	m, err := types.NewMessageFromJSON(v)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stream data -> %s", err)
	}

	return m, nil
}

// WaitThenReadStream reads the content of a stream after a period of time has passed
func WaitThenReadStream(waitTime time.Duration, s net.Stream, cb func(error, []byte)) {
	time.AfterFunc(waitTime, func() {
		bs, err := ReadStream(s)
		cb(err, bs)
	})
}

// WaitThenReadHexStream reads the content of a stream after a period of time has passed
func WaitThenReadHexStream(waitTime time.Duration, s net.Stream, cb func(error, []byte)) {
	time.AfterFunc(waitTime, func() {
		bs, err := ReadHexStream(s)
		cb(err, bs)
	})
}
