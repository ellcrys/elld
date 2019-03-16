package types

import (
	"math/big"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/merkletree"
)

// Chainer (a.k.a Chains) defines an interface for accessing
// mutating and managing a collection of blocks
type Chainer interface {

	// GetStore returns the store
	GetStore() ChainStorer

	// NewStateTree returns a new tree
	NewStateTree(opts ...CallOp) (Tree, error)

	// Current gets the header of the tip block
	Current(opts ...CallOp) (Header, error)

	// GetID gets the chain ID
	GetID() util.String

	// GetBlock gets a block in the chain
	GetBlock(uint64, ...CallOp) (Block, error)

	// GetParentBlock gets the chain's parent block if it has one
	GetParentBlock() Block

	// GetInfo gets the chain's parent information
	GetInfo() ChainInfo

	// CreateAccount creates an account on a target block
	CreateAccount(targetBlockNum uint64, account Account, opts ...CallOp) error

	// GetAccount gets an account
	GetAccount(address util.String, opts ...CallOp) (Account, error)

	// PutTransactions stores a collection of transactions
	PutTransactions(txs []Transaction, blockNumber uint64, opts ...CallOp) error

	// GetTransaction gets a transaction by hash
	GetTransaction(hash util.Hash, opts ...CallOp) (Transaction, error)

	// ChainReader gets a chain reader for this chain
	ChainReader() ChainReaderFactory

	// GetRoot fetches the root block of this chain. If the chain
	// has more than one parents/ancestors, it will traverse
	// the parents to return the root parent block.
	GetRoot() Block
}

// ChainStorer defines an interface that provides
// every functionality need to mutate or query a
// chain.
type ChainStorer interface {

	// PutBlock adds a block to the store
	PutBlock(block Block, opts ...CallOp) error

	// GetBlock finds and returns a block in the chain.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, opts ...CallOp) (Block, error)

	// GetBlockByHash finds a block by its hash
	GetBlockByHash(hash util.Hash, opts ...CallOp) (Block, error)

	// GetBlockByNumberAndHash finds by number and hash
	GetBlockByNumberAndHash(number uint64, hash util.Hash, opts ...CallOp) (Block, error)

	// GetHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetHeader(number uint64, opts ...CallOp) (Header, error)

	// GetHeaderByHash finds and returns the header of a block matching hash
	GetHeaderByHash(hash util.Hash, opts ...CallOp) (Header, error)

	// GetTransaction gets a transaction (by hash) belonging to the chain
	GetTransaction(hash util.Hash, opts ...CallOp) (Transaction, error)

	// CreateAccount creates an account on a target block
	CreateAccount(targetBlockNum uint64, account Account, opts ...CallOp) error

	// GetAccount gets an account
	GetAccount(address util.String, opts ...CallOp) (Account, error)

	// GetAccounts gets an account
	GetAccounts(opts ...CallOp) ([]Account, error)

	// PutTransactions stores a collection of transactions
	PutTransactions(txs []Transaction, blockNumber uint64, opts ...CallOp) error

	// PutMinedBlock stores a brief information about a
	// block that was created by the blockchain's coinbase key
	PutMinedBlock(block Block, opts ...CallOp) error

	// Current gets the current block at the tip of the chain
	Current(opts ...CallOp) (Block, error)

	// Delete deletes objects
	Delete(key []byte, opts ...CallOp) error

	// NewTx creates and returns a transaction
	NewTx() (elldb.Tx, error)

	// DB gets the database
	DB() elldb.DB
}

// Blockchain defines an interface for a blockchain manager
type Blockchain interface {

	// Up initializes and loads the blockchain manager
	Up() error

	// SetCoinbase sets the coinbase key that is used to
	// identify the current blockchain instance
	SetCoinbase(coinbase *crypto.Key)

	// GetBestChain gets the chain that is currently considered the main chain
	GetBestChain() Chainer

	// IsKnownBlock checks if a block is stored in the main or side chain or orphan
	IsKnownBlock(hash util.Hash) (bool, string, error)

	// HaveBlock checks whether we have a block matching the hash in any of the known chains
	HaveBlock(hash util.Hash) (bool, error)

	// GetTransaction finds and returns a transaction on the main chain
	GetTransaction(util.Hash, ...CallOp) (Transaction, error)

	// ProcessBlock attempts to process and append a block to the main or side chains
	ProcessBlock(Block, ...CallOp) (ChainReaderFactory, error)

	// Generate creates a new block for a target chain.
	// The Chain is specified by passing to OpChain.
	Generate(*GenerateBlockParams, ...CallOp) (Block, error)

	// ChainReader gets a Reader for reading the main chain
	ChainReader() ChainReaderFactory

	// GetChainsReader gets chain reader for all known chains
	GetChainsReader() (readers []ChainReaderFactory)

	// SetDB sets the database
	SetDB(elldb.DB)

	// OrphanBlocks gets a reader for the orphan cache
	OrphanBlocks() CacheReader

	// GetEventEmitter gets the event emitter
	GetEventEmitter() *emitter.Emitter

	// GetBlock finds a block in any chain with a matching
	// block number and hash.
	GetBlock(number uint64, hash util.Hash) (Block, error)

	// GetBlockByHash finds a block in any chain with a matching hash.
	GetBlockByHash(hash util.Hash, opts ...CallOp) (Block, error)

	// GetChainReaderByHash returns a chain reader to a chain
	// where a block with the given hash exists
	GetChainReaderByHash(hash util.Hash) ChainReaderFactory

	// SetGenesisBlock sets the genesis block
	SetGenesisBlock(block Block)

	// GetTxPool gets the transaction pool
	GetTxPool() TxPool

	// CreateAccount creates an account that is associated with
	// the given block number and chain.
	CreateAccount(blockNo uint64, chain Chainer, account Account) error

	// GetAccount gets an account
	GetAccount(address util.String, opts ...CallOp) (Account, error)

	// GetAccountNonce gets the nonce of an account
	GetAccountNonce(address util.String, opts ...CallOp) (uint64, error)

	// GetLocators fetches a list of blockhashes used to
	// compare and sync the local chain with a remote chain.
	GetLocators() ([]util.Hash, error)

	// SelectTransactions sets transactions from
	// the transaction pool. These transactions must
	// be suitable for inclusion in blocks.
	SelectTransactions(maxSize int64) ([]Transaction, error)
}

// BlockMaker defines an interface providing the
// necessary functions to create new blocks
type BlockMaker interface {

	// Generate creates a new block for a target chain.
	// The Chain is specified by passing to OpChain.
	Generate(*GenerateBlockParams, ...CallOp) (Block, error)

	// ChainReader gets a Reader for reading the main chain
	ChainReader() ChainReaderFactory

	// ProcessBlock attempts to process and append a block to the main or side chains
	ProcessBlock(Block, ...CallOp) (ChainReaderFactory, error)

	// IsMainChain checks whether a chain is the main chain
	IsMainChain(ChainReaderFactory) bool
}

// ChainReaderFactory defines an interface for reading a chain
type ChainReaderFactory interface {

	// GetID gets the chain ID
	GetID() util.String

	// GetBlock finds and returns a block associated with chainID.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, opts ...CallOp) (Block, error)

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(hash util.Hash, opts ...CallOp) (Block, error)

	// GetHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetHeader(number uint64, opts ...CallOp) (Header, error)

	// GetHeaderByHash finds and returns the header of a block matching hash
	GetHeaderByHash(hash util.Hash, opts ...CallOp) (Header, error)

	// Current gets the current block at the tip of the chain
	Current(opts ...CallOp) (Block, error)

	// GetParent returns a chain reader to the parent chain.
	// Returns nil if chain has no parent.
	GetParent() ChainReaderFactory

	// GetParentBlock returns the parent block
	GetParentBlock() Block

	// GetRoot fetches the root block of this chain. If the chain
	// has more than one parents/ancestors, it will traverse
	// the parents to return the root parent block.
	GetRoot() Block
}

// CacheReader provides an interface for reading the orphan cache
type CacheReader interface {

	// Len gets the number of orphans
	Len() int

	// Hash checks whether an item exists in the cache
	Has(key interface{}) bool

	// Get gets an item from the cache
	Get(key interface{}) interface{}
}

// Block defines a block
type Block interface {
	ComputeHash() util.Hash
	GetBytes() []byte
	GetHeader() Header
	SetHeader(h Header)
	ReplaceHeader(header Header) Block
	GetNumber() uint64
	GetHash() util.Hash
	SetHash(util.Hash)
	GetTransactions() []Transaction
	GetSignature() []byte
	SetSignature(sig []byte)
	GetHashAsHex() string
	GetBytesNoHashSig() []byte
	GetSize() int64
	GetSizeNoTxs() int64
	GetBytesNoTxs() []byte
	GetValidationContexts() []ValidationContext
	SetValidationContexts(...ValidationContext)
	SetSig([]byte)
}

// Header defines a block header containing
// metadata about the block
type Header interface {
	GetNumber() uint64
	SetNumber(uint64)
	GetHashNoNonce() util.Hash
	GetBytes() []byte
	ComputeHash() util.Hash
	GetExtra() []byte
	GetTimestamp() int64
	SetTimestamp(int64)
	GetDifficulty() *big.Int
	SetDifficulty(*big.Int)
	GetNonce() util.BlockNonce
	SetNonce(nonce util.BlockNonce)
	GetParentHash() util.Hash
	SetParentHash(util.Hash)
	SetCreatorPubKey(util.String)
	GetCreatorPubKey() util.String
	Copy() Header
	SetStateRoot(util.Hash)
	GetStateRoot() util.Hash
	SetTransactionsRoot(txRoot util.Hash)
	GetTransactionsRoot() util.Hash
	GetTotalDifficulty() *big.Int
	SetTotalDifficulty(*big.Int)
}

// Account defines an interface for an account
type Account interface {
	GetAddress() util.String
	GetBalance() util.String
	SetBalance(util.String)
	GetNonce() uint64
	IncrNonce()
}

// Transaction represents a transaction
type Transaction interface {
	GetHash() util.Hash
	SetHash(util.Hash)
	GetBytesNoHashAndSig() []byte
	GetSizeNoFee() int64
	ComputeHash() util.Hash
	GetID() string
	Sign(privKey string) ([]byte, error)
	GetType() int64
	GetFrom() util.String
	SetFrom(util.String)
	GetTo() util.String
	GetValue() util.String
	SetValue(util.String)
	GetFee() util.String
	GetNonce() uint64
	GetTimestamp() int64
	SetTimestamp(int64)
	GetSenderPubKey() util.String
	SetSenderPubKey(util.String)
	GetSignature() []byte
	SetSignature(sig []byte)
}

// CallOp describes an interface to be used to define store method options
type CallOp interface {
	GetName() string
}

// Tree defines a merkle tree
type Tree interface {
	Add(item merkletree.Content)
	GetItems() []merkletree.Content
	Build() error
	Root() util.Hash
}

// TxContainer represents a container
// a container responsible for holding
// and sorting transactions
type TxContainer interface {
	ByteSize() int64
	Add(tx Transaction) bool
	Has(tx Transaction) bool
	Size() int64
	First() Transaction
	Last() Transaction
	Sort()
	IFind(predicate func(Transaction) bool) Transaction
	Remove(txs ...Transaction)
}

// TxPool represents a transactions pool
type TxPool interface {
	Put(tx Transaction) error
	Has(tx Transaction) bool
	HasByHash(hash string) bool
	Remove(txs ...Transaction)
	ByteSize() int64
	Size() int64
	Container() TxContainer
	GetByHash(hash string) Transaction
	GetByFrom(address util.String) []Transaction
}

// ChainInfo represents a chain's metadata
type ChainInfo interface {
	GetID() util.String
	GetParentChainID() util.String
	GetParentBlockNumber() uint64
	GetTimestamp() int64
}
