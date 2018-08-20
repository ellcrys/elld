package blockchain

import (
	"math/big"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var WorldReaderTest = func() bool {
	return Describe("WorldReader", func() {

		var wr *WorldReader

		BeforeEach(func() {
			wr = bc.NewWorldReader()
			_ = Expect
		})

		// save the genesis chain
		BeforeEach(func() {
			err = bc.saveChain(genesisChain, "", 0)
			Expect(err).To(BeNil())
		})

		Describe(".GetAccount as user", func() {

			It("should set reader to ReaderUser if chain is not provided", func() {
				wr.GetAccount(nil, util.String(sender.Addr()))
				Expect(wr.reader).To(Equal(ReaderUser))
			})

			It("should return error if best chain is not set", func() {
				bc.bestChain = nil
				_, err := wr.GetAccount(nil, util.String(sender.Addr()))
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("no best chain yet"))
			})

			It("should successfully get account", func() {
				account, err := wr.GetAccount(nil, util.String(sender.Addr()))
				Expect(err).To(BeNil())
				Expect(account).ToNot(BeNil())
			})

			It("should return err if account is not found", func() {
				_, err := wr.GetAccount(nil, util.String("abc"))
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrAccountNotFound))
			})
		})

		Describe(".GetAccount as miner", func() {

			When("chain has no parent", func() {
				It("should successfully get account", func() {
					result, err := wr.GetAccount(genesisChain, util.String(sender.Addr()))
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())
				})
			})

			When("chain has parent block in another chain", func() {

				var sidechain1, sidechain2 *Chain
				var genesisChainBlocks []core.Block
				var sidechain1Blocks []core.Block
				var sidechain2Blocks []core.Block

				// create a sidechain 1 and sidechain 2
				// set its parent chain to the genesis chain
				// set its parent block number block 2 of the parent
				//
				// Target blockchain shape:
				//
				// [B1]-[B2]-[B3]-[B4] - Genesis chain shape
				//       |___[B3]-[B4] - Side chain 1
				//             |__[B4] - Side chain 2
				BeforeEach(func() {
					sidechain1 = NewChain("c1", db, cfg, log)
					err = bc.saveChain(sidechain1, genesisChain.id, 2)
					Expect(err).To(BeNil())

					sidechain2 = NewChain("c2", db, cfg, log)
					err = bc.saveChain(sidechain2, sidechain1.id, 3)
					Expect(err).To(BeNil())
				})

				// extend the genesis chain by 4 blocks
				// duplicate the
				BeforeEach(func() {

					// block 2
					genesisChainBlocks = append(genesisChainBlocks, MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 191, util.String(sender.Addr()), sender, "1", "0.1", 1532730723),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					}))
					err = genesisChain.append(genesisChainBlocks[0])
					Expect(err).To(BeNil())

					// block 3
					genesisChainBlocks = append(genesisChainBlocks, MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 191, util.String(sender.Addr()), sender, "1", "0.1", 1532730723),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					}))

					// create the forked sidechain block
					// block 3
					sidechain1Blocks = append(sidechain1Blocks, MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 191, util.String(sender.Addr()), sender, "2", "0.1", 1532730723),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					}))

					err = genesisChain.append(genesisChainBlocks[1])
					Expect(err).To(BeNil())

					err = sidechain1.append(sidechain1Blocks[0])
					Expect(err).To(BeNil())

					// block 4
					genesisChainBlocks = append(genesisChainBlocks, MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 191, util.String(sender.Addr()), sender, "1", "0.1", 1532730723),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					}))
					err = genesisChain.append(genesisChainBlocks[2])
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {

					// block 4
					sidechain1Blocks = append(sidechain1Blocks, MakeTestBlock(bc, sidechain1, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 191, util.String(sender.Addr()), sender, "2", "0.1", 1532730723),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					}))

					// create the forked sidechain 2 block
					// block 4
					sidechain2Blocks = append(sidechain2Blocks, MakeTestBlock(bc, sidechain1, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 191, util.String(sender.Addr()), sender, "2", "0.1", 1532730723),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					}))
					err = sidechain2.append(sidechain2Blocks[0])
					Expect(err).To(BeNil())

					err = sidechain1.append(sidechain1Blocks[1])
					Expect(err).To(BeNil())
				})

				// Ensure the target blockchain shape
				// is as expected.
				// Target blockchain shape:
				//
				// [B1]-[B2]-[B3]-[B4] - Genesis chain shape
				//       |___[B3]-[B4] - Side chain 1
				//             |__[B4] - Side chain 2
				BeforeEach(func() {
					tip, _ := genesisChain.Current()
					Expect(tip.GetNumber()).To(Equal(uint64(4)))
					Expect(genesisChain.GetParent()).To(BeNil())

					sidechain1Tip, _ := sidechain1.Current()
					Expect(sidechain1Tip.GetNumber()).To(Equal(uint64(4)))
					Expect(sidechain1.GetParent().GetID()).To(Equal(genesisChain.GetID()))

					sidechain2Tip, _ := sidechain2.Current()
					Expect(sidechain2Tip.GetNumber()).To(Equal(uint64(4)))
					Expect(sidechain2.GetParent().GetID()).To(Equal(sidechain1.GetID()))
				})

				When("account is created in block 2 of parent chain", func() {

					var account core.Account

					BeforeEach(func() {
						account = &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(2, account)
						Expect(err).To(BeNil())
					})

					It("should successfully get account", func() {
						result, err := wr.GetAccount(sidechain1, util.String(account.GetAddress()))
						Expect(err).To(BeNil())
						Expect(result).To(Equal(account))
					})
				})

				When("account is created in block 2 and updated in block 3 of parent chain", func() {

					var account core.Account

					BeforeEach(func() {
						account = &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(2, account)
						Expect(err).To(BeNil())
					})

					BeforeEach(func() {
						account2 := &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "1000",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(3, account2)
						Expect(err).To(BeNil())
					})

					It("should successfully get only account created on or before the parent's block number", func() {
						result, err := wr.GetAccount(sidechain1, util.String(account.GetAddress()))
						Expect(err).To(BeNil())
						Expect(result).To(Equal(account))
					})
				})

				When("account is created in block 3", func() {

					var account core.Account

					BeforeEach(func() {
						account = &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(3, account)
						Expect(err).To(BeNil())
					})

					It("should return ErrAccountNotFound when account does not exist on blocks created on or before the parent block number", func() {
						result, err := wr.GetAccount(sidechain1, util.String(account.GetAddress()))
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(core.ErrAccountNotFound))
						Expect(result).To(BeNil())
					})
				})

				When("account is created in block 2 of grand parent chain", func() {

					var account core.Account

					BeforeEach(func() {
						account = &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(2, account)
						Expect(err).To(BeNil())
					})

					It("should successfully get account", func() {
						result, err := wr.GetAccount(sidechain2, util.String(account.GetAddress()))
						Expect(err).To(BeNil())
						Expect(result).To(Equal(account))
					})
				})

				When("account is created in block 2 and updated in block 3 of grand parent chain", func() {

					var account core.Account

					BeforeEach(func() {
						account = &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(2, account)
						Expect(err).To(BeNil())
					})

					BeforeEach(func() {
						account2 := &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "1000",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(3, account2)
						Expect(err).To(BeNil())
					})

					It("should successfully get only account created on or before the chain's parent's parent block number", func() {
						result, err := wr.GetAccount(sidechain2, util.String(account.GetAddress()))
						Expect(err).To(BeNil())
						Expect(result).To(Equal(account))
					})
				})

				When("account is created in block 3 of grand parent", func() {

					var account core.Account

					BeforeEach(func() {
						account = &wire.Account{
							Type:    wire.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(3, account)
						Expect(err).To(BeNil())
					})

					It("should return ErrAccountNotFound when account does not exist on blocks created on or before the parent block number", func() {
						result, err := wr.GetAccount(sidechain2, util.String(account.GetAddress()))
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(core.ErrAccountNotFound))
						Expect(result).To(BeNil())
					})
				})
			})
		})
	})
}
