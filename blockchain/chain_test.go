package blockchain

import (
	"math/big"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var ChainTest = func() bool {
	return Describe("Chain", func() {
		Describe(".NewStateTree", func() {

			var emptyTreeRoot, chainRoot util.Hash
			var item1, item2 common.TreeItem

			BeforeEach(func() {
				item1 = common.TreeItem([]byte("age"))
				item2 = common.TreeItem([]byte("gender"))

				emptyTreeRoot = util.Hash{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				}

				chainRoot = util.Hash{
					0xad, 0xfa, 0xd6, 0x2b, 0x04, 0x74, 0x82, 0x37, 0x4b, 0x59, 0x48, 0x77, 0x82, 0x50, 0x96, 0xf2,
					0x1d, 0xdf, 0xfa, 0x06, 0x78, 0x22, 0x35, 0xb5, 0xb1, 0x87, 0x43, 0x5d, 0xd5, 0x2b, 0xf1, 0x24,
				}
			})

			Context("with empty chain", func() {
				When("and back linking enabled", func() {
					It("should successfully create a tree, added 2 items and return root", func() {
						emptyChain := NewChain("my_chain", db, cfg, log)
						tree, err := emptyChain.NewStateTree(false)
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(emptyTreeRoot))

						tree.Add(item1)
						tree.Add(item2)
						err = tree.Build()
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(chainRoot))
					})
				})

				When("and back linking disabled", func() {
					It("should successfully create a tree and same root as an empty chain", func() {
						emptyChain := NewChain("my_chain", db, cfg, log)
						tree, err := emptyChain.NewStateTree(true)
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(emptyTreeRoot))

						tree.Add(item1)
						tree.Add(item2)
						err = tree.Build()
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(chainRoot))
					})
				})
			})

			Context("with a non-empty chain", func() {
				When("back linking enabled", func() {

					It("should successfully create seeded state tree", func() {
						tree, err := genesisChain.NewStateTree(false)
						Expect(err).To(BeNil())
						initialRoot := tree.Root()

						Describe("tree must be seeded with the state root of the tip block", func() {
							curItems := tree.GetItems()
							Expect(curItems).To(HaveLen(1))
							Expect(curItems[0]).To(Equal(common.TreeItem(genesisBlock.GetHeader().GetStateRoot().Bytes())))
						})

						Describe("must derive new state root after adding items to tree", func() {
							tree.Add(item1)
							tree.Add(item2)
							err = tree.Build()
							Expect(err).To(BeNil())

							Expect(tree.Root()).NotTo(Equal(initialRoot))
							expected := util.Hash{52, 30, 117, 201, 22, 55, 159, 68, 20, 187, 94, 245, 61, 222, 41, 233, 6, 200, 210, 149, 34, 141, 47, 14, 1, 218, 56, 1, 138, 221, 141, 57}
							Expect(tree.Root()).To(Equal(expected))
						})

					})
				})

				When("and back linking disabled", func() {
					It("should successfully create state tree that has not been seeded", func() {
						tree, err := genesisChain.NewStateTree(true)
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(emptyTreeRoot))

						tree.Add(item1)
						tree.Add(item2)
						err = tree.Build()
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(chainRoot))
					})
				})
			})
		})

		Describe(".append", func() {
			BeforeEach(func() {
				genesisBlock = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730722),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			It("should return err when the block number does not serially match the current tip number", func() {
				genesisBlock.GetHeader().SetNumber(3)
				err = genesisChain.append(genesisBlock)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to append: candidate block number {3} is not the expected block number {expected=2}"))
			})

			It("should return err when the block's parent hash does not match the hash of the current tip block", func() {
				err = genesisChain.append(genesisBlock)
				Expect(err).To(BeNil())

				genesisBlock.GetHeader().SetNumber(3)
				genesisBlock.GetHeader().SetParentHash(util.StrToHash("incorrect"))
				err = genesisChain.append(genesisBlock)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to append block: parent hash does not match the hash of the current block"))
			})

			It("should return no error", func() {
				err = genesisChain.append(genesisBlock)
				Expect(err).To(BeNil())
			})
		})

		Describe(".hashBlock", func() {

			It("should return false if block does not exist in the chain", func() {
				exist, err := genesisChain.hasBlock(util.Hash{1, 2, 3})
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
			})

			It("should return true if block exist in the chain", func() {
				exist, err := genesisChain.hasBlock(genesisBlock.GetHash())
				Expect(err).To(BeNil())
				Expect(exist).To(BeTrue())
			})
		})

		Describe(".height", func() {

			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", db, cfg, log)
			})

			It("should return zero if chain has no block", func() {
				height, err := chain.height()
				Expect(err).To(BeNil())
				Expect(height).To(Equal(uint64(0)))
			})

			It("should return 1 if chain contains 1 block", func() {
				err := chain.append(genesisBlock)
				Expect(err).To(BeNil())

				height, err := chain.height()
				Expect(err).To(BeNil())
				Expect(height).To(Equal(uint64(1)))
			})
		})

		Describe(".getBlockHeaderByHash", func() {

			It("should return err if block was not found", func() {
				header, err := genesisChain.getBlockHeaderByHash(util.Hash{1, 2, 3})
				Expect(err).To(Equal(core.ErrBlockNotFound))
				Expect(header).To(BeNil())
			})

			It("should successfully get block header by hash", func() {
				header, err := genesisChain.getBlockHeaderByHash(genesisBlock.GetHash())
				Expect(err).To(BeNil())
				Expect(header).ToNot(BeNil())
			})
		})

		Describe(".getBlockByHash", func() {
			It("should return error if block is not found", func() {
				block, err := genesisChain.getBlockByHash(util.Hash{1, 2, 3})
				Expect(err).To(Equal(core.ErrBlockNotFound))
				Expect(block).To(BeNil())
			})

			It("should successfully get block by hash", func() {
				block, err := genesisChain.getBlockByHash(genesisBlock.GetHash())
				Expect(err).To(BeNil())
				Expect(block).ToNot(BeNil())
			})
		})

		Describe(".removeBlock", func() {

			var block2 core.Block

			BeforeEach(func() {
				block2 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730722),
						objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730722),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				_, err := bc.ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			It("should return ErrBlockNotFound if block does not exist", func() {
				err := genesisChain.removeBlock(100)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBlockNotFound))
			})

			It("should successfully delete a block", func() {
				blockNum := block2.GetNumber()
				err := genesisChain.removeBlock(blockNum)
				Expect(err).To(BeNil())

				Describe("the block must be deleted", func() {
					blockKey := common.MakeBlockKey(genesisChain.id.Bytes(), blockNum)
					result := db.GetByPrefix(blockKey)
					Expect(result).To(HaveLen(0))
				})

				Describe("all account must be deleted", func() {
					acctKeys := common.MakeAccountsKey(genesisChain.id.Bytes())
					result := db.GetByPrefix(acctKeys)
					for _, r := range result {
						bn := common.DecodeBlockNumber(r.Key)
						Expect(bn).ToNot(Equal(blockNum))
					}
				})

				Describe("all transactions must be deleted", func() {
					txsKeys := common.MakeTxsQueryKey(genesisChain.id.Bytes())
					result := db.GetByPrefix(txsKeys)
					for _, r := range result {
						bn := common.DecodeBlockNumber(r.Key)
						Expect(bn).ToNot(Equal(blockNum))
					}
				})
			})
		})

		Describe(".save", func() {

			It("should be successful and return nil", func() {
				chain1 := NewChain("c1", db, cfg, log)
				err := chain1.save()
				Expect(err).To(BeNil())

				Describe("should exist in database", func() {
					result := chain1.store.DB().GetByPrefix(common.MakeChainKey(chain1.id.Bytes()))
					Expect(result).To(HaveLen(1))
				})
			})
		})

		Describe(".findParent", func() {

			var parentChain1, parentChain2 *Chain

			BeforeEach(func() {
				parentChain1 = NewChain("p1", db, cfg, log)
				err := parentChain1.save()
				Expect(err).To(BeNil())

				parentChain2 = NewChain("p2", db, cfg, log)
				err = parentChain2.save()
				Expect(err).To(BeNil())
			})

			It("should return nil chain and nil error", func() {
				chain1 := NewChain("c1", db, cfg, log)
				ch, err := chain1.loadParent()
				Expect(ch).To(BeNil())
				Expect(err).To(BeNil())
			})

			It("should return ErrChainParentNotFound when chain parent was not found", func() {
				chain1 := NewChain("c1", db, cfg, log)
				chain1.info.ParentChainID = "xyz"
				ch, err := chain1.loadParent()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrChainParentNotFound))
				Expect(ch).To(BeNil())
			})

			It("should return ErrChainParentNotFound when chain parent block was not found", func() {
				chain1 := NewChain("c1", db, cfg, log)
				chain1.info.ParentChainID = parentChain1.GetID()
				chain1.info.ParentBlockNumber = 100

				_, err = chain1.loadParent()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrChainParentBlockNotFound))
			})

			It("should return parent chain and parent block", func() {
				parentChain1Block := makeBlock(parentChain1)
				err = parentChain1.store.PutBlock(parentChain1Block)
				Expect(err).To(BeNil())

				chain1 := NewChain("c1", db, cfg, log)
				chain1.info.ParentChainID = parentChain1.GetID()
				chain1.info.ParentBlockNumber = parentChain1Block.GetNumber()

				result, err := chain1.loadParent()
				Expect(err).To(BeNil())
				Expect(result.GetID()).To(Equal(parentChain1.GetID()))
				Expect(result.info).To(Equal(parentChain1.info))
			})

			It("should get parents of chains", func() {

				// parent chain 2, child of parent chain 1
				parentChain1Block := makeBlock(parentChain1)
				err = parentChain1.store.PutBlock(parentChain1Block)
				Expect(err).To(BeNil())
				parentChain2.info.ParentChainID = parentChain1.GetID()
				parentChain2.info.ParentBlockNumber = parentChain1Block.GetNumber()
				err = parentChain2.save()
				Expect(err).To(BeNil())

				// parent chain 3, child of parent 2
				p2Block := makeBlock(parentChain2)
				err = parentChain2.store.PutBlock(p2Block)
				Expect(err).To(BeNil())
				p3 := NewChain("p3", db, cfg, log)
				p3.info.ParentChainID = parentChain2.GetID()
				p3.info.ParentBlockNumber = p2Block.GetNumber()
				err = p3.save()
				Expect(err).To(BeNil())

				p3Parent, err := p3.loadParent()
				Expect(err).To(BeNil())
				Expect(p3Parent.info).To(Equal(parentChain2.info))

				p2Parent, err := parentChain2.loadParent()
				Expect(err).To(BeNil())
				Expect(p2Parent.info).To(Equal(parentChain1.info))
			})
		})

		Describe(".GetRoot", func() {

			var chainB, chainC *Chain
			var block2Main core.Block

			BeforeEach(func() {
				tip, _ := genesisChain.Current()
				Expect(tip.GetNumber()).To(Equal(uint64(1)))
			})

			// Target shape
			// [1]-[2]-[3]-[4]-[5]  Main
			//      |__[3]-[4]		Chain B
			//          |__[4]		Chain C
			BeforeEach(func() {

				// main chain blocks
				block2Main = makeBlock(genesisChain)
				_, err = bc.ProcessBlock(block2Main)
				Expect(err).To(BeNil())

				block3Main := makeBlock(genesisChain)
				block3ChainB := makeBlock(genesisChain)

				_, err = bc.ProcessBlock(block3Main)
				Expect(err).To(BeNil())

				_, err = bc.ProcessBlock(makeBlock(genesisChain))
				Expect(err).To(BeNil())

				_, err = bc.ProcessBlock(makeBlock(genesisChain))
				Expect(err).To(BeNil())

				// start a fork (Chain B)
				chainBReader, err := bc.ProcessBlock(block3ChainB)
				Expect(err).To(BeNil())
				Expect(len(bc.chains)).To(Equal(2))
				chainB = bc.chains[chainBReader.GetID()]

				block4ChainC := makeBlock(chainB)
				block4ChainB := makeBlock(chainB)

				_, err = bc.ProcessBlock(block4ChainB)
				Expect(err).To(BeNil())

				// start a fork (Chain C)
				chainCReader, err := bc.ProcessBlock(block4ChainC)
				Expect(err).To(BeNil())
				Expect(len(bc.chains)).To(Equal(3))
				chainC = bc.chains[chainCReader.GetID()]

			})

			BeforeEach(func() {
				tip, _ := genesisChain.Current()
				Expect(tip.GetNumber()).To(Equal(uint64(5)))
				parent := genesisChain.GetParent()
				Expect(parent).To(BeNil())

				chainBTip, _ := chainB.Current()
				Expect(chainBTip.GetNumber()).To(Equal(uint64(4)))
				parent = chainB.GetParent()
				Expect(parent).ToNot(BeNil())
				Expect(parent.GetID()).To(Equal(genesisChain.GetID()))

				chainCTip, _ := chainC.Current()
				Expect(chainCTip.GetNumber()).To(Equal(uint64(4)))
				parent = chainC.GetParent()
				Expect(parent).ToNot(BeNil())
				Expect(parent.GetID()).To(Equal(chainB.GetID()))
			})

			It("should return nil if chains has no parent", func() {
				root := genesisChain.GetRoot()
				Expect(root).To(BeNil())
			})

			It("successfully get the root of chain C as block 2 of genesis", func() {
				root := chainC.GetRoot()
				Expect(root).ToNot(BeNil())
				Expect(root.GetHeader()).To(Equal(block2Main.GetHeader()))
			})

			It("successfully get the root of chain B as block 2 of genesis", func() {
				root := chainB.GetRoot()
				Expect(root).ToNot(BeNil())
				Expect(root.GetHeader()).To(Equal(block2Main.GetHeader()))
			})
		})
	})
}
