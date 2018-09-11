package blockchain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BlockTest = func() bool {
	return Describe("Block", func() {

		Describe(".HaveBlock", func() {

			var block core.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			It("should return false when block does not exist in any known chain", func() {
				has, err := bc.HaveBlock(block.GetHash())
				Expect(err).To(BeNil())
				Expect(has).To(BeFalse())
			})

			It("should return true of block exists in a chain", func() {
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
		})

		Describe(".GetBlock", func() {
			var block core.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				err = genesisChain.append(block)
				Expect(err).To(BeNil())
			})

			It("should return ErrBlockNotFound if not found in any chain", func() {
				_, err := bc.GetBlock(block.GetNumber(), util.StrToHash("invalid"))
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBlockNotFound))
			})

			It("should successfully find the block", func() {
				result, err := bc.GetBlock(block.GetNumber(), block.GetHash())
				Expect(err).To(BeNil())
				Expect(result.GetHash()).To(Equal(block.GetHash()))
			})

			Context("with two chains", func() {

				var block3 core.Block

				BeforeEach(func() {
					chain2 := NewChain("chain_2", db, cfg, log)
					err = chain2.append(genesisBlock)
					Expect(err).To(BeNil())
					err = chain2.append(block)
					Expect(err).To(BeNil())

					block3 = MakeTestBlock(bc, chain2, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(2),
						Difficulty: new(big.Int).SetInt64(131072),
					})

					err = genesisChain.append(block3)
					Expect(err).To(BeNil())
				})

				It("should successfully find the block", func() {
					result, err := bc.GetBlock(block3.GetNumber(), block3.GetHash())
					Expect(err).To(BeNil())
					Expect(result.GetHash()).To(Equal(block3.GetHash()))
				})
			})
		})

		Describe(".GetBlockByHash", func() {
			var block core.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				err = genesisChain.append(block)
				Expect(err).To(BeNil())
			})

			It("should return ErrBlockNotFound if not found in any chain", func() {
				_, err := bc.GetBlockByHash(util.StrToHash("invalid"))
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBlockNotFound))
			})

			It("should successfully find the block", func() {
				result, err := bc.GetBlockByHash(block.GetHash())
				Expect(err).To(BeNil())
				Expect(result.GetHash()).To(Equal(block.GetHash()))
			})

			Context("with two chains", func() {

				var block3 core.Block

				BeforeEach(func() {
					chain2 := NewChain("chain_2", db, cfg, log)
					err = chain2.append(genesisBlock)
					Expect(err).To(BeNil())
					err = chain2.append(block)
					Expect(err).To(BeNil())

					block3 = MakeTestBlock(bc, chain2, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(2),
						Difficulty: new(big.Int).SetInt64(131072),
					})

					err = genesisChain.append(block3)
					Expect(err).To(BeNil())
				})

				It("should successfully find the block", func() {
					result, err := bc.GetBlockByHash(block3.GetHash())
					Expect(err).To(BeNil())
					Expect(result.GetHash()).To(Equal(block3.GetHash()))
				})
			})
		})

		Describe(".IsKnownBlock", func() {
			var block core.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			It("should return false when block does not exist in any known chain or caches", func() {
				exist, reason, err := bc.IsKnownBlock(block.GetHash())
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
				Expect(reason).To(BeEmpty())
			})

			It("should return true when block exists in a chain", func() {
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

			It("should return true when block exist as an orphan", func() {
				bc.addOrphanBlock(block)
				known, reason, err := bc.IsKnownBlock(block.GetHash())
				Expect(err).To(BeNil())
				Expect(known).To(BeTrue())
				Expect(reason).To(Equal("orphan cache"))
			})
		})

		Describe(".GenerateBlock", func() {

			var txs []core.Transaction

			BeforeEach(func() {
				bc.bestChain = genesisChain
				txs = []core.Transaction{objects.NewTx(objects.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "0.1", "0.1", time.Now().Unix())}
			})

			It("should validate params", func() {
				var cases = map[*core.GenerateBlockParams]interface{}{
					&core.GenerateBlockParams{}:                                   fmt.Errorf("at least one transaction is required"),
					&core.GenerateBlockParams{Transactions: txs}:                  fmt.Errorf("creator's key is required"),
					&core.GenerateBlockParams{Transactions: txs, Creator: sender}: fmt.Errorf("difficulty is required"),
				}

				for m, r := range cases {
					_, err = bc.Generate(m)
					Expect(err).To(Equal(r))
				}
			})

			It("should successfully create a new and valid block", func() {
				blk, err := bc.Generate(&core.GenerateBlockParams{
					Transactions: txs,
					Creator:      sender,
					Nonce:        core.EncodeNonce(1),
					Difficulty:   new(big.Int).SetInt64(131072),
				})
				Expect(err).To(BeNil())
				Expect(blk).ToNot(BeNil())
				Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
				Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(2)))
				Expect(blk.GetHeader().GetParentHash()).To(Equal(genesisBlock.GetHash()))
			})

			When("chain is directly passed", func() {
				It("should successfully create a new and valid block", func() {
					blk, err := bc.Generate(&core.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        core.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, ChainOp{Chain: genesisChain})
					Expect(err).To(BeNil())
					Expect(blk).ToNot(BeNil())
					Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
					Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(2)))
					Expect(blk.GetHeader().GetParentHash()).To(Equal(genesisBlock.GetHash()))
				})
			})

			When("best chain is nil and no chain is passed directly", func() {

				BeforeEach(func() {
					bc.bestChain = nil
				})

				It("should return error if not target chain", func() {
					blk, err := bc.Generate(&core.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        core.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					})
					Expect(err).ToNot(BeNil())
					Expect(blk).To(BeNil())
					Expect(err.Error()).To(Equal("target chain not set"))
				})
			})

			When("target chain state does not include the sender account", func() {

				var targetChain *Chain

				BeforeEach(func() {
					targetChain = NewChain("abc", db, cfg, log)
					targetChain.parentBlock = genesisBlock
				})

				It("should return error sender account is not found in the target chain", func() {
					blk, err := bc.Generate(&core.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        core.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, ChainOp{Chain: targetChain})
					Expect(err).ToNot(BeNil())
					Expect(blk).To(BeNil())
					Expect(err.Error()).To(Equal("exec: transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})

			When("target chain has no block but has a parent block attached", func() {

				var targetChain *Chain

				BeforeEach(func() {
					targetChain = NewChain("abc", db, cfg, log)
					// block.GetHeader().GetParentHash() = "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2"
					targetChain.parentBlock = genesisBlock
				})

				BeforeEach(func() {
					err = bc.putAccount(1, targetChain, &objects.Account{
						Type:    objects.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "100",
					})
					Expect(err).To(BeNil())
				})

				It("should successfully create a new and valid block", func() {
					blk, err := bc.Generate(&core.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        core.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, ChainOp{Chain: targetChain})
					Expect(err).To(BeNil())
					Expect(blk).ToNot(BeNil())
					Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
					Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(2)))
					Expect(blk.GetHeader().GetParentHash()).To(Equal(genesisBlock.GetHash()))
				})
			})

			When("target chain has no tip block and no parent block", func() {

				var targetChain *Chain

				BeforeEach(func() {
					targetChain = NewChain("abc", db, cfg, log)
				})

				BeforeEach(func() {
					err = bc.putAccount(1, targetChain, &objects.Account{
						Type:    objects.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "100",
					})
					Expect(err).To(BeNil())
				})

				It("should create a 'genesis' block", func() {
					blk, err := bc.Generate(&core.GenerateBlockParams{
						Transactions: txs,
						Creator:      sender,
						Nonce:        core.EncodeNonce(1),
						Difficulty:   new(big.Int).SetInt64(131072),
					}, ChainOp{Chain: targetChain})
					Expect(err).To(BeNil())
					Expect(blk).ToNot(BeNil())
					Expect(blk.GetHeader().GetStateRoot()).ToNot(BeEmpty())
					Expect(blk.GetHeader().GetNumber()).To(Equal(uint64(1)))
					Expect(blk.GetHeader().GetParentHash().IsEmpty()).To(BeTrue())
				})
			})
		})
	})
}
