package blockchain

import (
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var WorldReaderTest = func() bool {
	return Describe("WorldReader", func() {

		var wr *WorldReader

		BeforeEach(func() {
			wr = bc.NewWorldReader()
		})

		Describe(".GetAccount", func() {

			Describe("with no chain provided", func() {

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
						genesisB2 := makeBlockWithSingleTx(genesisChain, 1)
						_, err = bc.ProcessBlock(genesisB2)
						Expect(err).To(BeNil())

						// sidechain1 block 3
						sidechain1B3 := makeBlockWithSingleTx(genesisChain, 2)

						// genesis block 3
						genesisB3 := makeBlockWithSingleTx(genesisChain, 2)
						_, err = bc.ProcessBlock(genesisB3)
						Expect(err).To(BeNil())

						// process sidechain1B3 creates a fork chain
						reader, err := bc.ProcessBlock(sidechain1B3)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
						sidechain1 = bc.chains[reader.GetID()]

						// block 4
						genesisB4 := makeBlockWithSingleTx(genesisChain, 3)
						_, err = bc.ProcessBlock(genesisB4)
						Expect(err).To(BeNil())

						// sidechain2 block 4
						sidechain2B4 := makeBlockWithSingleTx(sidechain1, 3)

						// sidechain1 block 4
						sidechain1B4 := makeBlockWithSingleTx(sidechain1, 3)
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
						var account core.Account

						BeforeEach(func() {
							account = &objects.Account{
								Type:    objects.AccountTypeBalance,
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

						var account core.Account

						BeforeEach(func() {
							account = &objects.Account{
								Type:    objects.AccountTypeBalance,
								Balance: "100",
								Address: "addr1",
							}
							err = genesisChain.CreateAccount(2, account)
							Expect(err).To(BeNil())

							account2 := &objects.Account{
								Type:    objects.AccountTypeBalance,
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

						var account core.Account

						BeforeEach(func() {
							account = &objects.Account{
								Type:    objects.AccountTypeBalance,
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

						var account core.Account

						BeforeEach(func() {
							account = &objects.Account{
								Type:    objects.AccountTypeBalance,
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

						var account core.Account

						BeforeEach(func() {
							account = &objects.Account{
								Type:    objects.AccountTypeBalance,
								Balance: "100",
								Address: "addr1",
							}
							err = genesisChain.CreateAccount(2, account)
							Expect(err).To(BeNil())

							account2 := &objects.Account{
								Type:    objects.AccountTypeBalance,
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

						var account core.Account

						BeforeEach(func() {
							account = &objects.Account{
								Type:    objects.AccountTypeBalance,
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
}
