package blockchain

import (
	"math/big"
	"os"
	"time"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReOrg", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock core.Block
	var genesisChain *Chain
	var sender, receiver *crypto.Key

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.DataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
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
			genesisChainBlock2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
				},
				Creator:                 sender,
				Nonce:                   util.EncodeNonce(1),
				Difficulty:              new(big.Int).SetInt64(1),
				OverrideTotalDifficulty: new(big.Int).SetInt64(10),
			})
			err := genesisChain.append(genesisChainBlock2)
			Expect(err).To(BeNil())
		})

		Context("test difficulty rule", func() {

			When("chainA has the most total difficulty", func() {

				BeforeEach(func() {
					chainA = NewChain("chain_a", db, cfg, log)
					err := bc.saveChain(chainA, "", 0)
					Expect(err).To(BeNil())

					chainABlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730724),
						},
						Creator:                 sender,
						Nonce:                   util.EncodeNonce(1),
						Difficulty:              new(big.Int).SetInt64(1),
						OverrideTotalDifficulty: new(big.Int).SetInt64(100),
					})

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

					chainBBlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730726),
						},
						Creator:                 sender,
						Nonce:                   util.EncodeNonce(1),
						Difficulty:              new(big.Int).SetInt64(1),
						OverrideTotalDifficulty: new(big.Int).SetInt64(5),
					})

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

					chainABlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730724),
						},
						Creator:                 sender,
						Nonce:                   util.EncodeNonce(1),
						Difficulty:              new(big.Int).SetInt64(1),
						OverrideTotalDifficulty: new(big.Int).SetInt64(10),
					})

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

					chainABlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730724),
						},
						Creator:                 sender,
						Nonce:                   util.EncodeNonce(1),
						Difficulty:              new(big.Int).SetInt64(1),
						OverrideTotalDifficulty: new(big.Int).SetInt64(10),
					})

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
			genesisB2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730723),
					objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730723),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})

			forkChainB2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
					objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730724),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})
			_, err = bc.ProcessBlock(genesisB2)
			Expect(err).To(BeNil())

			// process the forked block. It must create a new chain
			forkedChainReader, err := bc.ProcessBlock(forkChainB2)
			Expect(err).To(BeNil())
			Expect(len(bc.chains)).To(Equal(2))
			forkedChain = bc.chains[forkedChainReader.GetID()]

			// genesis block 3
			genesisB3 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 2, util.String(receiver.Addr()), sender, "1", "2.5", 1532730725),
					objects.NewTx(objects.TxTypeAlloc, 2, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})
			_, err = bc.ProcessBlock(genesisB3)
			Expect(err).To(BeNil())

			// genesis block 4
			genesisB4 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 3, util.String(receiver.Addr()), sender, "1", "2.5", 1532730726),
					objects.NewTx(objects.TxTypeAlloc, 3, util.String(sender.Addr()), sender, "2.5", "0", 1532730726),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})
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
			_, err := bc.reOrg(branch)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to get branch chain tip: block not found"))
		})

		It("should return error if best chain is empty", func() {
			branch := NewChain("empty_chain", db, cfg, log)
			bc.bestChain = branch
			_, err := bc.reOrg(branch)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to get best chain tip: block not found"))
		})

		It("should return error if branch chain does not have a parent block set", func() {
			forkedChain.parentBlock = nil
			_, err := bc.reOrg(bc.chains[forkedChain.GetID()])
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("parent block not set on branch"))
		})

		It("should be successful; return nil", func() {
			reOrgedChain, err := bc.reOrg(forkedChain)
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

	Describe(".reOrg: short chain to long chain", func() {

		var forkedChain *Chain

		// Build two chains having the following shapes:
		// [1]-[2] 			- Genesis chain
		//  |__[2]-[3]-[4] 	- forked chain 1
		BeforeEach(func() {

			// genesis block 2
			genesisB2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730723),
					objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730723),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})

			// forked chain block 2
			forkChainB2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
					objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730724),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})

			_, err = bc.ProcessBlock(genesisB2)
			Expect(err).To(BeNil())

			// process the forked block. It must create a new chain
			forkedChainReader, err := bc.ProcessBlock(forkChainB2)
			Expect(err).To(BeNil())
			Expect(len(bc.chains)).To(Equal(2))
			forkedChain = bc.chains[forkedChainReader.GetID()]

			// forked chain block 3
			forkChainB3 := MakeTestBlock(bc, forkedChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 2, util.String(receiver.Addr()), sender, "1", "2.5", 1532730725),
					objects.NewTx(objects.TxTypeAlloc, 0, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})
			_, err = bc.ProcessBlock(forkChainB3)
			Expect(err).To(BeNil())

			// forked chain block 4
			forkedChainB4 := MakeTestBlock(bc, forkedChain, &core.GenerateBlockParams{
				Transactions: []core.Transaction{
					objects.NewTx(objects.TxTypeBalance, 3, util.String(receiver.Addr()), sender, "1", "2.5", 1532730726),
					objects.NewTx(objects.TxTypeAlloc, 3, util.String(sender.Addr()), sender, "2.5", "0", 1532730726),
				},
				Creator:    sender,
				Nonce:      util.EncodeNonce(1),
				Difficulty: new(big.Int).SetInt64(131072),
			})
			_, err = bc.ProcessBlock(forkedChainB4)
			Expect(err).To(BeNil())
		})

		// verify chains shape
		BeforeEach(func() {
			tip, _ := genesisChain.Current()
			Expect(tip.GetNumber()).To(Equal(uint64(2)))
			Expect(genesisChain.GetParent()).To(BeNil())

			forkTip, _ := forkedChain.Current()
			Expect(forkTip.GetNumber()).To(Equal(uint64(4)))
			Expect(genesisChain.GetParent()).To(BeNil())
		})

		It("should be successful; return nil", func() {
			reOrgedChain, err := bc.reOrg(forkedChain)
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
