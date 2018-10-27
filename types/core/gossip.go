package core

import (
	"context"
	"math/big"
	"sync"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	net "github.com/libp2p/go-libp2p-net"
	"github.com/vmihailenco/msgpack"
)

// Handshake represents the first
// message between peers
type Handshake struct {
	Version                  string    `json:"version" msgpack:"version"`
	BestBlockHash            util.Hash `json:"bestBlockHash" msgpack:"bestBlockHash"`
	BestBlockTotalDifficulty *big.Int  `json:"bestBlockTD" msgpack:"bestBlockTD"`
	BestBlockNumber          uint64    `json:"bestBlockNumber" msgpack:"bestBlockNumber"`
}

// EncodeMsgpack implements
// msgpack.CustomEncoder
func (h *Handshake) EncodeMsgpack(enc *msgpack.Encoder) error {
	tdStr := h.BestBlockTotalDifficulty.String()
	return enc.Encode(h.Version, h.BestBlockHash, h.BestBlockNumber, tdStr)
}

// DecodeMsgpack implements
// msgpack.CustomDecoder
func (h *Handshake) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tdStr string
	if err := dec.Decode(&h.Version, &h.BestBlockHash, &h.BestBlockNumber, &tdStr); err != nil {
		return err
	}
	h.BestBlockTotalDifficulty, _ = new(big.Int).SetString(tdStr, 10)
	return nil
}

// GetAddr is used to request for peer
// addresses from other peers
type GetAddr struct {
}

// Addr is used to send peer addresses
// in response to a GetAddr
type Addr struct {
	Addresses []*Address `json:"addresses" msgpack:"addresses"`
}

// Address represents a peer's address
type Address struct {
	Address   util.NodeAddr `json:"address" msgpack:"address"`
	Timestamp int64         `json:"timestamp" msgpack:"timestamp"`
}

// Ping represents a ping message
type Ping struct {
	BestBlockHash            util.Hash `json:"bestBlockHash" msgpack:"bestBlockHash"`
	BestBlockTotalDifficulty *big.Int  `json:"bestBlockTD" msgpack:"bestBlockTD"`
	BestBlockNumber          uint64    `json:"bestBlockNumber" msgpack:"bestBlockNumber"`
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (p *Ping) EncodeMsgpack(enc *msgpack.Encoder) error {
	tdStr := p.BestBlockTotalDifficulty.String()
	return enc.Encode(p.BestBlockHash, p.BestBlockNumber, tdStr)
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (p *Ping) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tdStr string
	if err := dec.Decode(&p.BestBlockHash, &p.BestBlockNumber, &tdStr); err != nil {
		return err
	}
	p.BestBlockTotalDifficulty, _ = new(big.Int).SetString(tdStr, 10)
	return nil
}

// Pong represents a pong message
type Pong struct {
	BestBlockHash            util.Hash `json:"bestBlockHash" msgpack:"bestBlockHash"`
	BestBlockTotalDifficulty *big.Int  `json:"bestBlockTD" msgpack:"bestBlockTD"`
	BestBlockNumber          uint64    `json:"bestBlockNumber" msgpack:"bestBlockNumber"`
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (p *Pong) EncodeMsgpack(enc *msgpack.Encoder) error {
	tdStr := p.BestBlockTotalDifficulty.String()
	return enc.Encode(p.BestBlockHash, p.BestBlockNumber, tdStr)
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (p *Pong) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tdStr string
	if err := dec.Decode(&p.BestBlockHash, &p.BestBlockNumber, &tdStr); err != nil {
		return err
	}
	p.BestBlockTotalDifficulty, _ = new(big.Int).SetString(tdStr, 10)
	return nil
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
	Hash string `json:"hash" msgpack:"hash"`
}

// GetBlockHashes represents a message requesting
// for headers of blocks. The locator is used to
// compare with a remote node to determine which
// blocks to send back.
type GetBlockHashes struct {
	Locators  []util.Hash `json:"hash" msgpack:"hash"`
	MaxBlocks int64       `json:"maxBlocks" msgpack:"maxBlocks"`
}

// BlockHashes represents a message containing
// block hashes as a response to GetBlockHeaders
type BlockHashes struct {
	Hashes []util.Hash
}

// BlockHash represents a hash of a block
// sent by a remote peer
type BlockHash struct {
	Hash        util.Hash
	Broadcaster Engine
}

// BlockBody represents the body of a block
type BlockBody struct {
	Header       *Header        `json:"header" msgpack:"header"`
	Transactions []*Transaction `json:"transactions" msgpack:"transactions"`
	Hash         util.Hash      `json:"hash" msgpack:"hash"`
	Sig          []byte         `json:"sig" msgpack:"sig"`
}

// BlockBodies represents a collection of block bodies
type BlockBodies struct {
	Blocks []*BlockBody
}

// GetBlockBodies represents a message to fetch block bodies
// belonging to the given hashes
type GetBlockBodies struct {
	Hashes []util.Hash
}

// Intro represents a message describing a peer's ID.
type Intro struct {
	PeerID string `json:"id" msgpack:"id"`
}

// Hash returns the hash representation
func (m *Intro) Hash() util.Hash {
	bs := util.ObjectToBytes([]interface{}{m.PeerID})
	return util.BytesToHash(util.Blake2b256(bs))
}

// Gossip represent messages and interactions between nodes
type Gossip interface {

	// Address messages
	OnAddr(s net.Stream) error
	RelayAddresses(addrs []*Address) []error

	// Block messages
	RelayBlock(block types.Block, remotePeers []Engine) error
	OnBlockBody(s net.Stream) error
	RequestBlock(rp Engine, blockHash util.Hash) error
	OnRequestBlock(s net.Stream) error
	SendGetBlockHashes(rp Engine, locators []util.Hash) error
	OnGetBlockHashes(s net.Stream) error
	SendGetBlockBodies(rp Engine, hashes []util.Hash) error
	OnGetBlockBodies(s net.Stream) error

	// Handshake messages
	SendHandshake(rp Engine) error
	OnHandshake(s net.Stream) error

	// GetAddr messages
	SendGetAddrToPeer(remotePeer Engine) ([]*Address, error)
	SendGetAddr(remotePeers []Engine) error
	OnGetAddr(s net.Stream) error

	// Ping messages
	SendPing(remotePeers []Engine)
	SendPingToPeer(remotePeer Engine) error
	OnPing(s net.Stream) error

	// Node advertisement
	SelfAdvertise(connectedPeers []Engine) int

	// Introductory messages
	SendIntro(intro *Intro)
	OnIntro(s net.Stream) error

	// Transaction messages
	RelayTx(tx types.Transaction, remotePeers []Engine) error
	OnTx(s net.Stream) error

	// PickBroadcasters selects N random addresses from
	// the given slice of addresses and caches them to
	// be used as broadcasters.
	// They are returned on subsequent calls and only
	// renewed when there are less than N addresses or the
	// cache is over 24 hours since it was last updated.
	PickBroadcasters(addresses []*Address, n int) *BroadcastPeers

	// GetBroadcasters returns the broadcasters
	GetBroadcasters() *BroadcastPeers

	// NewStream creates a stream for a given protocol
	// ID and between the local peer and the given remote peer.
	NewStream(remotePeer Engine, msgVersion string) (net.Stream,
		context.CancelFunc, error)
}

// BroadcastPeers is a type that contains
// randomly chosen peers that messages will be
// broadcast to.
type BroadcastPeers struct {
	sync.RWMutex
	peers map[string]Engine
}

// NewBroadcastPeers creates a new BroadcastPeers instance
func NewBroadcastPeers() *BroadcastPeers {
	return &BroadcastPeers{
		peers: make(map[string]Engine),
	}
}

// Has checks whether a peer exists
func (b *BroadcastPeers) Has(p Engine) bool {
	b.RLock()
	defer b.RUnlock()
	_, has := b.peers[p.StringID()]
	return has
}

// Add adds a peer
func (b *BroadcastPeers) Add(p Engine) {
	b.Lock()
	defer b.Unlock()
	b.peers[p.StringID()] = p
}

// Clear removes all peers
func (b *BroadcastPeers) Clear() {
	b.Lock()
	defer b.Unlock()
	b.peers = make(map[string]Engine)
}

// Len returns the number of peers
func (b *BroadcastPeers) Len() int {
	b.RLock()
	defer b.RUnlock()
	return len(b.peers)
}

// Peers returns the stored peers
func (b *BroadcastPeers) Peers() (peers []Engine) {
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
