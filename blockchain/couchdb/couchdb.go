package couchdb

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	"gopkg.in/resty.v1"
)

// DBName production database name
const DBName = "elld"

// Store represents a store that implements the Store
// interface meant to be used for persisting and retrieving
// blockchain data.
type Store struct {
	dbName string
	addr   string
}

// New creates an instance of the store
func New(dbName, addr string) (*Store, error) {
	_, err := resty.R().Get(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to reach database at specified address. %s", err)
	}

	return &Store{
		dbName: dbName,
		addr:   addr,
	}, nil
}

// Initialize initializes the database and creates necessary tables.
// - Created database if it does not exist.
func (s *Store) Initialize() error {

	// Attempt to get database info.
	resp, err := resty.R().Get(s.addr + "/" + s.dbName)
	if err != nil {
		return fmt.Errorf("failed to check database existence. %s", err)
	}

	switch resp.StatusCode() {
	case 404:

		// Since database does not exists, we attempt create it
		resp, err = resty.R().Put(s.addr + "/" + s.dbName)
		if err != nil {
			return fmt.Errorf("failed to create database. %s", resp.String())
		}
		if resp.StatusCode() != 201 {
			return fmt.Errorf("failed to create database. %s", resp.String())
		}

		// Next we create indexes
		if err = s.createIndexes(); err != nil {
			return fmt.Errorf("failed to create indexes. %s", resp.String())
		}

	case 200:
		return nil
	default:
		return fmt.Errorf("failed to check database existence. %s", resp.String())
	}

	return nil
}

func (s *Store) createIndexes() error {

	body := map[string]interface{}{
		"name": "number_index",
		"type": "json",
		"index": map[string]interface{}{
			"fields": []string{"header.number"},
		},
	}

	var _result map[string]interface{}
	_, err := resty.R().
		SetBody(body).
		SetResult(&_result).
		SetPathParams(map[string]string{"db": s.dbName}).
		Post(s.addr + "/{db}/_index")
	if err != nil {
		return fmt.Errorf("failed to get block. %s", err)
	}

	if docs, ok := _result["docs"].([]interface{}); ok && len(docs) > 0 {
		return nil
	}

	return nil
}

func (s *Store) DropDB() error {
	resp, err := resty.R().SetPathParams(map[string]string{"db": s.dbName}).Delete(s.addr + "/{db}")
	if err != nil {
		return fmt.Errorf("failed to delete database. %s", err)
	}

	switch resp.StatusCode() {
	case 200, 404:
		return nil
	default:
		return fmt.Errorf("failed to delete database. %s", resp.String())
	}
}

func (s *Store) hasBlock(number uint64) (bool, error) {

	query := map[string]interface{}{
		"limit": 1,
		"selector": map[string]interface{}{
			"header.number": number,
		},
	}

	var _result map[string]interface{}
	_, err := resty.R().
		SetBody(query).
		SetResult(&_result).
		SetPathParams(map[string]string{"db": s.dbName}).
		Post(s.addr + "/{db}/_find")
	if err != nil {
		return false, fmt.Errorf("failed to get block. %s", err)
	}

	if docs, ok := _result["docs"].([]interface{}); ok && len(docs) > 0 {
		return true, nil
	}

	return false, nil
}

func (s *Store) getBlock(number int64, fields []string, result interface{}) error {

	if fields == nil {
		fields = []string{}
	}

	var query map[string]interface{}

	if number != -1 {
		query = map[string]interface{}{
			"limit": 1,
			"selector": map[string]interface{}{
				"header.number": number,
				"type":          types.TypeBlock,
			},
		}
	} else {
		query = map[string]interface{}{
			"limit": 1,
			"selector": map[string]interface{}{
				"type": types.TypeBlock,
			},
			"sort": []map[string]string{
				{"header.number": "desc"},
			},
		}
	}

	query["fields"] = fields

	var _result map[string]interface{}
	resp, err := resty.R().
		SetBody(query).
		SetResult(&_result).
		SetPathParams(map[string]string{"db": s.dbName}).
		Post(s.addr + "/{db}/_find")
	if err != nil {
		return fmt.Errorf("failed to get block. %s", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to get block. %s", resp.String())
	}

	docs, ok := _result["docs"].([]interface{})
	if ok && len(docs) == 0 {
		return types.ErrBlockNotFound
	}

	if ok && len(docs) > 0 {
		mapstructure.Decode(docs[0], &result)
	}

	return nil
}

// GetBlockHeader the header of the current block
func (s *Store) GetBlockHeader(number int64, header *wire.Header) error {

	var block map[string]interface{}
	if err := s.getBlock(number, []string{"header"}, &block); err != nil {
		return err
	}

	return mapstructure.Decode(block["header"], header)
}

// GetBlock fetches a block by its block number.
// If the block number begins with -1, the block with the highest block number is returned.
func (s *Store) GetBlock(number int64, result types.Block) error {
	return s.getBlock(number, nil, result)
}

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *Store) PutBlock(block types.Block) error {

	
	exist, err := s.hasBlock(block.GetNumber())
	if err != nil {
		return fmt.Errorf("failed to check if block exists")
	} else if exist {
		return fmt.Errorf("block with same number already exists")
	}

	var blockM = util.StructToMap(block)
	blockM["type"] = types.TypeBlock

	resp, err := resty.R().
		SetBody(blockM).
		SetPathParams(map[string]string{"db": s.dbName}).
		SetPathParams(map[string]string{"docid": util.RandString(32)}).
		Put(s.addr + "/{db}/{docid}")
	if err != nil {
		return fmt.Errorf("failed to store block. %s", err)
	}

	if resp.StatusCode() != 201 {
		return fmt.Errorf("failed to store block. %s", resp.String())
	}

	return nil
}

// Close closes the store
func (s *Store) Close() error {
	return nil
}
