package blockchain

import (
	"os"

	. "github.com/onsi/ginkgo"

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

var _ = Describe("WorldReader", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock types.Block
	var genesisChain *Chain
	var sender, receiver *crypto.Key
	var wr *WorldReader

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
		wr = bc.NewWorldReader()
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

	Describe(".GetAccount", func() {

		Describe("with no chain provided", func() {

			It("should return error if best chain is not set", func() {
				bc.bestChain = nil
				_, err := wr.GetAccount(nil, util.String(sender.Addr()))
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBestChainUnknown))
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

		Describe("with a chain", func() {

			Context("chain has no parent", func() {
				It("should successfully get account", func() {
					result, err := wr.GetAccount(genesisChain, util.String(sender.Addr()))
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())
				})
			})

			Context("chain with parents", func() {
				var sidechain1, sidechain2 *Chain

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
					Expect(bc.chains).To(HaveLen(1))

					// genesis block 2
					genesisB2 := MakeBlockWithSingleTx(bc, genesisChain, sender, receiver, 1)
					_, err = bc.ProcessBlock(genesisB2)
					Expect(err).To(BeNil())

					// sidechain1 block 3
					sidechain1B3 := MakeBlockWithSingleTx(bc, genesisChain, sender, receiver, 2)

					// genesis block 3
					genesisB3 := MakeBlockWithSingleTx(bc, genesisChain, sender, receiver, 2)
					_, err = bc.ProcessBlock(genesisB3)
					Expect(err).To(BeNil())

					// process sidechain1B3 creates a fork chain
					reader, err := bc.ProcessBlock(sidechain1B3)
					Expect(err).To(BeNil())
					Expect(bc.chains).To(HaveLen(2))
					sidechain1 = bc.chains[reader.GetID()]

					// block 4
					genesisB4 := MakeBlockWithSingleTx(bc, genesisChain, sender, receiver, 3)
					_, err = bc.ProcessBlock(genesisB4)
					Expect(err).To(BeNil())

					// sidechain2 block 4
					sidechain2B4 := MakeBlockWithSingleTx(bc, sidechain1, sender, receiver, 3)

					// sidechain1 block 4
					sidechain1B4 := MakeBlockWithSingleTx(bc, sidechain1, sender, receiver, 3)
					_, err = bc.ProcessBlock(sidechain1B4)
					Expect(err).To(BeNil())

					// process sidechain2B4 create a fork chain
					reader, err = bc.ProcessBlock(sidechain2B4)
					Expect(err).To(BeNil())
					Expect(bc.chains).To(HaveLen(3))
					sidechain2 = bc.chains[reader.GetID()]
				})

				// Ensure the target blockchain shape
				// is as expected.
				// Target blockchain shape:
				//
				// [B1]-[B2]-[B3]-[B4] - Genesis chain shape
				//       |___[B3]-[B4] - Side chain 1
				//             |__[B4] - Side chain 2
				BeforeEach(func() {
					tip, _ := sidechain1.Current()
					Expect(tip.GetNumber()).To(Equal(uint64(4)))
					Expect(genesisChain.GetParent()).To(BeNil())

					sidechain1Tip, _ := sidechain1.Current()
					Expect(sidechain1Tip.GetNumber()).To(Equal(uint64(4)))
					parent := sidechain1.GetParent()
					Expect(parent).ToNot(BeNil())
					Expect(parent.GetID()).To(Equal(genesisChain.GetID()))

					sidechain2Tip, _ := sidechain2.Current()
					Expect(sidechain2Tip.GetNumber()).To(Equal(uint64(4)))
					parent = sidechain2.GetParent()
					Expect(parent).ToNot(BeNil())
					Expect(parent.GetID()).To(Equal(sidechain1.GetID()))
				})

				Describe("account is created in block 2 of parent chain", func() {
					var account types.Account

					BeforeEach(func() {
						account = &core.Account{
							Type:    core.AccountTypeBalance,
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

				Describe("account is created in block 2 and updated in block 3 of parent chain", func() {

					var account types.Account

					BeforeEach(func() {
						account = &core.Account{
							Type:    core.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(2, account)
						Expect(err).To(BeNil())

						account2 := &core.Account{
							Type:    core.AccountTypeBalance,
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

				Describe("account is created in block 3", func() {

					var account types.Account

					BeforeEach(func() {
						account = &core.Account{
							Type:    core.AccountTypeBalance,
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

				Describe("account is created in block 2 of grand parent chain", func() {

					var account types.Account

					BeforeEach(func() {
						account = &core.Account{
							Type:    core.AccountTypeBalance,
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

				Describe("account is created in block 2 and updated in block 3 of grand parent chain", func() {

					var account types.Account

					BeforeEach(func() {
						account = &core.Account{
							Type:    core.AccountTypeBalance,
							Balance: "100",
							Address: "addr1",
						}
						err = genesisChain.CreateAccount(2, account)
						Expect(err).To(BeNil())

						account2 := &core.Account{
							Type:    core.AccountTypeBalance,
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

				Describe("account is created in block 3 of grand parent", func() {

					var account types.Account

					BeforeEach(func() {
						account = &core.Account{
							Type:    core.AccountTypeBalance,
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
})
