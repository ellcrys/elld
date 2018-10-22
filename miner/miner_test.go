package miner

import (
	"os"
	"time"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/miner/blakimoto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Account", func() {

	var err error
	var bc *blockchain.Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock core.Block
	var sender *crypto.Key

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.DataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)

		bc = blockchain.New(txpool.New(100), cfg, log)
		bc.SetDB(db)
	})

	BeforeEach(func() {
		genesisBlock, err = blockchain.LoadBlockFromFile("genesis-test.json")
		Expect(err).To(BeNil())
		bc.SetGenesisBlock(genesisBlock)
		err = bc.Up()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe("Miner", func() {

		var miner *Miner

		BeforeEach(func() {
			cfg.Node.Mode = config.ModeDev
		})

		BeforeEach(func() {
			cfg.Miner.Mode = blakimoto.ModeTest
			miner = New(sender, bc, bc.GetEventEmitter(), cfg, log)
		})

		Describe(".getProposedBlock", func() {
			It("should get a block", func() {
				b, err := miner.getProposedBlock(nil)
				Expect(err).To(BeNil())
				Expect(b).ToNot(BeNil())
			})
		})

		Describe(".Stop", func() {
			It("should stop miner", func() {
				time.AfterFunc(1*time.Second, func() {
					defer GinkgoRecover()
					miner.Stop()
					Expect(miner.stop).To(BeTrue())
				})
				miner.Mine()
			})
		})

		// Describe(".Mine", func() {

		// 	var newBlock core.Block

		// 	BeforeEach(func() {
		// 		newBlock, err = miner.getProposedBlock([]core.Transaction{
		// 			objects.NewTx(objects.TxTypeBalance, 125, util.String(miner.minerKey.Addr()), miner.minerKey, "0.1", "0.1", time.Now().Unix()),
		// 		})
		// 		Expect(err).To(BeNil())
		// 	})

		// 	It("should abort when a new block has been found", func() {
		// 		cfg.Miner.Mode = blakimoto.ModeNormal
		// 		miner.setFakeDelay(2 * time.Second)
		// 		// go func() {
		// 		// 	for range miner.event.On(EventAborted) {
		// 		// 		miner.Stop()
		// 		// 	}
		// 		// }()

		// 		// time.AfterFunc(1*time.Second, func() {
		// 		// 	miner.event.Emit(common.EventNewBlock, newBlock)
		// 		// })
		// 		// _ = newBlock
		// 		miner.Mine()
		// 	})
		// })
	})
})
