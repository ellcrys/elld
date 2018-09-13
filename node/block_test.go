package node_test

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Block", func() {

	var err error
	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
	var receiver, _ = crypto.NewKey(nil)

	BeforeEach(func() {
		lp = makeTestNode(30000)
		Expect(lp.GetBlockchain().Up()).To(BeNil())
		lp.SetProtocolHandler(config.RequestBlockVersion, lp.Gossip().OnRequestBlock)

		rp = makeTestNode(30001)
		Expect(rp.GetBlockchain().Up()).To(BeNil())
		rp.SetProtocolHandler(config.BlockBodyVersion, rp.Gossip().OnBlockBody)
		rp.SetProtocolHandler(config.GetBlockHashesVersion, rp.Gossip().OnGetBlockHashes)
		rp.SetProtocolHandler(config.GetBlockBodiesVersion, rp.Gossip().OnGetBlockBodies)

		// Create sender account on the remote peer
		Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &objects.Account{
			Type:    objects.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "100",
		})).To(BeNil())

		// Create sender account on the local peer
		Expect(lp.GetBlockchain().CreateAccount(1, lp.GetBlockchain().GetBestChain(), &objects.Account{
			Type:    objects.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "100",
		})).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".RelayBlock", func() {

		Context("when block is successfully relayed to a remote peer", func() {

			var evtArgs emitter.Event
			var block core.Block

			BeforeEach(func(done Done) {
				block = makeBlock(lp.GetBlockchain(), sender, sender, time.Now().Unix())
				go func() {
					defer GinkgoRecover()
					err = lp.Gossip().RelayBlock(block, []types.Engine{rp})
					Expect(err).To(BeNil())
				}()
				evtArgs = <-rp.GetEventEmitter().On(node.EventBlockProcessed)
				close(done)
			})

			Specify("relayed block must be processed by the remote peer", func() {
				Expect(evtArgs.Args).To(HaveLen(2))
				processedBlock, err := evtArgs.Args[0], evtArgs.Args[1]
				Expect(err).To(BeNil())
				Expect(processedBlock).ToNot(BeNil())
				Expect(block.HashToHex()).To(Equal(processedBlock.(core.Block).HashToHex()))

				rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())
				Expect(rpCurBlock.GetNumber()).To(Equal(block.GetNumber()))
			})
		})

		Context("when relayed block is considered an orphan by the remote peer", func() {

			var block2, block3 core.Block
			var evtOrphanBlock emitter.Event

			BeforeEach(func(done Done) {
				block2 = makeBlock(lp.GetBlockchain(), sender, sender, time.Now().Unix()-1)
				_, err = lp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				block3 = makeBlock(lp.GetBlockchain(), sender, sender, time.Now().Unix())
				_, err = lp.GetBlockchain().ProcessBlock(block3)
				Expect(err).To(BeNil())

				go func() {
					defer GinkgoRecover()
					err = lp.Gossip().RelayBlock(block3, []types.Engine{rp})
					Expect(err).To(BeNil())
					evtOrphanBlock = <-rp.GetBlockchain().GetEventEmitter().On(core.EventOrphanBlock)
				}()
				<-rp.GetEventEmitter().On(node.EventBlockProcessed)
				close(done)
			})

			It("should emit core.EventOrphanBlock", func() {
				Expect(evtOrphanBlock).ToNot(BeNil())
				Expect(evtOrphanBlock.Args).To(HaveLen(1))
				orphanBlock := evtOrphanBlock.Args[0].(*objects.Block)
				Expect(orphanBlock.GetNumber()).To(Equal(block3.GetNumber()))
				Expect(orphanBlock.GetBroadcaster().StringID()).To(Equal(lp.StringID()))
				Expect(rp.GetBlockchain().OrphanBlocks().Len()).To(Equal(1))
			})
		})
	})

	Describe(".RequestBlock", func() {

		var block2 core.Block
		var evtArgs emitter.Event

		BeforeEach(func(done Done) {
			block2 = makeBlock(lp.GetBlockchain(), sender, sender, time.Now().Unix())
			_, err := lp.GetBlockchain().ProcessBlock(block2)
			Expect(err).To(BeNil())

			go func() {
				defer GinkgoRecover()
				err = rp.Gossip().RequestBlock(lp, block2.GetHash())
				Expect(err).To(BeNil())
			}()
			evtArgs = <-rp.GetEventEmitter().On(node.EventBlockProcessed)
			close(done)
		})

		It("Should request and process block from remote peer", func() {
			Expect(evtArgs.Args).To(HaveLen(2))
			processedBlock, err := evtArgs.Args[0], evtArgs.Args[1]
			Expect(err).To(BeNil())
			Expect(processedBlock).ToNot(BeNil())
			Expect(block2.HashToHex()).To(Equal(processedBlock.(core.Block).HashToHex()))
			curBlock, err := rp.GetBlockchain().ChainReader().Current()
			Expect(err).To(BeNil())
			Expect(curBlock.HashToHex()).To(Equal(block2.HashToHex()))
		})
	})

	Describe(".SendGetBlockHashes", func() {

		var block2, block3 core.Block

		// Target shape:
		// Remote Peer
		// [1]-[2]-[3]
		//
		// Local Peer
		// [1]
		Context("when remote blockchain shape is [1]-[2]-[3] and local blockchain shape: [1]", func() {

			BeforeEach(func(done Done) {
				block2 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix()-1)
				_, err := rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				block3 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix())
				_, err = rp.GetBlockchain().ProcessBlock(block3)
				Expect(err).To(BeNil())

				go func() {
					defer GinkgoRecover()
					err := lp.Gossip().SendGetBlockHashes(rp, util.Hash{})
					Expect(err).To(BeNil())
				}()

				<-lp.GetEventEmitter().On(node.EventReceivedBlockHashes)
				close(done)
			})

			It("should get 2 block hashes from remote peer", func() {
				Expect(lp.GetBlockHashQueue().Size()).To(Equal(2))
			})

			Specify("first header number = 2 and second header number = 3", func() {
				h2 := lp.GetBlockHashQueue().Shift()
				h3 := lp.GetBlockHashQueue().Shift()
				Expect(h2).To(BeAssignableToTypeOf(&node.BlockHash{}))
				Expect(h2.(*node.BlockHash).Hash).To(Equal(block2.GetHash()))
				Expect(h3).To(BeAssignableToTypeOf(&node.BlockHash{}))
				Expect(h3.(*node.BlockHash).Hash).To(Equal(block3.GetHash()))
			})
		})

		// Target shape:
		// Remote Peer
		// [1]-[2]
		//
		// Local Peer
		// [1]
		Context("when remote blockchain shape is [1]-[2] and local peer blockchain shape is [1]", func() {

			var block2 core.Block

			BeforeEach(func(done Done) {
				block2 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix()-1)
				_, err := rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				go func() {
					err = lp.Gossip().SendGetBlockHashes(rp, util.Hash{})
				}()
				<-lp.GetEventEmitter().On(node.EventReceivedBlockHashes)
				Expect(err).To(BeNil())
				close(done)
			})

			It("should get 1 block hash from remote peer", func() {
				Expect(lp.GetBlockHashQueue().Size()).To(Equal(1))
			})

			Specify("the block hash must match hash of block [2]", func() {
				hash := lp.GetBlockHashQueue().Shift()
				Expect(hash).To(BeAssignableToTypeOf(&node.BlockHash{}))
				Expect(hash.(*node.BlockHash).Hash).To(Equal(block2.GetHash()))
			})
		})

		// Target shape:
		// Remote Peer
		// [1]
		//
		// Local Peer
		// [1]
		Context("when remote peer's blockchain shape is [1] and local peer's blockchain shape is [1]", func() {

			BeforeEach(func(done Done) {
				go func() {
					err = lp.Gossip().SendGetBlockHashes(rp, util.Hash{})
				}()
				<-lp.GetEventEmitter().On(node.EventReceivedBlockHashes)
				Expect(err).To(BeNil())
				close(done)
			})

			Specify("local peer block hash queue should be empty", func() {
				Expect(lp.GetBlockHashQueue().Size()).To(Equal(0))
			})
		})

		// Target shape:
		// Remote Peer
		// [1]-[2]-[3] 	ChainA
		//  |__[2]		ChainB
		Context("when remote peer's blockchain shape is: [1]-[2] and local peer's blockchain shape is: [1]", func() {

			var block2, block3, chainBBlock2 core.Block

			BeforeEach(func() {
				block2 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix()-10)
				chainBBlock2 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix()-9)

				_, err = rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				// processing chainBBlock2 will create a fork
				_, err = rp.GetBlockchain().ProcessBlock(chainBBlock2)
				Expect(err).To(BeNil())
				Expect(rp.GetBlockchain().GetChainsReader()).To(HaveLen(2))

				block3 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix())
				_, err = rp.GetBlockchain().ProcessBlock(block3)
				Expect(err).To(BeNil())
			})

			When("locator hash is block [2] of Chain B", func() {

				BeforeEach(func(done Done) {
					go func() {
						err = lp.Gossip().SendGetBlockHashes(rp, chainBBlock2.GetHash())
					}()
					<-lp.GetEventEmitter().On(node.EventReceivedBlockHashes)
					Expect(err).To(BeNil())
					close(done)
				})

				Specify("the block hash queue should contain 2 hashes from ChainA", func() {
					Expect(lp.GetBlockHashQueue().Size()).To(Equal(2))
					Context("first header number = [2] and second header number = [3]", func() {
						h2 := lp.GetBlockHashQueue().Shift()
						h3 := lp.GetBlockHashQueue().Shift()
						Expect(h2).To(BeAssignableToTypeOf(&node.BlockHash{}))
						Expect(h2.(*node.BlockHash).Hash).To(Equal(block2.GetHash()))
						Expect(h3).To(BeAssignableToTypeOf(&node.BlockHash{}))
						Expect(h3.(*node.BlockHash).Hash).To(Equal(block3.GetHash()))
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
		Context("when remote peer's blockchain shape is [1] and local peer's blockchain shape is: [1]", func() {

			BeforeEach(func(done Done) {
				go func() {
					err = lp.Gossip().SendGetBlockHashes(rp, util.StrToHash("unknown"))
				}()
				<-lp.GetEventEmitter().On(node.EventReceivedBlockHashes)
				Expect(err).To(BeNil())
				close(done)
			})

			It("local peer's block hash queue should be empty", func() {
				Expect(lp.GetBlockHashQueue().Size()).To(Equal(0))
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
		Context("when remote peer's blockchain shape is [1]-[2]-[3] and local peer's blockchain shape is: [1]", func() {

			Context("at least one hash is requested", func() {
				BeforeEach(func(done Done) {
					block2 = makeBlock(rp.GetBlockchain(), sender, sender, time.Now().Unix()-1)
					_, err := rp.GetBlockchain().ProcessBlock(block2)
					Expect(err).To(BeNil())

					block3 = makeBlock(rp.GetBlockchain(), sender, sender, time.Now().Unix())
					_, err = rp.GetBlockchain().ProcessBlock(block3)
					Expect(err).To(BeNil())

					go func() {
						defer GinkgoRecover()
						hashes := []util.Hash{block2.GetHash(), block3.GetHash()}
						err = lp.Gossip().SendGetBlockBodies(rp, hashes)
						Expect(err).To(BeNil())
					}()
					<-lp.GetEventEmitter().On(node.EventBlockBodiesProcessed)
					close(done)
				})

				It("should successfully fetch block [2] and [3] and append to local peer's chain", func() {
					lpTip, err := lp.GetBlockchain().ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(lpTip.GetNumber()).To(Equal(uint64(3)))
				})
			})

			Context("no hash is requested", func() {
				BeforeEach(func() {
					err = lp.Gossip().SendGetBlockBodies(rp, []util.Hash{})
					Expect(err).To(BeNil())
				})

				Specify("local chain tip should remain unchanged", func() {
					lpTip, err := lp.GetBlockchain().ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(lpTip.GetNumber()).To(Equal(uint64(1)))
				})
			})
		})
	})
})
