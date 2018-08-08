package blockchain

import (
	"math/big"

	"github.com/ellcrys/elld/blockchain/common"
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
				When("and back linking enabled", func() {

					It("should successfully create seeded state tree", func() {
						tree, err := genesisChain.NewStateTree(false)
						Expect(err).To(BeNil())
						initialRoot := tree.Root()

						Describe("tree must be seeded with the state root of the tip block", func() {
							curItems := tree.GetItems()
							Expect(curItems).To(HaveLen(1))
							Expect(curItems[0]).To(Equal(common.TreeItem(genesisBlock.Header.StateRoot.Bytes())))
						})

						Describe("must derive new state root after adding items to tree", func() {
							tree.Add(item1)
							tree.Add(item2)
							err = tree.Build()
							Expect(err).To(BeNil())

							Expect(tree.Root()).NotTo(Equal(initialRoot))
							expected := util.Hash{160, 11, 29, 115, 220, 4, 126, 243, 116, 182, 16, 77, 80, 58, 165, 104, 68, 64, 109, 64, 214, 208, 168, 170, 86, 185, 40, 236, 101, 140, 46, 150}
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
				genesisBlock = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, sender.Addr(), sender, "1", "0.1", 1532730722),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
				})
			})

			It("should return err when the block number does not serially match the current tip number", func() {
				genesisBlock.Header.Number = 3
				err = genesisChain.append(genesisBlock)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to append: candidate block number {3} is not the expected block number {expected=2}"))
			})

			It("should return err when the block's parent hash does not match the hash of the current tip block", func() {
				err = genesisChain.append(genesisBlock)
				Expect(err).To(BeNil())

				genesisBlock.Header.Number = 3
				genesisBlock.Header.ParentHash = util.StrToHash("incorrect")
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
				exist, err := genesisChain.hasBlock("some_unknown_hash")
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
			})

			It("should return true if block exist in the chain", func() {
				exist, err := genesisChain.hasBlock(genesisBlock.GetHash().HexStr())
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
				header, err := genesisChain.getBlockHeaderByHash("unknown")
				Expect(err).To(Equal(common.ErrBlockNotFound))
				Expect(header).To(BeNil())
			})

			It("should successfully get block header by hash", func() {
				header, err := genesisChain.getBlockHeaderByHash(genesisBlock.GetHash().HexStr())
				Expect(err).To(BeNil())
				Expect(header).ToNot(BeNil())
			})
		})

		Describe(".getBlockByHash", func() {
			It("should return error if block is not found", func() {
				block, err := genesisChain.getBlockByHash("unknown")
				Expect(err).To(Equal(common.ErrBlockNotFound))
				Expect(block).To(BeNil())
			})

			It("should successfully get block by hash", func() {
				block, err := genesisChain.getBlockByHash(genesisBlock.GetHash().HexStr())
				Expect(err).To(BeNil())
				Expect(block).ToNot(BeNil())
			})
		})
	})
}
