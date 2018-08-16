package messages

// Handshake represents the first message between peers
type Handshake struct {
	SubVersion string `json:"subversion" msgpack:"subversion"`
}

// GetAddr is used to request for peer addresses from other peers
type GetAddr struct {
}

// Addr is used to send peer addresses in response to a GetAddr
type Addr struct {
	Addresses []*Address `json:"addresses" msgpack:"addresses"`
}

// Address represents a peer address
type Address struct {
	Address   string `json:"address" msgpack:"address"`
	Timestamp int64  `json:"timestamp" msgpack:"timestamp"`
}

// Ping represents a ping message
type Ping struct {
}

// Pong represents a pong message
type Pong struct {
}

// Reject defines information about a rejected action
type Reject struct {
	Message   string `json:"message" msgpack:"message"`
	Code      int32  `json:"code" msgpack:"code"`
	Reason    string `json:"reason" msgpack:"reason"`
	ExtraData []byte `json:"extraData" msgpack:"extraData"`
}

// RequestBlock represents a message requesting for a block
type RequestBlock struct {
	Hash   string `json:"hash" msgpack:"hash"`
	Number uint64 `json:"number" msgpack:"number"`
}
