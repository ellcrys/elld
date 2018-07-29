package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var ChainTest = func() bool {
	return Describe("Chain", func() {
		Describe(".NewStateTree", func() {

			var emptyChainRoot = []uint8{4, 1, 123, 153, 29, 46, 145, 109, 24, 125, 58, 216, 184, 21, 82, 235, 52, 105, 246, 181, 195, 203, 61, 165, 193, 22, 243, 98, 55, 44, 162, 75}
			var item1 = common.TreeItem([]byte("age"))
			var item2 = common.TreeItem([]byte("sex"))

			Context("with empty chain", func() {
				When("and back linking enabled", func() {
					It("should successfully create a tree and return root", func() {
						newChain := NewChain("my_chain", store, cfg, log)
						tree, err := newChain.NewStateTree(false)
						Expect(err).To(BeNil())
						tree.Add(item1)
						tree.Add(item2)

						err = tree.Build()
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(emptyChainRoot))
					})
				})

				When("and back linking disabled", func() {
					It("should successfully create a tree and same root as an empty chain", func() {
						newChain := NewChain("my_chain", store, cfg, log)
						tree, err := newChain.NewStateTree(true)
						Expect(err).To(BeNil())
						tree.Add(item1)
						tree.Add(item2)

						err = tree.Build()
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(emptyChainRoot))
					})
				})
			})

			Context("with a non-empty chain", func() {
				When("and back linking enabled", func() {
					It("should successfully create without seed a new state tree and return its root", func() {
						tree, err := chain.NewStateTree(false)
						Expect(err).To(BeNil())
						tree.Add(item1)
						tree.Add(item2)

						err = tree.Build()
						Expect(err).To(BeNil())
						Expect(tree.Root()).NotTo(Equal(emptyChainRoot))
						expected := []uint8{191, 47, 2, 172, 153, 59, 66, 122, 196, 204, 250, 4, 210, 29, 53, 102, 49, 94, 168, 246, 156, 182, 191, 115, 39, 232, 105, 68, 116, 238, 91, 160}
						Expect(tree.Root()).To(Equal(expected))
					})
				})

				When("and back linking disabled", func() {
					It("should successfully create without seed a new state tree and return its root", func() {
						tree, err := chain.NewStateTree(true)
						Expect(err).To(BeNil())
						tree.Add(item1)
						tree.Add(item2)

						err = tree.Build()
						Expect(err).To(BeNil())
						Expect(tree.Root()).To(Equal(emptyChainRoot))
					})
				})
			})
		})

		Describe(".append", func() {

			var block, block2, block3 *wire.Block

			BeforeEach(func() {
				block, _ = wire.BlockFromString(testdata.ChainDotJSON[1])
				block2, _ = wire.BlockFromString(testdata.ChainDotJSON[2])
				block3, _ = wire.BlockFromString(testdata.ChainDotJSON[3])
			})

			It("should return err when the block number does not serial match the current tip number", func() {
				err = chain.append(block3)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to append: candidate block number {1} is not the expected block number {expected=2}"))
			})

			It("should return err when the block's parent hash does not match the hash of the current tip block", func() {
				err = chain.append(block)
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
				exist, err := chain.hasBlock(block.GetHash())
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
				block, _ := wire.BlockFromString(testdata.ChainDotJSON[3])
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
				header, err := chain.getBlockHeaderByHash(block.GetHash())
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
				block, err := chain.getBlockByHash(block.GetHash())
				Expect(err).To(BeNil())
				Expect(block).ToNot(BeNil())
			})
		})

	})
}
