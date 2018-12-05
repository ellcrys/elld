package gossip_test

import (
	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Block", func() {

	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
	var receiver, _ = crypto.NewKey(nil)
	var lpPort, rpPort int

	BeforeEach(func() {
		lpPort = getPort()
		rpPort = getPort()

		lp = makeTestNode(lpPort)
		Expect(lp.GetBlockchain().Up()).To(BeNil())

		rp = makeTestNode(rpPort)
		Expect(rp.GetBlockchain().Up()).To(BeNil())

		// Create sender account on the remote peer
		Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &core.Account{
			Type:    core.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "100",
		})).To(BeNil())

		// Create sender account on the local peer
		Expect(lp.GetBlockchain().CreateAccount(1, lp.GetBlockchain().GetBestChain(), &core.Account{
			Type:    core.AccountTypeBalance,
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

			var block types.Block

			BeforeEach(func() {
				block = MakeBlockWithSingleTx(lp.GetBlockchain(), lp.GetBlockchain().GetBestChain(), sender, sender, 1)
				bm := node.NewBlockManager(rp)
				go bm.Handle()
			})

			It("remote peer must emit core.EventProcessBlock and core.EventBlockProcessed", func(done Done) {
				wait := make(chan bool)

				err := lp.Gossip().RelayBlock(block, []core.Engine{rp})
				Expect(err).To(BeNil())

				go func() {
					defer GinkgoRecover()

					evtArgs := <-rp.GetEventEmitter().Once(core.EventProcessBlock)
					Expect(evtArgs.Args).To(HaveLen(1))
					relayed := evtArgs.Args[0].(*core.Block)
					Expect(relayed.GetHash()).To(Equal(block.GetHash()))

					evtArgs = <-rp.GetEventEmitter().Once(core.EventBlockProcessed)
					Expect(evtArgs.Args).To(HaveLen(2))
					close(wait)
				}()

				<-wait
				close(done)
			})
		})

		Context("when a block has a transaction that is not in the txs pool of the remote peer", func() {

			var block types.Block

			BeforeEach(func() {
				block = MakeBlockWithSingleTx(lp.GetBlockchain(), lp.GetBlockchain().GetBestChain(), sender, sender, 1)
				bm := node.NewBlockManager(rp)
				go bm.Handle()
			})

			It("should return error about the missing transaction in the pool", func(done Done) {
				wait := make(chan bool)

				err := lp.Gossip().RelayBlock(block, []core.Engine{rp})
				Expect(err).To(BeNil())

				go func() {
					defer GinkgoRecover()

					evtArgs := <-rp.GetEventEmitter().Once(core.EventProcessBlock)
					Expect(evtArgs.Args).To(HaveLen(1))
					relayed := evtArgs.Args[0].(*core.Block)
					Expect(relayed.GetHash()).To(Equal(block.GetHash()))

					evtArgs = <-rp.GetEventEmitter().Once(core.EventBlockProcessed)
					Expect(evtArgs.Args).To(HaveLen(2))
					err := evtArgs.Args[1]
					Expect(err.(error).Error()).To(Equal("tx:0, error:transaction does not exist in the transactions pool"))
					close(wait)
				}()

				<-wait
				close(done)
			})
		})

		Context("when a block is successfully relayed to a remote peer", func() {

			var evtArgs emitter.Event
			var block types.Block

			BeforeEach(func() {
				block = MakeBlockWithSingleTx(lp.GetBlockchain(), lp.GetBlockchain().GetBestChain(), sender, sender, 1)

				// Add the transaction to the remote node's
				// transaction pool to prevent block rejection
				// due to missing tx in the remote node's tx pool
				err := rp.GetTxPool().Put(block.GetTransactions()[0])
				Expect(err).To(BeNil())

				bm := node.NewBlockManager(rp)
				go bm.Handle()
			})

			Specify("relayed block must be processed by the remote peer", func(done Done) {
				wait := make(chan bool)

				err := lp.Gossip().RelayBlock(block, []core.Engine{rp})
				Expect(err).To(BeNil())

				go func() {
					evtArgs = <-rp.GetEventEmitter().Once(core.EventBlockProcessed)
					Expect(evtArgs.Args).To(HaveLen(2))
					processedBlock, err := evtArgs.Args[0], evtArgs.Args[1]
					Expect(err).To(BeNil())
					Expect(processedBlock).ToNot(BeNil())
					Expect(block.GetHashAsHex()).To(Equal(processedBlock.(types.Block).GetHashAsHex()))

					rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(rpCurBlock.GetNumber()).To(Equal(block.GetNumber()))

					close(wait)
				}()

				<-wait
				close(done)
			})
		})

		Context("when relayed block is considered an orphan by the remote peer", func() {

			var block2, block3 types.Block
			var evtOrphanBlock = make(chan emitter.Event)

			BeforeEach(func() {
				bm := node.NewBlockManager(rp)
				go bm.Handle()

				block2 = MakeBlockWithSingleTx(lp.GetBlockchain(), lp.GetBlockchain().GetBestChain(), sender, sender, 1)
				_, err := lp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				block3 = MakeBlockWithSingleTx(lp.GetBlockchain(), lp.GetBlockchain().GetBestChain(), sender, sender, 2)
				_, err = lp.GetBlockchain().ProcessBlock(block3)
				Expect(err).To(BeNil())
			})

			It("should emit core.EventOrphanBlock", func(done Done) {
				wait := make(chan bool)

				err := lp.Gossip().RelayBlock(block3, []core.Engine{rp})
				Expect(err).To(BeNil())

				go func() {
					evt := <-rp.GetBlockchain().GetEventEmitter().Once(core.EventOrphanBlock)
					evtOrphanBlock <- evt
				}()

				go func() {
					<-rp.GetEventEmitter().Once(core.EventBlockProcessed)

					evt := <-evtOrphanBlock
					Expect(evt).ToNot(BeNil())
					Expect(evt.Args).To(HaveLen(1))
					orphanBlock := evt.Args[0].(*core.Block)

					Expect(orphanBlock.GetNumber()).To(Equal(block3.GetNumber()))
					Expect(orphanBlock.GetBroadcaster().StringID()).To(Equal(lp.StringID()))
					Expect(rp.GetBlockchain().OrphanBlocks().Len()).To(Equal(1))
					close(wait)
				}()

				<-wait
				close(done)
			})
		})
	})

	Describe(".RequestBlock", func() {

		var block2 types.Block

		BeforeEach(func() {
			bm := node.NewBlockManager(rp)
			go bm.Handle()

			block2 = MakeBlockWithSingleTx(lp.GetBlockchain(), lp.GetBlockchain().GetBestChain(), sender, sender, 1)
			_, err := lp.GetBlockchain().ProcessBlock(block2)
			Expect(err).To(BeNil())
		})

		It("Should request and process block from remote peer", func(done Done) {
			err := rp.Gossip().RequestBlock(lp, block2.GetHash())
			Expect(err).To(BeNil())

			go func() {
				<-rp.GetEventEmitter().Once(core.EventBlockProcessed)
				curBlock, err := rp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())
				Expect(curBlock.GetHashAsHex()).To(Equal(block2.GetHashAsHex()))
				close(done)
			}()
		})
	})

	Describe(".SendGetBlockHashes", func() {
		var block2, block3 types.Block

		// Target shape:
		// Remote Peer
		// [1]-[2]-[3]
		//
		// Local Peer
		// [1]
		Context("when remote blockchain shape is [1]-[2]-[3] and local blockchain shape: [1]", func() {

			var result *core.BlockHashes

			BeforeEach(func() {
				block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
				_, err := rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				block3 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 2)
				_, err = rp.GetBlockchain().ProcessBlock(block3)
				Expect(err).To(BeNil())

				result, err = lp.Gossip().SendGetBlockHashes(rp, nil, util.Hash{})
				Expect(err).To(BeNil())
			})

			It("should get 2 block hashes from remote peer", func() {
				Expect(len(result.Hashes)).To(Equal(2))
			})

			Specify("first header number = 2 and second header number = 3", func() {
				Expect(len(result.Hashes)).To(Equal(2))
				h2 := result.Hashes[0]
				h3 := result.Hashes[1]
				Expect(h2).To(Equal(block2.GetHash()))
				Expect(h3).To(Equal(block3.GetHash()))
			})
		})

		When("a valid seek hash is provided", func() {

			// Target shape:
			// Remote Peer
			// [1]-[2]-[3]
			//
			// Local Peer
			// [1]
			Context("when remote blockchain shape is [1]-[2]-[3] and local blockchain shape: [1]", func() {

				var result *core.BlockHashes

				BeforeEach(func() {
					block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
					_, err := rp.GetBlockchain().ProcessBlock(block2)
					Expect(err).To(BeNil())

					block3 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 2)
					_, err = rp.GetBlockchain().ProcessBlock(block3)
					Expect(err).To(BeNil())

					result, err = lp.Gossip().SendGetBlockHashes(rp, nil, block2.GetHash())
					Expect(err).To(BeNil())
				})

				It("should get 1 block hash from remote peer", func() {
					Expect(len(result.Hashes)).To(Equal(1))
				})

				Specify("returned hash should equal block3's hash", func() {
					Expect(len(result.Hashes)).To(Equal(1))
					h := result.Hashes[0]
					Expect(h).To(Equal(block3.GetHash()))
				})
			})
		})

		When("seek hash does not exist on remote peer", func() {

			// Target shape:
			// Remote Peer
			// [1]-[2]-[3]
			//
			// Local Peer
			// [1]
			Context("when remote blockchain shape is [1]-[2]-[3] and local blockchain shape: [1]", func() {

				var result *core.BlockHashes

				BeforeEach(func() {
					block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
					_, err := rp.GetBlockchain().ProcessBlock(block2)
					Expect(err).To(BeNil())

					block3 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 2)
					_, err = rp.GetBlockchain().ProcessBlock(block3)
					Expect(err).To(BeNil())

					result, err = lp.Gossip().SendGetBlockHashes(rp, nil, util.Hash{1, 2, 3})
					Expect(err).To(BeNil())
				})

				It("should get 2 block hash from remote peer", func() {
					Expect(len(result.Hashes)).To(Equal(2))
				})

				Specify("first header number = 2 and second header number = 3", func() {
					Expect(len(result.Hashes)).To(Equal(2))
					h2 := result.Hashes[0]
					h3 := result.Hashes[1]
					Expect(h2).To(Equal(block2.GetHash()))
					Expect(h3).To(Equal(block3.GetHash()))
				})
			})
		})

		When("seek hash does does not exist on the main chain of the remote peer", func() {

			// Target shape:
			// Remote Peer
			// [1]-[2]-[3]
			//  |__[2]
			// Local Peer
			// [1]
			Context("when remote blockchain shape is [1]-[2]-[3] and local blockchain shape: [1]", func() {

				var result *core.BlockHashes

				BeforeEach(func() {
					block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
					chain2block2 := MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
					_, err := rp.GetBlockchain().ProcessBlock(block2)
					Expect(err).To(BeNil())

					_, err = rp.GetBlockchain().ProcessBlock(chain2block2)
					Expect(err).To(BeNil())
					Expect(rp.GetBlockchain().GetChainsReader()).To(HaveLen(2))

					block3 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 2)
					_, err = rp.GetBlockchain().ProcessBlock(block3)
					Expect(err).To(BeNil())

					result, err = lp.Gossip().SendGetBlockHashes(rp, nil, chain2block2.GetHash())
					Expect(err).To(BeNil())
				})

				It("should get 2 block hash from remote peer", func() {
					Expect(len(result.Hashes)).To(Equal(2))
				})

				Specify("first header number = 2 and second header number = 3", func() {
					Expect(len(result.Hashes)).To(Equal(2))
					h2 := result.Hashes[0]
					h3 := result.Hashes[1]
					Expect(h2).To(Equal(block2.GetHash()))
					Expect(h3).To(Equal(block3.GetHash()))
				})
			})
		})

		// Target shape:
		// Remote Peer
		// [1]-[2]
		//
		// Local Peer
		// [1]
		Context("when remote blockchain shape is [1]-[2] and local peer blockchain shape is [1]", func() {

			var block2 types.Block
			var result *core.BlockHashes

			BeforeEach(func() {
				block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
				_, err := rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				result, err = lp.Gossip().SendGetBlockHashes(rp, nil, util.Hash{})
				Expect(err).To(BeNil())
			})

			It("should get 1 block hash from remote peer", func() {
				Expect(result.Hashes).To(HaveLen(1))
			})

			Specify("the block hash must match hash of block [2]", func() {
				Expect(result.Hashes).To(HaveLen(1))
				hash := result.Hashes[0]
				Expect(hash).To(Equal(block2.GetHash()))
			})
		})

		// Target shape:
		// Remote Peer
		// [1]
		//
		// Local Peer
		// [1]
		Context("when remote peer's blockchain shape is [1] and local peer's blockchain shape is [1]", func() {

			var result *core.BlockHashes
			var err error

			BeforeEach(func() {
				result, err = lp.Gossip().SendGetBlockHashes(rp, nil, util.Hash{})
				Expect(err).To(BeNil())
			})

			Specify("local peer block hash queue should be empty", func() {
				Expect(result.Hashes).To(HaveLen(0))
			})
		})

		// Target shape:
		// Remote Peer
		// [1]-[2]-[3] 	ChainA
		//  |__[2]		ChainB
		Context("when remote peer's blockchain shape is: [1]-[2]-[3] and local peer's blockchain shape is: [2]", func() {

			var block2, block3, chainBBlock2 types.Block

			BeforeEach(func() {
				block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
				chainBBlock2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)

				_, err := rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				// processing chainBBlock2 will create a fork
				_, err = rp.GetBlockchain().ProcessBlock(chainBBlock2)
				Expect(err).To(BeNil())
				Expect(rp.GetBlockchain().GetChainsReader()).To(HaveLen(2))

				block3 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 2)
				_, err = rp.GetBlockchain().ProcessBlock(block3)
				Expect(err).To(BeNil())
			})

			When("locator hash is block [2] of Chain B", func() {

				var err error
				var result *core.BlockHashes

				BeforeEach(func() {
					result, err = lp.Gossip().SendGetBlockHashes(rp, []util.Hash{chainBBlock2.GetHash()}, util.Hash{})
					Expect(err).To(BeNil())
				})

				Specify("the block hash queue should contain 2 hashes from ChainA", func() {
					Expect(result.Hashes).To(HaveCap(2))

					Context("first header number = [2] and second header number = [3]", func() {
						h2 := result.Hashes[0]
						h3 := result.Hashes[1]
						Expect(h2).To(Equal(block2.GetHash()))
						Expect(h3).To(Equal(block3.GetHash()))
					})
				})
			})
		})

		Context("when no known locator/block hash is shared with the remote peer", func() {

			var err error
			var result *core.BlockHashes

			BeforeEach(func() {
				result, err = lp.Gossip().SendGetBlockHashes(rp, []util.Hash{util.StrToHash("unknown")}, util.Hash{})
				Expect(err).To(BeNil())
			})

			It("local peer's block hash queue should be empty", func() {
				Expect(result.Hashes).To(HaveLen(0))
			})
		})
	})

	Describe(".SendGetBlockBodies", func() {

		var block2, block3 types.Block

		// Target shape:
		// Remote Peer
		// [1]-[2]-[3]
		//
		// Local Peer
		// [1]
		Context("when remote peer's blockchain shape is [1]-[2]-[3] and local peer's blockchain shape is: [1]", func() {
			Context("at least one hash is requested", func() {
				BeforeEach(func() {
					block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, sender, 1)
					_, err := rp.GetBlockchain().ProcessBlock(block2)
					Expect(err).To(BeNil())

					block3 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, sender, 2)
					_, err = rp.GetBlockchain().ProcessBlock(block3)
					Expect(err).To(BeNil())
				})

				It("should successfully fetch block [2] and [3]", func() {
					hashes := []util.Hash{block2.GetHash(), block3.GetHash()}
					blockBodies, err := lp.Gossip().SendGetBlockBodies(rp, hashes)
					Expect(err).To(BeNil())
					Expect(blockBodies.Blocks).To(HaveCap(2))
					Expect(blockBodies.Blocks[0].Header.Number).To(Equal(uint64(2)))
					Expect(blockBodies.Blocks[1].Header.Number).To(Equal(uint64(3)))
				})
			})
		})
	})

})
