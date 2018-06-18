package couchdb

import (
	"strings"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Couchdb", func() {

	var err error
	var addr = "http://127.0.0.1:5984"
	var dbName = strings.ToLower(util.RandString(5))
	var store *Store

	BeforeEach(func() {
		store, err = New(dbName, addr)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err = store.deleteDB()
		Expect(err).To(BeNil())
	})

	Describe(".New", func() {
		It("should return error when unable to reach to database server", func() {
			_, err := New("mydb", "http://127.0.0.1:5982")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to reach database at specified address. Get http://127.0.0.1:5982/: dial tcp 127.0.0.1:5982: connect: connection refused"))
		})

		It("should return no error if able to reach database server", func() {
			store, err := New("mydb", addr)
			Expect(err).To(BeNil())
			Expect(store).ToNot(BeNil())
			Expect(store.dbName).To(Equal("mydb"))
			Expect(store.addr).To(Equal(addr))
		})
	})

	Describe(".Initialize", func() {
		It("should return error if unable to create database", func() {
			store, err := New("mydb2*&^*.,", addr)
			Expect(err).To(BeNil())

			err = store.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(`illegal_database_name`))
		})

		It("should successfully create the database", func() {
			store, err := New(dbName, addr)
			Expect(err).To(BeNil())
			err = store.Initialize()
			Expect(err).To(BeNil())
		})
	})

	Describe(".PutBlock", func() {

		BeforeEach(func() {
			err := store.Initialize()
			Expect(err).To(BeNil())
		})

		It("should successfully put block", func() {
			err := store.PutBlock(&blockchain.Block{
				DocType: blockchain.TypeBlock,
				Header: &blockchain.Header{
					Number: 1,
				},
			})
			Expect(err).To(BeNil())
		})

		It("should return err if block with same number exists", func() {
			block := &blockchain.Block{
				DocType: blockchain.TypeBlock,
				Header: &blockchain.Header{
					Number: 1,
				},
			}
			err := store.PutBlock(block)
			Expect(err).To(BeNil())
			err = store.PutBlock(block)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("block with same number already exists"))
		})
	})

	Describe(".GetBlock", func() {

		BeforeEach(func() {
			err := store.Initialize()
			Expect(err).To(BeNil())
		})

		It("should successfully get block", func() {

			var block = &blockchain.Block{
				DocType: blockchain.TypeBlock,
				Header: &blockchain.Header{
					Number: 1,
				},
			}

			err := store.PutBlock(block)
			Expect(err).To(BeNil())

			var result blockchain.Block
			err = store.GetBlock(1, &result)
			Expect(err).To(BeNil())
			Expect(&result).To(Equal(block))
		})

		It("should return err = 'block not found' if block does not exist", func() {
			var result blockchain.Block
			err = store.GetBlock(1, &result)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(types.ErrBlockNotFound))
		})

	})
})
