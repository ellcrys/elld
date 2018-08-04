package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
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
						emptyChain := NewChain("my_chain", store, cfg, log)
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
						emptyChain := NewChain("my_chain", store, cfg, log)
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
				When("and back linking enabled", func() {
					It("should successfully create seeded state tree", func() {
						tree, err := chain.NewStateTree(false)
						Expect(err).To(BeNil())
						initialRoot := tree.Root()

						Describe("tree must be seeded with the state root of the tip block", func() {
							curItems := tree.GetItems()
							Expect(curItems).To(HaveLen(1))
							Expect(curItems[0]).To(Equal(common.TreeItem(testdata.GenesisBlock.Header.StateRoot.Bytes())))
						})

						Describe("must derive new state root after adding items to tree", func() {
							tree.Add(item1)
							tree.Add(item2)
							err = tree.Build()
							Expect(err).To(BeNil())

							Expect(tree.Root()).NotTo(Equal(initialRoot))
							expected := util.Hash{62, 249, 156, 202, 216, 252, 54, 66, 242, 156, 178, 183, 239, 44, 150, 250, 124, 8, 58, 154, 119, 151, 123, 153, 31, 143, 132, 164, 15, 184, 85, 139}
							Expect(tree.Root()).To(Equal(expected))
						})

					})
				})

				When("and back linking disabled", func() {
					It("should successfully create state tree that has not been seeded", func() {
						tree, err := chain.NewStateTree(true)
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

			var block2, block2_2, block3 *wire.Block

			BeforeEach(func() {
				block2 = testdata.BlockSet1[0]
				block3 = testdata.BlockSet1[1]
				block2_2 = testdata.BlockSet1[3]
			})

			It("should return err when the block number does not serially match the current tip number", func() {
				err = chain.append(block3)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to append: candidate block number {3} is not the expected block number {expected=2}"))
			})

			It("should return err when the block's parent hash does not match the hash of the current tip block", func() {
				err = chain.append(block2)
				Expect(err).To(BeNil())

				err = chain.append(block2_2)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to append block: parent hash does not match the hash of the current block"))
			})

			It("should return no error", func() {
				err = chain.append(block2)
				Expect(err).To(BeNil())
			})
		})

		Describe(".hashBlock", func() {

			It("should return false if block does not exist in the chain", func() {
				exist, err := chain.hasBlock("some_unknown_hash")
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
			})

			It("should return true if block exist in the chain", func() {
				exist, err := chain.hasBlock(block.GetHash().HexStr())
				Expect(err).To(BeNil())
				Expect(exist).To(BeTrue())
			})
		})

		Describe(".height", func() {

			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", store, cfg, log)
			})

			It("should return zero if chain has no block", func() {
				height, err := chain.height()
				Expect(err).To(BeNil())
				Expect(height).To(Equal(uint64(0)))
			})

			It("should return 1 if chain contains 1 block", func() {
				err := chain.append(block)
				Expect(err).To(BeNil())

				height, err := chain.height()
				Expect(err).To(BeNil())
				Expect(height).To(Equal(uint64(1)))
			})
		})

		Describe(".getBlockHeaderByHash", func() {

			It("should return err if block was not found", func() {
				header, err := chain.getBlockHeaderByHash("unknown")
				Expect(err).To(Equal(common.ErrBlockNotFound))
				Expect(header).To(BeNil())
			})

			It("should successfully get block header by hash", func() {
				header, err := chain.getBlockHeaderByHash(block.GetHash().HexStr())
				Expect(err).To(BeNil())
				Expect(header).ToNot(BeNil())
			})
		})

		Describe(".getBlockByHash", func() {
			It("should return error if block is not found", func() {
				block, err := chain.getBlockByHash("unknown")
				Expect(err).To(Equal(common.ErrBlockNotFound))
				Expect(block).To(BeNil())
			})

			It("should successfully get block by hash", func() {
				block, err := chain.getBlockByHash(block.GetHash().HexStr())
				Expect(err).To(BeNil())
				Expect(block).ToNot(BeNil())
			})
		})
	})
}
