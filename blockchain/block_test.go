package blockchain

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	. "github.com/ncodes/goblin"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types"

	"github.com/ellcrys/elld/blockchain/common"
	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/gomega"
)

func TestBlock(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Block", func() {

		var err error
		var bc *Blockchain
		var cfg *config.EngineConfig
		var db elldb.DB
		var genesisBlock types.Block
		var genesisChain *Chain
		var sender, receiver *crypto.Key

		g.BeforeEach(func() {
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

		g.BeforeEach(func() {
			genesisBlock, err = LoadBlockFromFile("genesis-test.json")
			Expect(err).To(BeNil())
			bc.SetGenesisBlock(genesisBlock)
			err = bc.Up()
			Expect(err).To(BeNil())
			genesisChain = bc.bestChain
		})

		g.AfterEach(func() {
			db.Close()
			err = os.RemoveAll(cfg.DataDir())
			Expect(err).To(BeNil())
		})

		g.It("", func() {

		})

		g.Describe(".HaveBlock", func() {
			var block types.Block

			g.BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
			})

			g.It("should return false when block does not exist in any known chain", func() {
				has, err := bc.HaveBlock(block.GetHash())
				Expect(err).To(BeNil())
				Expect(has).To(BeFalse())
			})

			g.When("block exist in the chain", func() {

				g.BeforeEach(func() {
					chain := NewChain("chain", db, cfg, log)
					bc.addChain(chain)
					Expect(err).To(BeNil())
					err = chain.append(block)
					Expect(err).To(BeNil())
				})

				g.It("should return true", func() {
					has, err := bc.HaveBlock(block.GetHash())
					Expect(err).To(BeNil())
					Expect(has).To(BeTrue())
				})
			})
		})

		g.Describe(".GetBlock", func() {
			var block types.Block

			g.BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
				err = genesisChain.append(block)
				Expect(err).To(BeNil())
			})

			g.When("block does not exist", func() {
				g.It("should return ErrBlockNotFound", func() {
					_, err := bc.GetBlock(block.GetNumber(), util.StrToHash("invalid"))
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(core.ErrBlockNotFound))
				})
			})

			g.When("block exists", func() {
				g.It("should successfully find the block", func() {
					result, err := bc.GetBlock(block.GetNumber(), block.GetHash())
					Expect(err).To(BeNil())
					Expect(result.GetHash()).To(Equal(block.GetHash()))
				})
			})

			g.Context("with two chains", func() {

				var block3 types.Block

				g.BeforeEach(func() {
					chain2 := NewChain("chain_2", db, cfg, log)
					err = chain2.append(genesisBlock)
					Expect(err).To(BeNil())
					err = chain2.append(block)
					Expect(err).To(BeNil())

					block3 = MakeTestBlock(bc, chain2, &types.GenerateBlockParams{
						Transactions: []types.Transaction{
							core.NewTx(core.TxTypeAlloc, 1, sender.Addr(), sender, "1", "2.36", 1532730724),
						},
						Creator:    sender,
						Nonce:      util.EncodeNonce(2),
						Difficulty: new(big.Int).SetInt64(131072),
					})

					err = genesisChain.append(block3)
					Expect(err).To(BeNil())
				})

				g.It("should successfully find the block", func() {
					result, err := bc.GetBlock(block3.GetNumber(), block3.GetHash())
					Expect(err).To(BeNil())
					Expect(result.GetHash()).To(Equal(block3.GetHash()))
				})
			})
		})

		g.Describe(".GetBlockByHash", func() {
			var block types.Block

			g.BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions: []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730724),
					},
					Creator:    sender,
					Nonce:      util.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				err = genesisChain.append(block)
				Expect(err).To(BeNil())
			})

			g.It("should return ErrBlockNotFound if not found in any chain", func() {
				_, err := bc.GetBlockByHash(util.StrToHash("invalid"))
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBlockNotFound))
			})

			g.It("should successfully find the block", func() {
				result, err := bc.GetBlockByHash(block.GetHash())
				Expect(err).To(BeNil())
				Expect(result.GetHash()).To(Equal(block.GetHash()))
			})

			g.Context("with two chains", func() {

				var block3 types.Block

				g.BeforeEach(func() {
					chain2 := NewChain("chain_2", db, cfg, log)
					err = chain2.append(genesisBlock)
					Expect(err).To(BeNil())
					err = chain2.append(block)
					Expect(err).To(BeNil())

					block3 = MakeTestBlock(bc, chain2, &types.GenerateBlockParams{
						Transactions: []types.Transaction{
							core.NewTx(core.TxTypeAlloc, 1, sender.Addr(), sender, "1", "2.36", 1532730724),
						},
						Creator:    sender,
						Nonce:      util.EncodeNonce(2),
						Difficulty: new(big.Int).SetInt64(131072),
					})

					err = genesisChain.append(block3)
					Expect(err).To(BeNil())
				})

				g.It("should successfully find the block", func() {
					result, err := bc.GetBlockByHash(block3.GetHash())
					Expect(err).To(BeNil())
					Expect(result.GetHash()).To(Equal(block3.GetHash()))
				})
			})
		})

		g.Describe(".IsKnownBlock", func() {
			var block types.Block

			g.BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions: []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730724),
					},
					Creator:    sender,
					Nonce:      util.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			g.It("should return false when block does not exist in any known chain or caches", func() {
				exist, reason, err := bc.IsKnownBlock(block.GetHash())
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
				Expect(reason).To(BeEmpty())
			})

			g.It("should return true when block exists in a chain", func() {
				chain2 := NewChain("chain2", db, cfg, log)
				Expect(err).To(BeNil())
				err = chain2.append(block)
				Expect(err).To(BeNil())

				bc.addChain(chain2)
				err = chain2.store.PutBlock(block)
				Expect(err).To(BeNil())

				has, err := bc.HaveBlock(block.GetHash())
				Expect(err).To(BeNil())
				Expect(has).To(BeTrue())
			})

			g.It("should return true when block exist as an orphan", func() {
				bc.addOrphanBlock(block)
				known, reason, err := bc.IsKnownBlock(block.GetHash())
				Expect(err).To(BeNil())
				Expect(known).To(BeTrue())
				Expect(reason).To(Equal("orphan cache"))
			})
		})

		g.Describe(".Generate", func() {

			var txs []types.Transaction

			g.BeforeEach(func() {
				bc.bestChain = genesisChain
				txs = []types.Transaction{core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "0.1", "2.38", time.Now().Unix())}
			})

			g.It("should validate params", func() {
				var cases = map[*types.GenerateBlockParams]interface{}{
					{Transactions: txs}:                  fmt.Errorf("creator's key is required"),
					{Transactions: txs, Creator: sender}: fmt.Errorf("difficulty is required"),
				}

				for m, r := range cases {
					_, err = bc.Generate(m)
					Expect(err).To(Equal(r))
				}
			})

			g.It("should successfully create a new and valid block", func() {
				blk, err := bc.Generate(&types.GenerateBlockParams{
					Transactions: txs,
					Creator:      sender,
					Nonce:        util.EncodeNonce(1),
					Difficulty:   new(big.Int).SetInt64(131072),
				})
				Expect(err).To(BeNil())
				Expect(blk).ToNot(BeNil())
				Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
				Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(2)))
				Expect(blk.GetHeader().GetParentHash()).To(Equal(genesisBlock.GetHash()))
			})

			g.When("transaction pool contains transactions", func() {

				var tx, tx2 types.Transaction

				g.BeforeEach(func() {
					tx = core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "0.1", "2.38", time.Now().Unix())
					tx.SetHash(tx.ComputeHash())
					tx2 = core.NewTx(core.TxTypeBalance, 2, receiver.Addr(), sender, "0.1", "2.38", time.Now().Unix()+100)
					tx2.SetHash(tx2.ComputeHash())
					err = bc.txPool.Put(tx)
					Expect(err).To(BeNil())
					err = bc.txPool.Put(tx2)
					Expect(err).To(BeNil())
				})

				g.It("should successfully create a new and valid block with transactions from the transaction pool", func() {
					blk, err := bc.Generate(&types.GenerateBlockParams{Creator: sender, Nonce: util.EncodeNonce(1), Difficulty: new(big.Int).SetInt64(131072)})
					Expect(err).To(BeNil())
					Expect(blk).ToNot(BeNil())
					Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
					Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(2)))
					Expect(blk.GetHeader().GetParentHash()).To(Equal(genesisBlock.GetHash()))
					Expect(blk.GetTransactions()).To(HaveLen(2))
					Expect(blk.GetTransactions()[0]).To(Equal(tx))
					Expect(blk.GetTransactions()[1]).To(Equal(tx2))
				})
			})

			g.When("chain is directly passed", func() {
				g.It("should successfully create a new and valid block", func() {
					blk, err := bc.Generate(&types.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        util.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, &common.ChainerOp{Chain: genesisChain})
					Expect(err).To(BeNil())
					Expect(blk).ToNot(BeNil())
					Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
					Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(2)))
					Expect(blk.GetHeader().GetParentHash()).To(Equal(genesisBlock.GetHash()))
				})
			})

			g.When("best chain is nil and no chain is passed directly", func() {

				g.BeforeEach(func() {
					bc.bestChain = nil
				})

				g.It("should return error if not target chain", func() {
					blk, err := bc.Generate(&types.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        util.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					})
					Expect(err).ToNot(BeNil())
					Expect(blk).To(BeNil())
					Expect(err.Error()).To(Equal("target chain not set"))
				})
			})

			g.When("target chain state does not include the sender account", func() {

				var targetChain *Chain

				g.BeforeEach(func() {
					targetChain = NewChain("abc", db, cfg, log)
					targetChain.parentBlock = genesisBlock
				})

				g.It("should return error sender account is not found in the target chain", func() {
					blk, err := bc.Generate(&types.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        util.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, &common.ChainerOp{Chain: targetChain})
					Expect(err).ToNot(BeNil())
					Expect(blk).To(BeNil())
					Expect(err.Error()).To(Equal("exec: transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})

			g.When("target chain has no block but has a parent block attached", func() {

				var targetChain *Chain

				g.BeforeEach(func() {
					targetChain = NewChain("abc", db, cfg, log)
					targetChain.parentBlock = genesisBlock
				})

				g.BeforeEach(func() {
					err = bc.CreateAccount(1, targetChain, &core.Account{
						Type:    core.AccountTypeBalance,
						Address: sender.Addr(),
						Balance: "100",
					})
					Expect(err).To(BeNil())
				})

				g.It("should successfully create a new and valid block", func() {
					blk, err := bc.Generate(&types.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        util.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, &common.ChainerOp{Chain: targetChain})
					Expect(err).To(BeNil())
					Expect(blk).ToNot(BeNil())
					Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
					Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(2)))
					Expect(blk.GetHeader().GetParentHash()).To(Equal(genesisBlock.GetHash()))
				})
			})

			g.When("target chain has no tip block and no parent block", func() {

				var targetChain *Chain

				g.BeforeEach(func() {
					targetChain = NewChain("abc", db, cfg, log)
				})

				g.BeforeEach(func() {
					err = bc.CreateAccount(1, targetChain, &core.Account{
						Type:    core.AccountTypeBalance,
						Address: sender.Addr(),
						Balance: "100",
					})
					Expect(err).To(BeNil())
				})

				g.It("should create a 'genesis' block", func() {
					blk, err := bc.Generate(&types.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        util.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, &common.ChainerOp{Chain: targetChain})
					Expect(err).To(BeNil())
					Expect(blk).ToNot(BeNil())
					Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
					Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(1)))
					Expect(blk.GetHeader().GetParentHash().IsEmpty()).To(BeTrue())
				})
			})

			g.When("fee allocation is enabled", func() {
				g.It("should successfully add an allocation allocation transaction", func() {
					blk, err := bc.Generate(&types.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        util.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
						AddFeeAlloc:  true,
					})
					Expect(err).To(BeNil())

					nTxs := blk.GetTransactions()

					g.Describe("that there are 2 transactions", func() {
						Expect(nTxs).To(HaveLen(2))
					})

					g.Describe("that the last transaction to be TxTypeAlloc", func() {
						Expect(nTxs[1].GetType()).To(Equal(core.TxTypeAlloc))
					})

					g.Describe("that the allocation value is the total fee of all txs", func() {
						Expect(nTxs[1].GetValue().Decimal()).To(Equal(txs[0].GetFee().Decimal()))
					})
				})
			})
		})
	})
}
