package node

import (
	"bufio"
	"context"
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

// Gossip represents the peer protocol
type Gossip struct {
	mtx                         *sync.Mutex         // main mutex
	version                     string              // the protocol version
	engine                      *Node               // the local peer
	log                         logger.Logger       // the logger
	lastRelayPeersSelectionTime time.Time           // the time the last peers responsible for relaying "addr" messages where selected
	addrRelayPeers              [2]*Node            // peers to relay addr msgs to
	unsignedTxPool              *txpool.TxPool      // the transaction pool for unsigned transactions
	openTxSessions              map[string]struct{} // Holds the id of transactions awaiting endorsement. Protected by mtx.
	txsRelayQueue               *txpool.TxQueue     // stores transactions waiting to be relayed
}

// NewGossip creates a new instance of the protocol codenamed "Gossip"
func NewGossip(p *Node, log logger.Logger) *Gossip {
	return &Gossip{
		engine:         p,
		log:            log,
		mtx:            &sync.Mutex{},
		unsignedTxPool: txpool.NewTxPool(p.cfg.TxPool.Capacity),
		openTxSessions: make(map[string]struct{}),
		txsRelayQueue:  txpool.NewQueueNoSort(p.cfg.TxPool.Capacity),
	}
}

func (g *Gossip) newStream(ctx context.Context, remotePeer *Node, msgVersion string) (net.Stream, error) {
	return g.engine.addToPeerStore(remotePeer).newStream(ctx, remotePeer.ID(), msgVersion)
}

func readStream(s net.Stream, dest interface{}) error {
	return pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(dest)
}

func writeStream(s net.Stream, msg interface{}) error {
	w := bufio.NewWriter(s)
	if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
		return err
	}
	w.Flush()
	return nil
}

// Engine returns the local node
func (g *Gossip) Engine() *Node {
	return g.engine
}

// PM returns the local peer's peer manager
func (g *Gossip) PM() *Manager {
	if g.engine == nil {
		return nil
	}
	return g.engine.PM()
}

// sign takes an object, marshals it to JSON and signs it
func (g *Gossip) sign(msg interface{}) []byte {
	bs, _ := json.Marshal(msg)
	key := g.Engine().PrivKey()
	sig, _ := key.Sign(bs)
	return sig
}

// verify verifies a signature
func (g *Gossip) verify(msg interface{}, sig []byte, pKey ic.PubKey) error {
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
// The caller is expected to close the stream after the call.
func (g *Gossip) reject(s net.Stream, msg string, code int, reason string, extraData []byte) error {
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
func (g *Gossip) isRejected(s net.Stream) (*wire.Reject, error) {

	var msg wire.Reject
	if err := readStream(s, &msg); err != nil {
		return nil, fmt.Errorf("failed to read from stream. %s", err)
	}

	if msg.Code != 0 {
		return &msg, constants.ErrRejected
	}

	return nil, nil
}

// GetUnSignedTxPool returns the unsigned transaction pool
func (g *Gossip) GetUnSignedTxPool() *txpool.TxPool {
	return g.unsignedTxPool
}

// GetUnsignedTxRelayQueue returns the unsigned transaction relay queue
func (g *Gossip) GetUnsignedTxRelayQueue() *txpool.TxQueue {
	return g.txsRelayQueue
}
