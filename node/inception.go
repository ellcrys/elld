package node

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/constants"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	ic "github.com/libp2p/go-libp2p-crypto"
	net "github.com/libp2p/go-libp2p-net"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// Inception represents the peer protocol
type Inception struct {
	mtx                         *sync.Mutex         // main mutex
	arm                         *sync.Mutex         // address relay mutex
	version                     string              // the protocol version
	peer                        *Node               // the local peer
	log                         logger.Logger       // the logger
	lastRelayPeersSelectionTime time.Time           // the time the last addr msg relay peers where selected
	addrRelayPeers              [2]*Node            // peers to relay addr msgs to
	unsignedTxPool              *txpool.TxPool      // the transaction pool for unsigned transactions
	openTxSessions              map[string]struct{} // Holds the id of transactions awaiting endorsement. Protected by mtx.
	txsRelayQueue               *txpool.TxQueue     // stores transactions waiting to be relayed
}

// NewInception creates a new instance of the protocol codenamed "Inception"
func NewInception(p *Node, log logger.Logger) *Inception {
	return &Inception{
		peer:           p,
		log:            log,
		mtx:            &sync.Mutex{},
		arm:            &sync.Mutex{},
		unsignedTxPool: txpool.NewTxPool(p.cfg.TxPool.Capacity),
		openTxSessions: make(map[string]struct{}),
	}
}

// LocalPeer returns the local peer
func (protoc *Inception) LocalPeer() *Node {
	return protoc.peer
}

// PM returns the local peer's peer manager
func (protoc *Inception) PM() *Manager {
	if protoc.peer == nil {
		return nil
	}
	return protoc.peer.PM()
}

// sign takes an object, marshals it to JSON and signs it
func (protoc *Inception) sign(msg interface{}) []byte {
	bs, _ := json.Marshal(msg)
	key := protoc.LocalPeer().PrivKey()
	sig, _ := key.Sign(bs)
	return sig
}

// verify verifies a signature
func (protoc *Inception) verify(msg interface{}, sig []byte, pKey ic.PubKey) error {
	bs, _ := json.Marshal(msg)
	result, err := pKey.Verify(bs, sig)
	if err != nil {
		return fmt.Errorf("failed to verify -> %s", err)
	}
	if !result {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// reject sends a reject message.
// The caller is expected to close the stream
func (protoc *Inception) reject(s net.Stream, msg string, code int, reason string, extraData []byte) error {
	rMsg := wire.NewRejectMsg(msg, int32(code), reason, extraData)
	w := bufio.NewWriter(s)
	if err := pc.Multicodec(nil).Encoder(w).Encode(rMsg); err != nil {
		return fmt.Errorf("reject message failed. failed to write to stream")
	}
	w.Flush()
	return nil
}

// isRejected checks if the message is a `reject`.
// Returns the message`
func (protoc *Inception) isRejected(s net.Stream) (*wire.Reject, error) {

	var msg wire.Reject

	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to read from stream. %s", err)
	}

	if msg.Code != 0 {
		return &msg, constants.ErrRejected
	}

	return nil, nil
}

// GetUnSignedTxPool returns the unsigned transaction pool
func (protoc *Inception) GetUnSignedTxPool() *txpool.TxPool {
	return protoc.unsignedTxPool
}

// GetUnsignedTxRelayQueue returns the unsigned transaction relay queue
func (protoc *Inception) GetUnsignedTxRelayQueue() *txpool.TxQueue {
	return protoc.txsRelayQueue
}
