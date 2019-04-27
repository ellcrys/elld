package blockchain

import (
	"math/big"
	"os"
	"time"

	"github.com/ellcrys/elld/params"

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReOrg", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock types.Block
	var genesisChain *Chain
	var sender, receiver *crypto.Key

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.NetDataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
		bc.SetCoinbase(crypto.NewKeyFromIntSeed(1234))
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

	Describe(".chooseBestChain", func() {

		var chainA, chainB *Chain

		BeforeEach(func() {
			genesisChainBlock2 := MakeBlockWithTotalDifficulty(bc, genesisChain, sender,
				new(big.Int).SetInt64(10))
			err := genesisChain.append(genesisChainBlock2)
			Expect(err).To(BeNil())
		})

		Context("test difficulty rule", func() {

			When("chainA has the most total difficulty", func() {

				BeforeEach(func() {
					chainA = NewChain("chain_a", db, cfg, log)
					err := bc.saveChain(chainA, "", 0)
					Expect(err).To(BeNil())

					chainABlock1 := MakeBlockWithTotalDifficulty(bc, genesisChain, sender,
						new(big.Int).SetInt64(100))

					err = chainA.append(chainABlock1)
					Expect(err).To(BeNil())
				})

				It("should return chainA as the best chain since it has a higher total difficulty than the genesis chain", func() {
					bc.bestChain = nil
					Expect(bc.chains).To(HaveLen(2))
					bestChain, err := bc.chooseBestChain()
					Expect(err).To(BeNil())
					Expect(bestChain.id).To(Equal(chainA.id))
				})
			})

			When("chainB has the lowest total difficulty", func() {
				BeforeEach(func() {
					chainB = NewChain("chain_b", db, cfg, log)
					err := bc.saveChain(chainB, "", 0)
					Expect(err).To(BeNil())

					chainBBlock1 := MakeBlockWithTotalDifficulty(bc, genesisChain, sender,
						new(big.Int).SetInt64(5))

					err = chainB.append(chainBBlock1)
					Expect(err).To(BeNil())
				})

				It("should return genesis chain as the best chain since it has a higher total difficulty than chainB", func() {
					bc.bestChain = nil
					Expect(bc.chains).To(HaveLen(2))
					bestChain, err := bc.chooseBestChain()
					Expect(err).To(BeNil())
					Expect(bestChain.id).To(Equal(genesisChain.id))
				})
			})
		})

		Context("test oldest chain rule", func() {

			When("chainA and genesis chain have the same total difficulty but the genesis chain is older", func() {

				BeforeEach(func() {
					chainA = NewChain("chain_a", db, cfg, log)
					err := bc.saveChain(chainA, "", 0)
					Expect(err).To(BeNil())

					chainABlock1 := MakeBlockWithTotalDifficulty(bc, genesisChain, sender,
						new(big.Int).SetInt64(10))

					err = chainA.append(chainABlock1)
					Expect(err).To(BeNil())
				})

				It("should return genesis chain as the best chain since it has an older chain timestamp", func() {
					bc.bestChain = nil
					Expect(bc.chains).To(HaveLen(2))
					bestChain, err := bc.chooseBestChain()
					Expect(err).To(BeNil())
					Expect(bestChain.id).To(Equal(genesisChain.id))
				})
			})

		})

		Context("test largest point address rule", func() {
			When("chainA and genesis chain have the same total difficulty and chain age", func() {

				BeforeEach(func() {
					chainA = NewChain("chain_a", db, cfg, log)
					chainA.info.Timestamp = genesisChain.info.Timestamp
					err := bc.saveChain(chainA, "", 0)
					Expect(err).To(BeNil())

					chainABlock1 := MakeBlockWithTotalDifficulty(bc, genesisChain, sender,
						new(big.Int).SetInt64(10))

					err = chainA.append(chainABlock1)
					Expect(err).To(BeNil())
				})

				It("should return the chain with the largest pointer address", func() {
					bc.bestChain = nil
					Expect(bc.chains).To(HaveLen(2))
					bestChain, err := bc.chooseBestChain()
					Expect(err).To(BeNil())
					delete(bc.chains, bestChain.id)
					for _, leastChain := range bc.chains {
						Expect(util.GetPtrAddr(leastChain).Cmp(util.GetPtrAddr(bestChain))).To(Equal(-1))
					}
				})
			})
		})
	})

	Describe(".reOrg: long chain to short chain", func() {

		var forkedChain *Chain

		// Build two chains having the following shapes:
		// [1]-[2]-[3]-[4] 	- Genesis chain
		//  |__[2] 			- forked chain 1
		BeforeEach(func() {
			// genesis block 2
			genesisB2 := MakeBlockWithTx(bc, genesisChain, sender, 1)

			forkChainB2 := MakeBlockWithTx(bc, genesisChain, sender, 1)

			_, err = bc.ProcessBlock(genesisB2)
			Expect(err).To(BeNil())

			// process the forked block. It must create a new chain
			forkedChainReader, err := bc.ProcessBlock(forkChainB2)
			Expect(err).To(BeNil())
			Expect(len(bc.chains)).To(Equal(2))
			forkedChain = bc.chains[forkedChainReader.GetID()]

			// genesis block 3
			genesisB3 := MakeBlockWithTx(bc, genesisChain, sender, 2)
			_, err = bc.ProcessBlock(genesisB3)
			Expect(err).To(BeNil())

			// genesis block 4
			genesisB4 := MakeBlockWithTx(bc, genesisChain, sender, 3)
			_, err = bc.ProcessBlock(genesisB4)
			Expect(err).To(BeNil())
		})

		// verify chains shape
		BeforeEach(func() {
			tip, _ := genesisChain.Current()
			Expect(tip.GetNumber()).To(Equal(uint64(4)))
			Expect(genesisChain.GetParent()).To(BeNil())

			forkTip, _ := bc.chains[forkedChain.GetID()].Current()
			Expect(forkTip.GetNumber()).To(Equal(uint64(2)))
			Expect(genesisChain.GetParent()).To(BeNil())
		})

		It("should return error if branch chain is empty", func() {
			branch := NewChain("empty_chain", db, cfg, log)
			_, err := bc.reOrg(genesisChain, branch)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to get branch chain tip: block not found"))
		})

		It("should return error if main/best chain is empty", func() {
			branch := NewChain("empty_chain", db, cfg, log)
			_, err := bc.reOrg(branch, branch)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to get best chain tip: block not found"))
		})

		It("should return error if branch chain does not have a parent block set", func() {
			forkedChain.parentBlock = nil
			_, err := bc.reOrg(genesisChain, bc.chains[forkedChain.GetID()])
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("parent block not set on branch"))
		})

		When("branch chain parent block does not exist on the main chain", func() {
			var chain *Chain
			var parentBlock types.Block

			BeforeEach(func() {
				// make parent block and set the hash to something else
				// so that a query for it in the main chain fails.
				parentBlock = MakeBlock(bc, genesisChain, sender, receiver)
				parentBlock.SetHash(util.StrToHash("abc"))

				// create a new chain, set the required account to
				// allow block creation possible. Add the block to the chain.
				chain = NewChain("ch1", db, cfg, log)
				err := bc.CreateAccount(1, chain, &core.Account{
					Type:    core.AccountTypeBalance,
					Address: util.String(sender.Addr()),
					Balance: "100",
				})
				Expect(err).To(BeNil())
				block := MakeBlock(bc, chain, sender, receiver)
				err = chain.append(block)
				Expect(err).To(BeNil())

				// set the parent block the chain the parent block  (with unknown hash)
				chain.parentBlock = parentBlock
			})

			It("should return `parent block does not exist on the main chain`", func() {
				_, err := bc.reOrg(genesisChain, chain)
				Expect(err).To(Equal(params.ErrBranchParentNotInMainChain))
			})
		})

		It("should return error when branch chain's parent does not exist on the main chain", func() {
			forkedChain.parentBlock = nil
			_, err := bc.reOrg(genesisChain, bc.chains[forkedChain.GetID()])
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("parent block not set on branch"))
		})

		Describe("when successful", func() {

			var reOrgedChain *Chain
			var err error

			BeforeEach(func() {
				reOrgedChain, err = bc.reOrg(genesisChain, forkedChain)
				Expect(err).To(BeNil())
			})

			It("re-orged chain should have same length as side/fork chain", func() {
				reOrgedHeight, err := reOrgedChain.height()
				Expect(err).To(BeNil())
				forkedChainHeight, err := forkedChain.height()
				Expect(err).To(BeNil())
				Expect(reOrgedHeight).To(Equal(forkedChainHeight))
			})

			It("re-orged chain tip must equal side/fork chain tip", func() {
				reOrgedTip, err := reOrgedChain.Current()
				Expect(err).To(BeNil())
				forkedChainTip, err := reOrgedChain.Current()
				Expect(err).To(BeNil())
				Expect(reOrgedTip).To(Equal(forkedChainTip))
			})
		})
	})

	Describe(".reOrg: short chain to long chain", func() {

		var forkedChain *Chain

		// Build two chains having the following shapes:
		// [1]-[2] 			- Genesis chain
		//  |__[2]-[3]-[4] 	- forked chain 1
		BeforeEach(func() {

			// genesis block 2
			genesisB2 := MakeBlockWithTx(bc, genesisChain, sender, 1)

			// forked chain block 2
			forkChainB2 := MakeBlockWithTx(bc, genesisChain, sender, 1)

			_, err = bc.ProcessBlock(genesisB2)
			Expect(err).To(BeNil())

			// process the forked block. It must create a new chain
			forkedChainReader, err := bc.ProcessBlock(forkChainB2, common.OpAllowExec(true))
			Expect(err).To(BeNil())
			Expect(len(bc.chains)).To(Equal(2))
			forkedChain = bc.chains[forkedChainReader.GetID()]

			// forked chain block 3
			forkChainB3 := MakeBlockWithTx(bc, forkedChain, sender, 2)
			_, err = bc.ProcessBlock(forkChainB3, common.OpAllowExec(true))
			Expect(err).To(BeNil())

			// forked chain block 4
			forkedChainB4 := MakeBlockWithTx(bc, forkedChain, sender, 3)
			_, err = bc.ProcessBlock(forkedChainB4, common.OpAllowExec(true))
			Expect(err).To(BeNil())
		})

		// verify chains shape
		BeforeEach(func() {
			tip, _ := genesisChain.Current()
			Expect(tip.GetNumber()).To(Equal(uint64(2)))
			Expect(genesisChain.GetParent()).To(BeNil())

			forkTip, _ := forkedChain.Current()
			Expect(forkTip.GetNumber()).To(Equal(uint64(4)))
			Expect(forkedChain.GetParent()).To(Equal(genesisChain))
			Expect(forkedChain.GetParentBlock().GetNumber()).To(Equal(uint64(1)))
		})

		It("should be successful; return nil", func() {
			reOrgedChain, err := bc.reOrg(genesisChain, forkedChain)
			Expect(err).To(BeNil())

			Describe("reorged chain should have same length as side/fork chain", func() {
				reOrgedHeight, err := reOrgedChain.height()
				Expect(err).To(BeNil())
				forkedChainHeight, err := forkedChain.height()
				Expect(err).To(BeNil())
				Expect(reOrgedHeight).To(Equal(forkedChainHeight))
			})

			Describe("reorged chain tip must equal side/fork chain tip", func() {
				reOrgedTip, err := reOrgedChain.Current()
				Expect(err).To(BeNil())
				forkedChainTip, err := reOrgedChain.Current()
				Expect(err).To(BeNil())
				Expect(reOrgedTip).To(Equal(forkedChainTip))
			})
		})
	})

	Describe(".decideBestChain", func() {

		Describe("when no chain exist", func() {
			BeforeEach(func() {
				bc.chains = map[util.String]*Chain{}
			})

			It("should ruturn err='unable to choose best chain: no chain was proposed'", func() {
				err := bc.decideBestChain()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to choose best chain: no chain was proposed"))
			})
		})

		Describe("when main/best chain and proposed chain are the same", func() {
			It("should return nil and best chain remains unchanged", func() {
				mainChain := bc.bestChain
				err := bc.decideBestChain()
				Expect(err).To(BeNil())
				Expect(mainChain).To(Equal(bc.bestChain))
			})
		})

		When(".reOrg: main chain and proposed chain are not directly related", func() {

			var forkedChain, forkedChain2, forkedChain3 *Chain

			// Build two chains having the following shapes:
			// [1]-[2] 						- Genesis chain,  TotalDifficulty = 10 (main chain)
			//  |__[2]-[3]-[4] 				- forked chain 1, TotalDifficulty = 20
			//          |__[4]-[5]			- forked chain 2, TotalDifficulty = 30
			//              |__[5]-[6] 		- forked chain 3, TotalDifficulty = 40 (proposed chain)
			BeforeEach(func() {
				bc.setSkipDecideBestChain(true)

				// genesis block 2
				genesisB2 := MakeBlockWithTDAndNonce(bc, genesisChain, sender, 1,
					new(big.Int).SetInt64(10))

				// forked chain block 2
				forkChainB2 := MakeBlockWithTx(bc, genesisChain, sender, 1)

				_, err = bc.ProcessBlock(genesisB2)
				Expect(err).To(BeNil())

				// process the forkChainB2 block to create new chain
				forkedChainReader, err := bc.ProcessBlock(forkChainB2, common.OpAllowExec(true))
				Expect(err).To(BeNil())
				Expect(len(bc.chains)).To(Equal(2))
				forkedChain = bc.chains[forkedChainReader.GetID()]

				// forked chain block 3
				forkChainB3 := MakeBlockWithTx(bc, forkedChain, sender, 2)
				_, err = bc.ProcessBlock(forkChainB3, common.OpAllowExec(true))
				Expect(err).To(BeNil())

				// forked chain block 4
				forkChainB4 := MakeBlockWithTDAndNonce(bc, forkedChain, sender, 3,
					new(big.Int).SetInt64(20))

				// forked chain 2, block 4
				forkChain2B4 := MakeBlockWithTx(bc, forkedChain, sender, 3)

				_, err = bc.ProcessBlock(forkChainB4)
				Expect(err).To(BeNil())

				// process the forkChain2B4 block to create new chain
				forkedChain2Reader, err := bc.ProcessBlock(forkChain2B4, common.OpAllowExec(true))
				Expect(err).To(BeNil())
				Expect(len(bc.chains)).To(Equal(3))
				forkedChain2 = bc.chains[forkedChain2Reader.GetID()]

				// forked chain 2, block 5
				forkChain2B5 := MakeBlockWithTDAndNonce(bc, forkedChain2, sender, 4,
					new(big.Int).SetInt64(30))

				// forked chain 3, block 5
				forkChain3B5 := MakeBlockWithTx(bc, forkedChain2, sender, 4)

				_, err = bc.ProcessBlock(forkChain2B5)
				Expect(err).To(BeNil())

				// process the forkChain3B5 block to create new chain
				forkedChain3Reader, err := bc.ProcessBlock(forkChain3B5, common.OpAllowExec(true))
				Expect(err).To(BeNil())
				Expect(len(bc.chains)).To(Equal(4))
				forkedChain3 = bc.chains[forkedChain3Reader.GetID()]

				// forked chain 3, block 6
				forkChain3B6 := MakeBlockWithTDAndNonce(bc, forkedChain3, sender, 5,
					new(big.Int).SetInt64(40))
				_, err = bc.ProcessBlock(forkChain3B6)
				Expect(err).To(BeNil())
			})

			// verify chains shape
			BeforeEach(func() {
				tip, _ := genesisChain.Current()
				Expect(tip.GetNumber()).To(Equal(uint64(2)))
				Expect(tip.GetTotalDifficulty().Int64()).To(Equal(int64(10)))
				Expect(genesisChain.GetParent()).To(BeNil())

				forkTip, _ := forkedChain.Current()
				Expect(forkTip.GetNumber()).To(Equal(uint64(4)))
				Expect(forkTip.GetTotalDifficulty().Int64()).To(Equal(int64(20)))
				Expect(forkedChain.GetParent()).To(Equal(genesisChain))
				Expect(forkedChain.GetParentBlock().GetNumber()).To(Equal(uint64(1)))

				fork2Tip, _ := forkedChain2.Current()
				Expect(fork2Tip.GetNumber()).To(Equal(uint64(5)))
				Expect(fork2Tip.GetTotalDifficulty().Int64()).To(Equal(int64(30)))
				Expect(forkedChain2.GetParent()).To(Equal(forkedChain))
				Expect(forkedChain2.GetParentBlock().GetNumber()).To(Equal(uint64(3)))

				fork3Tip, _ := forkedChain3.Current()
				Expect(fork3Tip.GetNumber()).To(Equal(uint64(6)))
				Expect(fork3Tip.GetTotalDifficulty().Int64()).To(Equal(int64(40)))
				Expect(forkedChain3.GetParent()).To(Equal(forkedChain2))
				Expect(forkedChain3.GetParentBlock().GetNumber()).To(Equal(uint64(4)))

				bestChain, err := bc.chooseBestChain()
				Expect(err).To(BeNil())
				Expect(bestChain).To(Equal(forkedChain3))
			})

			BeforeEach(func() {
				bc.setSkipDecideBestChain(false)
				err := bc.decideBestChain()
				Expect(err).To(BeNil())
			})

			Specify("that forked chain 2 is re-orged with fork chain 3", func() {
				tip, err := forkedChain2.Current()
				Expect(err).To(BeNil())
				Expect(tip.GetNumber()).To(Equal(uint64(6)))
			})

			Specify("that forked chain 1 is re-orged with fork chain 2", func() {
				tip, err := forkedChain.Current()
				Expect(err).To(BeNil())
				tip2, err := forkedChain2.Current()
				Expect(err).To(BeNil())
				Expect(tip.GetNumber()).To(Equal(tip2.GetNumber()))
			})

			Specify("that main chain is re-orged with fork chain 1", func() {
				Expect(bc.bestChain).To(Equal(genesisChain))
				tip, err := genesisChain.Current()
				Expect(err).To(BeNil())
				tip2, err := forkedChain.Current()
				Expect(err).To(BeNil())
				Expect(tip.GetNumber()).To(Equal(tip2.GetNumber()))
			})
		})
	})

	Describe(".recordReOrg", func() {

		var branch *Chain

		BeforeEach(func() {
			branch = NewChain("s1", db, cfg, log)
			err := branch.append(genesisBlock)
			branch.parentBlock = genesisBlock
			Expect(err).To(BeNil())
		})

		It("should successfully store re-org info", func() {
			now := time.Now()
			err := bc.recordReOrg(now.UnixNano(), branch)
			Expect(err).To(BeNil())
		})
	})

	Describe(".getReOrgs", func() {
		var branch *Chain

		BeforeEach(func() {
			branch = NewChain("s1", db, cfg, log)
			err := branch.append(genesisBlock)
			branch.parentBlock = genesisBlock
			Expect(err).To(BeNil())
		})

		It("should get two re-orgs sorted by timestamp in decending order", func() {
			err := bc.recordReOrg(time.Now().UnixNano(), branch)
			Expect(err).To(BeNil())

			bc.recordReOrg(time.Now().UnixNano(), branch)
			Expect(err).To(BeNil())

			reOrgs := bc.getReOrgs()
			Expect(reOrgs).To(HaveLen(2))
			Expect(reOrgs[0].Timestamp > reOrgs[1].Timestamp).To(BeTrue())
		})
	})

})
