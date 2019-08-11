package ticket_test

import (
	"fmt"
	"os"

	"github.com/ellcrys/elld/testutil"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/ticket"
	"github.com/ellcrys/elld/crypto"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {

	var mgr *ticket.Manager

	BeforeEach(func() {
		mgr = ticket.NewManager(nil)
	})

	Describe(".DetermineTerm", func() {
		ticket.BlocksPerTerm = 10
		var cases = [][]uint{{1, uint(1)}, {20, uint(2)}, {201, uint(21)}}
		for _, c := range cases {
			It(fmt.Sprintf("should return term=%d when block number = %d", c[1], c[0]), func() {
				Expect(mgr.DetermineTerm(uint64(c[0]))).To(Equal(c[1]))
			})
		}
	})

	Context("With an initialized chain", func() {
		var err error
		var bc *blockchain.Blockchain
		var cfg *config.EngineConfig
		var db elldb.DB
		var genesisBlock types.Block
		var genesisChain *blockchain.Chain
		var nodeKey, sender, receiver *crypto.Key

		BeforeEach(func() {
			cfg, err = testutil.SetTestCfg()
			Expect(err).To(BeNil())

			db = elldb.NewDB(cfg.NetDataDir())
			err = db.Open(util.RandString(5))
			Expect(err).To(BeNil())

			sender = crypto.NewKeyFromIntSeed(1)
			receiver = crypto.NewKeyFromIntSeed(2)

			nodeKey = sender
			bc = blockchain.New(txpool.New(100), cfg, log)
			bc.SetDB(db)
			bc.SetNodeKey(nodeKey)
		})

		BeforeEach(func() {
			genesisBlock, err = blockchain.LoadBlockFromFile("genesis-test.json")
			Expect(err).To(BeNil())
			bc.SetGenesisBlock(genesisBlock)
			err = bc.Up()
			Expect(err).To(BeNil())
			genesisChain = bc.GetBestChain().(*blockchain.Chain)
		})

		AfterEach(func() {
			db.Close()
			err = os.RemoveAll(cfg.DataDir())
			Expect(err).To(BeNil())
		})

		Describe(".DetermineCurrentTerm", func() {

			BeforeEach(func() {
				mgr = ticket.NewManager(bc)
			})

			When("only genesis block (block number 1) exist", func() {
				It("should return term = 1", func() {
					term, err := mgr.DetermineCurrentTerm()
					Expect(err).To(BeNil())
					Expect(term).To(Equal(uint(1)))
				})
			})

			When("block number = 2 and BlocksPerTerm = 1", func() {
				var block types.Block

				BeforeEach(func() {
					ticket.BlocksPerTerm = 1
					block = MakeBlock(bc, genesisChain, sender, receiver)
					err = genesisChain.Append(block)
					Expect(err).To(BeNil())
				})

				It("should return term = 2", func() {
					term, err := mgr.DetermineCurrentTerm()
					Expect(err).To(BeNil())
					Expect(term).To(Equal(uint(2)))
				})
			})
		})
	})
})
