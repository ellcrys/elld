package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/ellcrys/elld/blockchain/couchdb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockchain", func() {

	var testDB = "testdb"
	var couchDBAddr = "http://127.0.0.1:5984/"
	var err error
	var store types.Store
	var chain *Chain

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store, err = couchdb.New(testDB, couchDBAddr)
		Expect(err).To(BeNil())
		err = store.Initialize()
		Expect(err).To(BeNil())
		chain = NewChain(store, cfg, log)
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	AfterEach(func() {
		err = store.DropDB()
		Expect(err).To(BeNil())
	})

	Describe(".getMatureTickets", func() {

		BeforeEach(func() {
			for _, b := range testdata.TestBlocks {
				var block wire.Block
				err := json.Unmarshal([]byte(b), &block)
				Expect(err).To(BeNil())
				err = store.PutBlock(&block)
				Expect(err).To(BeNil())
			}
		})

		It("", func() {
			ticketTxs, err := chain.getMatureTickets(4)
			Expect(err).To(BeNil())
			fmt.Println(ticketTxs)
		})
	})
})
