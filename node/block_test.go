package node

import (
	"math/big"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func BlockTest() bool {
	return Describe("Block", func() {

		var lp, rp *Node
		var err error
		var lpGossip, rpGossip *Gossip

		BeforeEach(func() {
			err := bc.Up()
			Expect(err).To(BeNil())
			err = bc2.Up()
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			lp, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(4), log)
			Expect(err).To(BeNil())
			lpGossip = NewGossip(lp, log)
			lp.SetBlockchain(bc)
		})

		BeforeEach(func() {
			rp, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(5), log)
			Expect(err).To(BeNil())
			rpGossip = NewGossip(rp, log)
			rp.SetProtocolHandler(config.BlockVersion, rpGossip.OnBlock)
			rp.SetBlockchain(bc2)
		})

		AfterEach(func() {
			lp.Host().Close()
			rp.Host().Close()
		})

		Describe(".RelayBlock", func() {

			var block core.Block

			BeforeEach(func() {
				block, err = bc.Generate(&core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				Expect(err).To(BeNil())
			})

			It("should successfully relay block to remote peer", func() {
				rpCurBlock, err := bc2.ChainReader().Current()
				Expect(err).To(BeNil())
				Expect(rpCurBlock.GetNumber()).To(Equal(uint64(1)))

				err = lpGossip.RelayBlock(block, []types.Engine{rp})
				Expect(err).To(BeNil())

				time.Sleep(10 * time.Millisecond)
				rpCurBlock, err = bc2.ChainReader().Current()
				Expect(err).To(BeNil())
				Expect(rpCurBlock.GetNumber()).To(Equal(block.GetNumber()))
			})

			Context("upon successful processing, remote peer", func() {
				It("should emit core.EventNewBlock", func() {
					err = lpGossip.RelayBlock(block, []types.Engine{rp})
					Expect(err).To(BeNil())
					evt := <-bc2.GetEventEmitter().Once(core.EventNewBlock)
					Expect(evt.Args[0].(core.Block).GetNumber()).To(Equal(block.GetNumber()))
				})
			})
		})
	})
}
