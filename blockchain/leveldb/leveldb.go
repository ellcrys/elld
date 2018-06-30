package leveldb

import (
	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/wire"
)

// DBName production database name
const DBName = "elld"

// Store represents a store that implements the Store
// interface meant to be used for persisting and retrieving
// blockchain data.
type Store struct {
	db        database.DB
	namespace string
}

// New creates an instance of the store
func New(namespace, cfgDir string) (*Store, error) {
	d := database.NewLevelDB(cfgDir)
	return &Store{
		db:        d,
		namespace: namespace,
	}, nil
}

// Initialize initializes the database and creates necessary tables.
// - Created database if it does not exist.
func (s *Store) Initialize() error {
	return s.db.Open(s.namespace)
}

// DropDB deletes the bucket
func (s *Store) DropDB() error {
	return nil
}

func (s *Store) hasBlock(number uint64) (bool, error) {
	return false, nil
}

func (s *Store) getBlock(number int64, fields []string, result interface{}) error {

	return nil
}

// GetBlockHeader the header of the current block
func (s *Store) GetBlockHeader(number int64, header *wire.Header) error {
	return nil
}

// GetBlock fetches a block by its block number.
// If the block number begins with -1, the block with the highest block number is returned.
func (s *Store) GetBlock(number int64, result types.Block) error {
	return s.getBlock(number, nil, result)
}

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *Store) PutBlock(block types.Block) error {
	return nil
}

// Close closes the store
func (s *Store) Close() error {
	return nil
}
