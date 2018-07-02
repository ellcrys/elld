package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockchain", func() {
	var err error
	var store types.Store
	var chain *Chain
	var db database.DB

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = database.NewLevelDB(cfg.ConfigDir())
		err = db.Open("")
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store, err = leveldb.New(db)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		chain = NewChain(store, cfg, log)
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
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
