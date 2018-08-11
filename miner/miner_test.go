package miner

import (
	"time"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/miner/ethash"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var MinerTest = func() bool {

	return Describe("Miner", func() {

		var miner *Miner

		BeforeEach(func() {
			cfg.Miner.Mode = ethash.ModeTest
			miner = New(sender, bc, event, cfg, log)
		})

		Describe(".getProposedBlock", func() {
			It("should get a block", func() {
				b, err := miner.getProposedBlock([]*wire.Transaction{
					wire.NewTx(wire.TxTypeBalance, 124, util.String(miner.minerKey.Addr()), miner.minerKey, "0.1", "0.1", time.Now().Unix()),
				})
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

		Describe(".Mine", func() {

			var newBlock *wire.Block

			BeforeEach(func() {
				newBlock, err = miner.getProposedBlock([]*wire.Transaction{
					wire.NewTx(wire.TxTypeBalance, 125, util.String(miner.minerKey.Addr()), miner.minerKey, "0.1", "0.1", time.Now().Unix()),
				})
				Expect(err).To(BeNil())
			})

			It("should abort when a new block has been found", func() {
				cfg.Miner.Mode = ethash.ModeFake
				miner = New(sender, bc, event, cfg, log)
				miner.setFakeDelay(2 * time.Second)
				go func() {
					for range miner.event.On(EventAborted) {
						miner.Stop()
					}
				}()

				time.AfterFunc(1*time.Second, func() {
					miner.event.Emit(common.EventNewBlock, newBlock)
				})

				miner.Mine()
			})
		})
	})
}
