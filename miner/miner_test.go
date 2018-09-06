package miner

import (
	"math/rand"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/miner/blakimoto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var MinerTest = func() bool {

	return Describe("Miner", func() {

		var miner *Miner

		BeforeEach(func() {
			cfg.Node.Mode = config.ModeDev
		})

		BeforeEach(func() {
			cfg.Miner.Mode = blakimoto.ModeTest
			miner = New(sender, bc, event, cfg, log)
		})

		Describe(".getProposedBlock", func() {
			It("should get a block", func() {
				b, err := miner.getProposedBlock([]core.Transaction{
					objects.NewTx(objects.TxTypeBalance, uint64(rand.Intn(100)), util.String(miner.minerKey.Addr()), miner.minerKey, "0.1", "0.1", time.Now().Unix()),
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

		// Describe(".Mine", func() {

		// 	var newBlock core.Block

		// 	BeforeEach(func() {
		// 		newBlock, err = miner.getProposedBlock([]core.Transaction{
		// 			wire.NewTx(wire.TxTypeBalance, 125, util.String(miner.minerKey.Addr()), miner.minerKey, "0.1", "0.1", time.Now().Unix()),
		// 		})
		// 		Expect(err).To(BeNil())
		// 	})

		// 	It("should abort when a new block has been found", func() {
		// 		cfg.Miner.Mode = blakimoto.ModeNormal
		// 		miner = New(sender, bc, event, cfg, log)
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
		// 		// miner.Mine()
		// 	})
		// })
	})
}
