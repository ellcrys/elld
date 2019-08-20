package blockchain

import (
	"os"

	"github.com/ellcrys/elld/util/logger"

	. "github.com/onsi/ginkgo"

	"github.com/ellcrys/elld/blockchain/common"
	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chain", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock types.Block
	var genesisChain *Chain
	var nodeKey, sender, receiver *crypto.Key
	var log = logger.NewLogrusNoOp()

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(log)
		err = db.Open(cfg.NetDataDir())
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)

		nodeKey = sender
		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
		bc.SetNodeKey(nodeKey)
	})

	BeforeEach(func() {
		genesisBlock, err = LoadBlockFromFile("genesis-test.json")
		Expect(err).To(BeNil())
		bc.SetGenesisBlock(genesisBlock)
		err = bc.Up()
		Expect(err).To(BeNil())
		genesisChain = bc.bestChain
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe(".GetRoot", func() {

		var chainB, chainC *Chain
		var block2Main types.Block

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
			block2Main = MakeBlockWithTx(bc, genesisChain, sender, 1)
			_, err = bc.ProcessBlock(block2Main)
			Expect(err).To(BeNil())

			block3Main := MakeBlockWithTx(bc, genesisChain, sender, 2)
			block3ChainB := MakeBlockWithTx(bc, genesisChain, sender, 2)

			_, err = bc.ProcessBlock(block3Main)
			Expect(err).To(BeNil())

			_, err = bc.ProcessBlock(MakeBlockWithTx(bc, genesisChain, sender, 3))
			Expect(err).To(BeNil())

			_, err = bc.ProcessBlock(MakeBlockWithTx(bc, genesisChain, sender, 4))
			Expect(err).To(BeNil())

			// start a fork (Chain B)
			chainBReader, err := bc.ProcessBlock(block3ChainB)
			Expect(err).To(BeNil())
			Expect(len(bc.chains)).To(Equal(2))
			chainB = bc.chains[chainBReader.GetID()]

			block4ChainB := MakeBlockWithTx(bc, chainB, sender, 2)
			block4ChainC := MakeBlockWithTx(bc, chainB, sender, 2)

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

			var tree types.Tree
			var err error

			BeforeEach(func() {
				emptyChain := NewChain("my_chain", db, cfg, log)
				tree, err = emptyChain.NewStateTree()
				Expect(err).To(BeNil())
				Expect(tree.Root()).To(Equal(emptyTreeRoot))
			})

			It("Should add 2 items and return expected root", func() {
				tree.Add(item1)
				tree.Add(item2)
				err = tree.Build()
				Expect(err).To(BeNil())
				Expect(tree.Root()).To(Equal(chainRoot))
				Expect(tree.GetItems()).To(HaveLen(2))
			})
		})

		Context("with a non-empty chain", func() {

			var tree types.Tree
			var initialRoot util.Hash
			var err error

			BeforeEach(func() {
				tree, err = genesisChain.NewStateTree()
				Expect(err).To(BeNil())
				Expect(tree.Root()).To(Equal(emptyTreeRoot))
				initialRoot = tree.Root()
			})

			Specify("tree must be seeded with the state root of the tip block", func() {
				curItems := tree.GetItems()
				Expect(curItems).To(HaveLen(1))
				Expect(curItems[0]).To(Equal(common.TreeItem(genesisBlock.GetHeader().GetStateRoot().Bytes())))
			})

			Specify("must derive new state root after adding items to tree", func() {
				tree.Add(item1)
				tree.Add(item2)
				err = tree.Build()
				Expect(err).To(BeNil())
				newRoot := tree.Root()
				Expect(newRoot).NotTo(Equal(initialRoot))
			})
		})
	})

	Describe(".append", func() {

		BeforeEach(func() {
			genesisBlock = MakeBlock(bc, genesisChain, sender, receiver)
		})

		It("should return err when the block number does not serially match the current tip number", func() {
			genesisBlock.GetHeader().SetNumber(3)
			err = genesisChain.Append(genesisBlock)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unable to append: candidate block number {3} is not the expected block number {expected=2}"))
		})

		It("should return err when the block's parent hash does not match the hash of the current tip block", func() {
			err = genesisChain.Append(genesisBlock)
			Expect(err).To(BeNil())

			genesisBlock.GetHeader().SetNumber(3)
			genesisBlock.GetHeader().SetParentHash(util.StrToHash("incorrect"))
			err = genesisChain.Append(genesisBlock)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unable to append block: parent hash does not match the hash of the current block"))
		})

		It("should return no error", func() {
			err = genesisChain.Append(genesisBlock)
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
			err := chain.Append(genesisBlock)
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

		var block2 types.Block

		It("should return ErrBlockNotFound if block does not exist", func() {
			_, err := genesisChain.removeBlock(100)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(core.ErrBlockNotFound))
		})

		Context("mined block must be deleted", func() {
			When("block creator and nodeKey are the same", func() {
				BeforeEach(func() {
					block2 = MakeBlock(bc, genesisChain, nodeKey, receiver)
					_, err := bc.ProcessBlock(block2)
					Expect(err).To(BeNil())

					minedBlockKey := common.MakeKeyMinedBlock(genesisChain.id.Bytes(), block2.GetNumber())
					result := db.GetByPrefix(minedBlockKey)
					Expect(result).To(HaveLen(1))
				})

				BeforeEach(func() {
					deleted, err := genesisChain.removeBlock(block2.GetNumber())
					Expect(err).To(BeNil())
					Expect(deleted.GetHash()).To(Equal(block2.GetHash()))
				})

				Specify("that no the mined block object must no longer exist", func() {
					blockKey := common.MakeKeyMinedBlock(genesisChain.id.Bytes(), block2.GetNumber())
					result := db.GetByPrefix(blockKey)
					Expect(result).To(HaveLen(0))
				})
			})
		})

		Context("when block is successfully deleted", func() {

			BeforeEach(func() {
				block2 = MakeBlock(bc, genesisChain, sender, receiver)
				_, err := bc.ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			BeforeEach(func() {
				deleted, err := genesisChain.removeBlock(block2.GetNumber())
				Expect(err).To(BeNil())
				Expect(deleted.GetHash()).To(Equal(block2.GetHash()))
			})

			Specify("no block with the deleted block number must exist", func() {
				blockKey := common.MakeKeyBlock(genesisChain.id.Bytes(), block2.GetNumber())
				result := db.GetByPrefix(blockKey)
				Expect(result).To(HaveLen(0))
			})

			Specify("accounts associated with the block must be deleted", func() {
				acctKeys := common.MakeQueryKeyAccounts(genesisChain.id.Bytes())
				result := db.GetByPrefix(acctKeys)
				for _, r := range result {
					bn := util.DecodeNumber(r.Key)
					Expect(bn).ToNot(Equal(block2.GetNumber()))
				}
			})

			Specify("transactions associated with the block must be deleted", func() {
				txsKeys := common.MakeQueryKeyTransactions(genesisChain.id.Bytes())
				result := db.GetByPrefix(txsKeys)
				for _, r := range result {
					bn := util.DecodeNumber(r.Key)
					Expect(bn).ToNot(Equal(block2.GetNumber()))
				}
			})

			Specify("block hash pointer associated with the block must be deleted", func() {
				blockHashPointer := common.MakeKeyBlockHash(genesisChain.id.Bytes(), block2.GetHash().Hex())
				result := db.GetByPrefix(blockHashPointer)
				Expect(result).To(BeEmpty())
			})
		})
	})

	Describe(".GetMinedBlocks", func() {
		When("there is only 1 mined blocks", func() {
			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("c1", db, cfg, log)
				Expect(bc.CreateAccount(1, chain, &core.Account{
					Type:    core.AccountTypeBalance,
					Address: util.String(sender.Addr()),
					Balance: "1000",
				})).To(BeNil())
			})

			BeforeEach(func() {
				block1 := MakeBlock(bc, chain, sender, receiver)
				Expect(chain.store.PutMinedBlock(block1)).To(BeNil())
			})

			It("should fetch 1 blocks", func() {
				result, _, err := chain.GetMinedBlocks(&core.ArgGetMinedBlock{})
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
			})
		})

		When("there are only 3 mined blocks", func() {
			var chain *Chain
			var block1, block2, block3 types.Block

			BeforeEach(func() {
				chain = NewChain("c1", db, cfg, log)
				Expect(bc.CreateAccount(1, chain, &core.Account{
					Type:    core.AccountTypeBalance,
					Address: util.String(sender.Addr()),
					Balance: "1000",
				})).To(BeNil())

				Expect(bc.CreateAccount(1, chain, &core.Account{
					Type:    core.AccountTypeBalance,
					Address: util.String(receiver.Addr()),
					Balance: "100",
				})).To(BeNil())
			})

			BeforeEach(func() {
				block1 = MakeBlock(bc, chain, sender, receiver)
				Expect(chain.Append(block1)).To(BeNil())
				Expect(chain.store.PutMinedBlock(block1)).To(BeNil())

				block2 = MakeBlock(bc, chain, receiver, receiver) // different creator pub key
				Expect(chain.Append(block2)).To(BeNil())
				Expect(chain.store.PutMinedBlock(block2)).To(BeNil())

				block3 = MakeBlock(bc, chain, sender, receiver)
				Expect(chain.Append(block3)).To(BeNil())
				Expect(chain.store.PutMinedBlock(block3)).To(BeNil())
			})

			Context("argument: `limit`", func() {
				When("limit is set to 1", func() {
					var result []*core.MinedBlock
					var hasMore bool

					BeforeEach(func() {
						var err error
						result, hasMore, err = chain.GetMinedBlocks(&core.ArgGetMinedBlock{Limit: 1})
						Expect(err).To(BeNil())
					})

					It("should return 1 block that has the highest block number", func() {
						Expect(result).To(HaveLen(1))
						Expect(result[0].Number).To(Equal(block3.GetNumber()))
					})

					Specify("that hasMore is true", func() {
						Expect(hasMore).To(BeTrue())
					})
				})

				When("limit is set to 3", func() {
					var result []*core.MinedBlock
					var hasMore bool

					BeforeEach(func() {
						var err error
						result, hasMore, err = chain.GetMinedBlocks(&core.ArgGetMinedBlock{Limit: 3})
						Expect(err).To(BeNil())
					})

					It("should return 3 blocks with the last one having the lowest block number", func() {
						Expect(result).To(HaveLen(3))
						Expect(result[0].Number).To(Equal(block3.GetNumber()))
						Expect(result[2].Number).To(Equal(block1.GetNumber()))
					})

					Specify("that hasMore is false", func() {
						Expect(hasMore).To(BeFalse())
					})
				})
			})

			Context("argument: `lastHash`", func() {
				When("lastHash is set to block3.Hash", func() {
					var result []*core.MinedBlock
					var hasMore bool

					BeforeEach(func() {
						var err error
						result, hasMore, err = chain.GetMinedBlocks(&core.ArgGetMinedBlock{
							LastHash: block3.GetHash().HexStr()})
						Expect(err).To(BeNil())
					})

					It("should return 2 blocks (block 2 and 3)", func() {
						Expect(result).To(HaveLen(2))
						Expect(result[0].Number).To(Equal(block2.GetNumber()))
						Expect(result[1].Number).To(Equal(block1.GetNumber()))
					})

					Specify("that hasMore is true", func() {
						Expect(hasMore).To(BeFalse())
					})
				})
			})

			Context("argument: `creatorPubKey`", func() {
				When("creatorPubKey is set to block1.Header.CreatorPubKey", func() {
					var result []*core.MinedBlock
					var hasMore bool

					BeforeEach(func() {
						var err error
						result, hasMore, err = chain.GetMinedBlocks(&core.ArgGetMinedBlock{
							CreatorPubKey: block3.GetHeader().GetCreatorPubKey().String()})
						Expect(err).To(BeNil())
					})

					It("should return 2 blocks (block 1 and 3)", func() {
						Expect(result).To(HaveLen(2))
						Expect(result[0].Number).To(Equal(block3.GetNumber()))
						Expect(result[1].Number).To(Equal(block1.GetNumber()))
					})

					Specify("that hasMore is true", func() {
						Expect(hasMore).To(BeFalse())
					})
				})
			})
		})
	})

	Describe(".save", func() {

		Context("on successful save", func() {
			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("c1", db, cfg, log)
				err := chain.save()
				Expect(err).To(BeNil())
			})

			It("should exist in database", func() {
				result := chain.store.DB().GetByPrefix(common.MakeKeyChain(chain.id.Bytes()))
				Expect(result).To(HaveLen(1))
			})
		})
	})

	Describe(".findParent", func() {

		var chain, chain2 *Chain

		BeforeEach(func() {
			chain = NewChain("p1", db, cfg, log)
			err := chain.save()
			Expect(err).To(BeNil())

			chain2 = NewChain("p2", db, cfg, log)
			err = chain2.save()
			Expect(err).To(BeNil())

			Expect(bc.CreateAccount(1, chain, &core.Account{
				Type:    core.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "1000",
			})).To(BeNil())
		})

		Context("when chain has no parent", func() {
			It("should return nil", func() {
				ch, err := chain.loadParent()
				Expect(ch).To(BeNil())
				Expect(err).To(BeNil())
			})
		})

		Context("when chain's parent chain was not found", func() {
			It("should return ErrChainParentNotFound", func() {
				chain.info.ParentChainID = "xyz"
				ch, err := chain.loadParent()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrChainParentNotFound))
				Expect(ch).To(BeNil())
			})
		})

		Context("when the chains parent block was not found", func() {
			It("should return ErrChainParentNotFound", func() {
				chain.info.ParentChainID = chain.GetID()
				chain.info.ParentBlockNumber = 100
				_, err = chain.loadParent()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrChainParentBlockNotFound))
			})
		})

		Context("when the chain's parent chain and parent block exist", func() {

			var block types.Block

			BeforeEach(func() {
				block = MakeBlock(bc, chain, sender, receiver)
				err = chain.store.PutBlock(block)
				Expect(err).To(BeNil())
			})

			It("should return parent chain and parent block", func() {
				ch := NewChain("c1", db, cfg, log)
				ch.info.ParentChainID = chain.GetID()
				ch.info.ParentBlockNumber = block.GetNumber()

				result, err := ch.loadParent()
				Expect(err).To(BeNil())
				Expect(result.GetID()).To(Equal(chain.GetID()))
				Expect(result.info).To(Equal(chain.info))
			})
		})

		It("should get parents of chains", func() {

			// parent block of chain2
			block := MakeBlock(bc, chain, sender, receiver)
			err = chain.store.PutBlock(block)
			Expect(err).To(BeNil())

			chain2.info.ParentChainID = chain.GetID()
			chain2.info.ParentBlockNumber = block.GetNumber()
			err = chain2.save()
			Expect(err).To(BeNil())

			// parent block of chain3
			block2 := MakeBlock(bc, chain2, sender, receiver)
			err = chain2.store.PutBlock(block2)
			Expect(err).To(BeNil())

			chain3 := NewChain("p3", db, cfg, log)
			chain3.info.ParentChainID = chain2.GetID()
			chain3.info.ParentBlockNumber = block2.GetNumber()
			err = chain3.save()
			Expect(err).To(BeNil())

			p3Parent, err := chain3.loadParent()
			Expect(err).To(BeNil())
			Expect(p3Parent.info).To(Equal(chain2.info))

			p2Parent, err := chain2.loadParent()
			Expect(err).To(BeNil())
			Expect(p2Parent.info).To(Equal(chain.info))
		})
	})
})
