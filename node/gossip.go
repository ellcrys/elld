package node

import (
	"bufio"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"

	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	ic "github.com/libp2p/go-libp2p-crypto"
	net "github.com/libp2p/go-libp2p-net"
)

const (
	// EventReceivedBlockHashes describes an event about
	// a receiving block hashes
	EventReceivedBlockHashes = "event.receivedBlockHashes"
	// EventRequestedBlockHashes describes an event about
	// sending a request for block hashes
	EventRequestedBlockHashes = "event.requestedBlockHashes"
	// EventTransactionProcessed describes an event about
	// a processed transaction
	EventTransactionProcessed = "event.transactionProcessed"
	// EventBlockProcessed describes an event about
	// a processed block
	EventBlockProcessed = "event.blockProcessed"
	// EventBlockBodiesProcessed describes an event about
	// processed block bodies
	EventBlockBodiesProcessed = "event.blockBodiesProcessed"
	//EventAddrProcessed describes an event about
	// a processed address
	EventAddrProcessed = "event.addrProcessed"
	// EventAddressesRelayed describes an event about
	// relayed addresses
	EventAddressesRelayed = "event.addressesRelayed"
	// EventReceivedAddr describes an event about a
	// received addresses
	EventReceivedAddr = "event.receivedAddr"
	// EventIntroReceived describes an event about
	// a received intro
	EventIntroReceived = "event.receivedIntro"
)

// BroadcastPeers is a type that contains
// randomly chosen peers that messages will be
// broadcast to.
type BroadcastPeers struct {
	sync.RWMutex
	peers map[string]*Node
}

// Has checks whether a peer exists
func (b *BroadcastPeers) Has(p *Node) bool {
	b.RLock()
	defer b.RUnlock()
	_, has := b.peers[p.StringID()]
	return has
}

// Add adds a peer
func (b *BroadcastPeers) Add(p *Node) {
	b.Lock()
	defer b.Unlock()
	b.peers[p.StringID()] = p
}

// Clear removes all peers
func (b *BroadcastPeers) Clear() {
	b.Lock()
	defer b.Unlock()
	b.peers = make(map[string]*Node)
}

// Len returns the number of peers
func (b *BroadcastPeers) Len() int {
	b.RLock()
	defer b.RUnlock()
	return len(b.peers)
}

// Peers returns the stored peers
func (b *BroadcastPeers) Peers() (peers []*Node) {
	b.RLock()
	defer b.RUnlock()
	for _, p := range b.peers {
		peers = append(peers, p)
	}
	return
}

// PeersID returns the id of the stored peers
func (b *BroadcastPeers) PeersID() (ids []string) {
	b.RLock()
	defer b.RUnlock()
	for id := range b.peers {
		ids = append(ids, id)
	}
	return
}

// Gossip represents the peer protocol
type Gossip struct {

	// mtx is the general mutex
	mtx *sync.RWMutex

	// engine represents the local node
	engine *Node

	// log is used for logging events
	log logger.Logger

	// broadcastersUpdatedAt is the time the
	// last relay peers where selected
	broadcastersUpdatedAt time.Time

	// broadcasters contains randomly selected
	// peers to broadcast messages to.
	broadcasters *BroadcastPeers
}

// NewGossip creates a new instance of the Gossip protocol
func NewGossip(p *Node, log logger.Logger) *Gossip {
	return &Gossip{
		engine: p,
		log:    log,
		mtx:    &sync.RWMutex{},
		broadcasters: &BroadcastPeers{
			peers: make(map[string]*Node),
		},
	}
}

// GetBlockchain returns the blockchain manager
func (g *Gossip) GetBlockchain() core.Blockchain {
	return g.engine.bchain
}

// NewStream creates a stream for a given protocol
// ID and between the local peer and the given remote peer.
func (g *Gossip) NewStream(remotePeer types.Engine, msgVersion string) (net.Stream, context.CancelFunc, error) {
	ctxDur := time.Second * time.Duration(g.engine.cfg.Node.MessageTimeout)
	ctx, cf := context.WithTimeout(context.TODO(), ctxDur)
	s, err := g.engine.addToPeerStore(remotePeer).
		newStream(ctx, remotePeer.ID(), msgVersion)
	if err != nil {
		cf()
	}
	return s, cf, err
}

// ReadStream reads the content of a steam into dest
func ReadStream(s net.Stream, dest interface{}) error {
	return msgpack.NewDecoder(bufio.NewReader(s)).Decode(dest)
}

// WriteStream writes msg to the given stream
func WriteStream(s net.Stream, msg interface{}) error {
	w := bufio.NewWriter(s)
	if err := msgpack.NewEncoder(w).Encode(msg); err != nil {
		return err
	}
	w.Flush()
	return nil
}

func (g *Gossip) logErr(err error, rp types.Engine, msg string) error {
	g.log.Debug(msg, "Err", err, "PeerID", rp.ShortID())
	return err
}

// logConnectErr updates the failure count record of a node
// that failed to connect. It will also add a 1 hour ban time
// if the node failed to connect after n tries.
func (g *Gossip) logConnectErr(err error, rp types.Engine, msg string) error {

	// Increase connection fail count
	g.PM().IncrConnFailCount(rp)

	// When the peer reaches the max allowed
	// failure count, add a ban time fo 3 hours
	if g.PM().GetConnFailCount(rp) >= 3 {
		g.PM().AddTimeBan(rp, 1*time.Hour)
	}

	return g.logErr(err, rp, msg)
}

// Engine returns the local node
func (g *Gossip) Engine() *Node {
	return g.engine
}

// GetBroadcasters returns the broadcasters
func (g *Gossip) GetBroadcasters() *BroadcastPeers {
	return g.broadcasters
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
	bs, _ := msgpack.Marshal(msg)
	key := g.Engine().PrivKey()
	sig, _ := key.Sign(bs)
	return sig
}

// verify verifies a signature
func (g *Gossip) verify(msg interface{}, sig []byte, pKey ic.PubKey) error {
	bs, _ := msgpack.Marshal(msg)
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
	rMsg := wire.Reject{
		Message:   msg,
		Code:      int32(code),
		Reason:    reason,
		ExtraData: extraData,
	}
	if err := WriteStream(s, rMsg); err != nil {
		return fmt.Errorf("reject message failed. failed to write to stream")
	}
	return nil
}

// isRejected checks if the message is a `reject`.
// Returns the message`
func (g *Gossip) isRejected(s net.Stream) (*wire.Reject, error) {

	var msg wire.Reject
	if err := ReadStream(s, &msg); err != nil {
		return nil, fmt.Errorf("failed to read from stream. %s", err)
	}

	if msg.Code != 0 {
		return &msg, params.ErrRejected
	}

	return nil, nil
}
