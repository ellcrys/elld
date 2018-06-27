package couchdb

import (
	"fmt"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/util"
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

	resp, err := resty.R().SetPathParams(map[string]string{"db": s.dbName}).Get(s.addr + "/{db}")
	if err != nil {
		return fmt.Errorf("failed to check database existence. %s", err)
	}

	switch resp.StatusCode() {
	case 404:
		resp, err = resty.R().SetPathParams(map[string]string{"db": s.dbName}).Put(s.addr + "/{db}")
		if err != nil {
			return fmt.Errorf("failed to create database. %s", resp.String())
		}
		if resp.StatusCode() != 201 {
			return fmt.Errorf("failed to create database. %s", resp.String())
		}
	case 200:
		return nil
	default:
		return fmt.Errorf("failed to check database existence. %s", resp.String())
	}

	return nil
}

func (s *Store) deleteDB() error {
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

// GetBlock fetches a block by its block number.
func (s *Store) GetBlock(number uint64, result types.Block) error {

	query := map[string]interface{}{
		"limit": 1,
		"selector": map[string]interface{}{
			"header.number": number,
			"type":          types.TypeBlock,
		},
	}

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

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *Store) PutBlock(block types.Block) error {

	exist, err := s.hasBlock(block.GetNumber())
	if err != nil {
		return fmt.Errorf("failed to check if block exists")
	} else if exist {
		return fmt.Errorf("block with same number already exists")
	}

	blockM := structs.Map(block)
	blockM["type"] = types.TypeBlock

	resp, err := resty.R().
		SetBody(block).
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
