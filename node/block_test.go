package node

import (
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
			rp.SetProtocolHandler(config.BlockBodyVersion, rpGossip.OnBlockBody)
			rp.SetProtocolHandler(config.GetBlockHashes, rpGossip.OnGetBlockHashes)
			rp.SetProtocolHandler(config.GetBlockBodies, rpGossip.OnGetBlockBodies)
			rp.SetBlockchain(rpBc)
		})

		AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		Describe(".RelayBlock", func() {

			var block core.Block

			Context("on success", func() {

				BeforeEach(func() {
					block = makeBlock(rpBc)
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
					block2 = makeBlock(lpBc)
					_, err = lpBc.ProcessBlock(block2)
					Expect(err).To(BeNil())

					block3 = makeBlock(lpBc)
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
						time.Sleep(50 * time.Millisecond)

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

			var block2, block3 core.Block

			// Target shape:
			// Remote Peer
			// [1]-[2]-[3]
			//
			// Local Peer
			// [1]
			Context("Remote Blockchain: [1]-[2]-[3] and Local Peer: [1]", func() {
				BeforeEach(func() {
					block2 = makeBlock(rpBc)
					Expect(err).To(BeNil())
					rpBc.ProcessBlock(block2)

					block3 = makeBlock(rpBc)
					Expect(err).To(BeNil())
					rpBc.ProcessBlock(block3)
				})

				It("should successfully send message", func() {
					err := lpGossip.SendGetBlockHashes(rp, util.Hash{})
					Expect(err).To(BeNil())
					Context("header queue must contain 2 headers", func() {
						Expect(lp.blockHashQueue.Size()).To(Equal(2))
					})

					Context("first header number = 2 and second header number = 3", func() {
						h2 := lp.blockHashQueue.Shift()
						h3 := lp.blockHashQueue.Shift()
						Expect(h2).To(BeAssignableToTypeOf(&BlockHash{}))
						Expect(h2.(*BlockHash).Hash).To(Equal(block2.GetHash()))
						Expect(h3).To(BeAssignableToTypeOf(&BlockHash{}))
						Expect(h3.(*BlockHash).Hash).To(Equal(block3.GetHash()))
					})
				})
			})

			// Target shape:
			// Remote Peer
			// [1]-[2]
			//
			// Local Peer
			// [1]
			Context("Remote Blockchain: [1]-[2] and Local Peer: [1]", func() {

				var block2 core.Block

				BeforeEach(func() {
					block2 = makeBlock(rpBc)
					Expect(err).To(BeNil())
					rpBc.ProcessBlock(block2)
				})

				It("should successfully send message", func() {
					err := lpGossip.SendGetBlockHashes(rp, util.Hash{})
					Expect(err).To(BeNil())
					Context("header queue must contain 2 headers", func() {
						Expect(lp.blockHashQueue.Size()).To(Equal(1))
					})
					Context("first header number = 2", func() {
						h2 := lp.blockHashQueue.Shift()
						Expect(h2).To(BeAssignableToTypeOf(&BlockHash{}))
						Expect(h2.(*BlockHash).Hash).To(Equal(block2.GetHash()))
					})
				})
			})

			// Target shape:
			// Remote Peer
			// [1]
			//
			// Local Peer
			// [1]
			Context("Remote Blockchain: [1] and Local Peer: [1]", func() {
				It("should successfully send message", func() {
					err := lpGossip.SendGetBlockHashes(rp, util.Hash{})
					Expect(err).To(BeNil())
					Context("header queue must contain 0 headers", func() {
						Expect(lp.blockHashQueue.Size()).To(Equal(0))
					})
				})
			})

			// Target shape:
			// Remote Peer
			// [1]-[2]-[3] 	ChainA
			//  |__[2]		ChainB
			Context("Remote Blockchain: [1]-[2] and Local Peer: [1]", func() {

				var block2, block3, chainBBlock2 core.Block

				BeforeEach(func() {
					block2 = makeBlock(rpBc)
					chainBBlock2 = makeBlock(rpBc)

					_, err = rpBc.ProcessBlock(block2)
					Expect(err).To(BeNil())

					// processing chainBBlock2 will create a fork
					_, err = rpBc.ProcessBlock(chainBBlock2)
					Expect(err).To(BeNil())
					Expect(rpBc.GetChainsReader()).To(HaveLen(2))

					block3 = makeBlock(rpBc)
					_, err = rpBc.ProcessBlock(block3)
					Expect(err).To(BeNil())
				})

				When("locator hash = [2] of Chain B", func() {
					It("should successfully send message", func() {
						err := lpGossip.SendGetBlockHashes(rp, chainBBlock2.GetHash())
						Expect(err).To(BeNil())
						Context("header queue must contain 2 headers", func() {
							Expect(lp.blockHashQueue.Size()).To(Equal(2))
						})
						Context("first header number = [2] and second header number = [3]", func() {
							h2 := lp.blockHashQueue.Shift()
							h3 := lp.blockHashQueue.Shift()
							Expect(h2).To(BeAssignableToTypeOf(&BlockHash{}))
							Expect(h2.(*BlockHash).Hash).To(Equal(block2.GetHash()))
							Expect(h3).To(BeAssignableToTypeOf(&BlockHash{}))
							Expect(h3.(*BlockHash).Hash).To(Equal(block3.GetHash()))
						})
					})
				})
			})

			// Target shape:
			// Remote Peer
			// [1]
			//
			// Local Peer
			// [1]
			Context("Remote Blockchain: [1] and Local Peer: [1]", func() {
				It("should successfully send message", func() {
					err := lpGossip.SendGetBlockHashes(rp, util.StrToHash("unknown"))
					Expect(err).To(BeNil())
					When("locator hash is not found in remote peer blockchain", func() {
						Context("header queue must contain 0 headers", func() {
							Expect(lp.blockHashQueue.Size()).To(Equal(0))
						})
					})
				})
			})
		})

		Describe(".SendGetBlockBodies", func() {

			var block2, block3 core.Block

			// Target shape:
			// Remote Peer
			// [1]-[2]-[3]
			//
			// Local Peer
			// [1]
			Context("Remote Blockchain: [1]-[2]-[3] and Local Peer: [1]", func() {
				BeforeEach(func() {
					block2 = makeBlock(rpBc)
					Expect(err).To(BeNil())
					rpBc.ProcessBlock(block2)

					block3 = makeBlock(rpBc)
					Expect(err).To(BeNil())
					rpBc.ProcessBlock(block3)
				})

				It("should successfully fetch block [2] and [3] and append to chain", func() {
					lpTip, err := lpBc.ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(lpTip.GetNumber()).To(Equal(uint64(1)))

					hashes := []util.Hash{block2.GetHash(), block3.GetHash()}
					err = lpGossip.SendGetBlockBodies(rp, hashes)
					Expect(err).To(BeNil())

					lpTip, err = lpBc.ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(lpTip.GetNumber()).To(Equal(uint64(3)))
				})

				It("should append nothing to main chain when no hash is requested", func() {
					lpTip, err := lpBc.ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(lpTip.GetNumber()).To(Equal(uint64(1)))

					hashes := []util.Hash{}
					err = lpGossip.SendGetBlockBodies(rp, hashes)
					Expect(err).To(BeNil())

					lpTip, err = lpBc.ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(lpTip.GetNumber()).To(Equal(uint64(1)))
				})
			})
		})
	})
}
