package node

import (
	"math/big"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func BlockTest() bool {
	return Describe("Block", func() {

		var lp, rp *Node
		var err error
		var lpGossip, rpGossip *Gossip

		BeforeEach(func() {
			err := lpBc.Up()
			Expect(err).To(BeNil())
			err = rpBc.Up()
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			lp, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(4), log)
			Expect(err).To(BeNil())
			lpGossip = NewGossip(lp, log)
			lp.SetGossipProtocol(lpGossip)
			lp.SetBlockchain(lpBc)
			lp.SetProtocolHandler(config.RequestBlockVersion, lpGossip.OnRequestBlock)
		})

		BeforeEach(func() {
			rp, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(5), log)
			Expect(err).To(BeNil())
			rpGossip = NewGossip(rp, log)
			rp.SetGossipProtocol(rpGossip)
			rp.SetProtocolHandler(config.BlockVersion, rpGossip.OnBlock)
			rp.SetProtocolHandler(config.GetBlockHeaders, rpGossip.OnGetBlockHeaders)
			rp.SetBlockchain(rpBc)
		})

		AfterEach(func() {
			lp.Host().Close()
			rp.Host().Close()
		})

		Describe(".RelayBlock", func() {

			var block core.Block

			Context("on success", func() {
				BeforeEach(func() {
					block, err = lpBc.Generate(&core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
					Expect(err).To(BeNil())
				})

				It("should relay block to remote peer", func() {
					rpCurBlock, err := rpBc.ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(rpCurBlock.GetNumber()).To(Equal(uint64(1)))

					err = lpGossip.RelayBlock(block, []types.Engine{rp})
					Expect(err).To(BeNil())

					time.Sleep(10 * time.Millisecond)
					rpCurBlock, err = rpBc.ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(rpCurBlock.GetNumber()).To(Equal(block.GetNumber()))
				})

				It("should emit core.EventNewBlock", func() {
					err = lpGossip.RelayBlock(block, []types.Engine{rp})
					Expect(err).To(BeNil())
					evt := <-rpBc.GetEventEmitter().Once(core.EventNewBlock)
					Expect(evt.Args[0].(core.Block).GetNumber()).To(Equal(block.GetNumber()))
				})
			})
		})

		Describe(".RelayBlock 2", func() {

			Context("with multiple blocks", func() {

				var block2, block3 core.Block

				BeforeEach(func() {
					block2, err = lpBc.Generate(&core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730725),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
					Expect(err).To(BeNil())
					_, err = lpBc.ProcessBlock(block2)
					Expect(err).To(BeNil())

					block3, err = lpBc.Generate(&core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730726),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
					Expect(err).To(BeNil())
					_, err = lpBc.ProcessBlock(block3)
					Expect(err).To(BeNil())
				})

				It("should emit core.EventOrphanBlock", func() {
					err = lpGossip.RelayBlock(block3, []types.Engine{rp})
					Expect(err).To(BeNil())

					evt := <-rpBc.GetEventEmitter().Once(core.EventOrphanBlock)
					orphanBlock := evt.Args[0].(*objects.Block)

					Expect(orphanBlock.GetNumber()).To(Equal(block3.GetNumber()))
					Expect(orphanBlock.Broadcaster.StringID()).To(Equal(lp.StringID()))
					Expect(rpBc.OrphanBlocks().Len()).To(Equal(1))

					Describe("", func() {
						err = rpGossip.RequestBlock(lp, orphanBlock.Header.ParentHash)
						Expect(err).To(BeNil())
						time.Sleep(10 * time.Millisecond)

						Describe("orphan block must no longer be in the orphan cache", func() {
							Expect(rpBc.OrphanBlocks().Len()).To(Equal(0))
						})

						Describe("current block must be the previously orphaned block", func() {
							curBlock, err := rpBc.ChainReader().Current()
							Expect(err).To(BeNil())
							Expect(curBlock.GetHash()).To(Equal(orphanBlock.GetHash()))
						})
					})
				})
			})
		})

		Describe(".SendGetBlockHeaders", func() {

			// add 2 more blocks to the remote peer's
			// blockchain.
			// Target shape:
			// [1]-[2]-[3]
			BeforeEach(func() {
				block2, err := rpBc.Generate(&core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				Expect(err).To(BeNil())
				rpBc.ProcessBlock(block2)

				block3, err := rpBc.Generate(&core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730725),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				Expect(err).To(BeNil())
				rpBc.ProcessBlock(block3)
			})

			It("should successfully send message", func() {
				err := lpGossip.SendGetBlockHeaders(rp)
				Expect(err).To(BeNil())
			})
		})
	})
}
