package types

import (
	"math/big"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
)

const (
	// NamespaceState is the namespace
	// for RPC methods that access the database
	NamespaceState = "state"

	// NamespaceEll is the namespace for RPC methods
	// that interact with the native currency
	NamespaceEll = "ell"

	// NamespaceNode is the namespace for RPC methods
	// that interact and access the node/client properties
	NamespaceNode = "node"

	// NamespacePool is the namespace for RPC methods
	// that access the transaction pool
	NamespacePool = "pool"

	// NamespaceMiner is the namespace for RPC methods
	// that interact with the miner
	NamespaceMiner = "miner"

	// NamespacePersonal is the namespace for RPC methods
	// that interact with private and sensitive data of the
	// client
	NamespacePersonal = "personal"

	// NamespaceAdmin is the namespace for RPC methods
	// that perform administrative actions
	NamespaceAdmin = "admin"

	// NamespaceNet is the namespace for RPC methods
	// that perform network actions
	NamespaceNet = "net"

	// NamespaceRPC is the namespace for RPC methods
	// that perform rpc actions
	NamespaceRPC = "rpc"

	// NamespaceLogger is the namespace for RPC methods
	// for configuring the logger
	NamespaceLogger = "logger"

	// NamespaceBurner is the namespace for RPC methods
	// for burning coins
	NamespaceBurner = "burner"

	// NamespaceDebug is the namespace for RPC methods
	// that offer debugging features
	NamespaceDebug = "debug"
)

// ValidationContext is used to
// represent a validation behaviour
type ValidationContext int

const (
	// ContextBlock represents validation
	// context of which the intent is to validate
	// a block that needs to be appended to a chain
	ContextBlock ValidationContext = iota + 1

	// ContextBranch represents validation
	// context of which the intent is to validate
	// a block that needs to be appended to a branch chain
	ContextBranch

	// ContextTxPool represents validation context
	// in which the intent is to validate a
	// transaction that needs to be included in
	// the transaction pool.
	ContextTxPool
)

// GenerateBlockParams represents parameters
// required for block generation.
type GenerateBlockParams struct {

	// OverrideParentHash explicitly sets the parent hash
	OverrideParentHash util.Hash

	// Transactions sets the block transaction.
	// If not provided, transactions are selected from
	// the transaction pool
	Transactions []Transaction

	// Creator sets the key of the block creator.
	// Required for setting the creator public key
	// and signing the block
	Creator *crypto.Key

	// Nonce is the special number that
	// indicates the completion of PoW
	Nonce util.BlockNonce

	// Difficulty represents the target
	// difficulty that the nonce satisfied
	Difficulty *big.Int

	// OverrideTotalDifficulty explicitly sets
	// the total difficulty.
	OverrideTotalDifficulty *big.Int

	// StateRoot sets the state root
	OverrideStateRoot util.Hash

	// Timestamp sets the time of block creation
	OverrideTimestamp int64

	// ChainTip explicitly sets the chain tip number.
	// This is used to alter what the tip of a chain is.
	// It is used in tests.
	OverrideChainTip uint64

	// AddFeeAlloc if set to true, will calculate the
	// miner fee reward and add an Alloc transaction
	AddFeeAlloc bool

	// NoPoolAdditionInTest prevents the transactions
	// from automatically added to the pool. Only used
	// during block generation in tests
	NoPoolAdditionInTest bool
}

// SyncPeerChainInfo holds information about
// a peer that is a potential sync peer
type SyncPeerChainInfo struct {

	// PeerID is the sync peer ID
	PeerID string

	// PeerIDShort is the short loggable
	// version of PeerID
	PeerIDShort string

	// PeerChainHeight is the height of the
	// sync peer main chain
	PeerChainHeight uint64

	// PeerChainTD is the total difficulty
	// of the sync peer's main chain
	PeerChainTD *big.Int

	// LastBlockSent is the last block
	// received from the sync peer and
	// was processed by the local peer.
	LastBlockSent util.Hash
}

// ConnectError represents a connection error
type ConnectError string

func (c ConnectError) Error() string {
	return string(c)
}
