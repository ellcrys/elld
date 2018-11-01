package blockchain

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

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
	"github.com/jinzhu/copier"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestProcess(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Process", func() {

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

		g.Describe(".addOp", func() {

			var curOps = []common.Transition{
				&common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr1"},
					Account: &core.Account{
						Balance: "10",
					},
				},
			}

			g.It("should add an additional op successfully", func() {
				op := &common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr2"},
					Account: &core.Account{
						Balance: "10",
					},
				}
				newOps := addOp(curOps, op)
				Expect(newOps).To(HaveLen(2))
			})

			g.It("should replace any equal op found", func() {
				op := &common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr1"},
					Account: &core.Account{
						Balance: "30",
					},
				}
				newOps := addOp(curOps, op)
				Expect(newOps).To(HaveLen(1))
				Expect(newOps[0]).To(Equal(op))
				Expect(newOps[0]).ToNot(Equal(curOps[0]))
			})
		})

		g.Describe(".processTransactions (only TxTypeBalance transactions)", func() {

			var account types.Account

			g.BeforeEach(func() {
				account = &core.Account{Type: core.AccountTypeBalance, Address: util.String(sender.Addr()), Balance: "10"}
				err = bc.CreateAccount(1, genesisChain, account)
				Expect(err).To(BeNil())
			})

			g.It("should return error if sender does not exist on the best chain", func() {
				var txs = []types.Transaction{
					&core.Transaction{
						Type:         1,
						Nonce:        1,
						To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
						From:         "unknown",
						SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value:        "1",
						Timestamp:    1532730724,
						Fee:          "0.1",
						Sig:          []uint8{},
						Hash:         util.Hash{},
					},
				}

				_, err := bc.ProcessTransactions(txs, genesisChain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: failed to get sender's account: account not found"))
			})

			g.It("should return error if sender account has insufficient value", func() {
				var txs = []types.Transaction{
					&core.Transaction{
						Type:         1,
						Nonce:        1,
						To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
						From:         sender.Addr(),
						SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value:        "10000000", Timestamp: 1532730724,
						Fee:  "0.1",
						Sig:  []uint8{},
						Hash: util.Hash{},
					},
				}

				_, err := bc.ProcessTransactions(txs, genesisChain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: insufficient sender account balance"))
			})

			g.It("should panic if sender value is could not be converted to decimal", func() {
				var txs = []types.Transaction{
					&core.Transaction{
						Type:         1,
						Nonce:        1,
						To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
						From:         sender.Addr(),
						SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value:        "100_333", Timestamp: 1532730724,
						Fee: "0.1", Sig: []uint8{},
						Hash: util.Hash{},
					},
				}
				Expect(func() {
					bc.ProcessTransactions(txs, genesisChain)
				}).To(Panic())
			})

			g.Context("recipient does not have an account", func() {

				var ops []common.Transition
				var txs []types.Transaction
				var err error

				g.BeforeEach(func() {
					txs = []types.Transaction{
						&core.Transaction{
							Type: 1, Nonce: 1,
							To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
							From:         sender.Addr(),
							SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value:        "1",
							Timestamp:    1532730724,
							Fee:          "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}
					ops, err = bc.ProcessTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
				})

				g.It("should return 3 operations", func() {
					Expect(ops).To(HaveLen(3))
				})

				g.It("first op should be an OpCreateAccount for the recipient", func() {
					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
					Expect(ops[0].Address()).To(Equal(txs[0].GetTo()))
				})

				g.It("second op should be an OpNewAccountBalance for the sender with account balance  = 8.900000000000000000", func() {
					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].GetFrom()))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("8.900000000000000000")))
				})

				g.It("third op should be an OpNewAccountBalance for the recipient with account balance  = 1.000000000000000000", func() {
					Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[2].Address()).To(Equal(txs[0].GetTo()))
					Expect(ops[2].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("1.000000000000000000")))
				})
			})

			g.Context("recipient has an account", func() {
				var receiver = crypto.NewKeyFromIntSeed(3)
				var ops []common.Transition
				var txs []types.Transaction
				var err error

				g.BeforeEach(func() {
					account = &core.Account{Type: core.AccountTypeBalance, Address: util.String(receiver.Addr()), Balance: "0"}
					err = bc.CreateAccount(1, genesisChain, account)
					Expect(err).To(BeNil())

					txs = []types.Transaction{
						&core.Transaction{
							Type: 1, Nonce: 1,
							To:           util.String(receiver.Addr()),
							From:         sender.Addr(),
							SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value:        "1",
							Timestamp:    1532730724,
							Fee:          "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}
					ops, err = bc.ProcessTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
				})

				g.It("should return 2 operations", func() {
					Expect(ops).To(HaveLen(2))
				})

				g.It("first op should be an OpNewAccountBalance for the sender with account balance = 8.900000000000000000", func() {
					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].Address()).To(Equal(txs[0].GetFrom()))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("8.900000000000000000")))
				})

				g.It("second op should be an OpNewAccountBalance for the recipient with account balance  = 1.000000000000000000", func() {
					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].GetTo()))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("1.000000000000000000")))
				})
			})

			g.Context("sender has two transactions sending to a recipient", func() {

				g.Context("recipient has an account", func() {
					var receiver = crypto.NewKeyFromIntSeed(3)
					var ops []common.Transition
					var txs []types.Transaction
					var err error

					g.BeforeEach(func() {
						account = &core.Account{Type: core.AccountTypeBalance, Address: util.String(receiver.Addr()), Balance: "0"}
						err = bc.CreateAccount(1, genesisChain, account)
						Expect(err).To(BeNil())

						txs = []types.Transaction{
							&core.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: sender.Addr(), SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
							&core.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: sender.Addr(), SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}
						ops, err = bc.ProcessTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
					})

					g.It("should return 2 operations", func() {
						Expect(ops).To(HaveLen(2))
					})

					g.It("first op should be an OpNewAccountBalance for the sender with account balance = 7.800000000000000000", func() {
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].Address()).To(Equal(txs[0].GetFrom()))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("7.800000000000000000")))
					})

					g.It("second op should be an OpNewAccountBalance for the recipient with account balance  = 2.000000000000000000", func() {
						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].GetTo()))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("2.000000000000000000")))
					})
				})

				g.Context("recipient does not have an account", func() {
					var ops []common.Transition
					var txs []types.Transaction
					var err error

					g.BeforeEach(func() {
						txs = []types.Transaction{
							&core.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: sender.Addr(), SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
							&core.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: sender.Addr(), SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}
						ops, err = bc.ProcessTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
					})

					g.It("should return 3 operations", func() {
						Expect(ops).To(HaveLen(3))
					})

					g.It("first op should be an OpCreateAccount for the recipient", func() {
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
						Expect(ops[0].Address()).To(Equal(txs[0].GetTo()))
					})

					g.It("second op should be an OpNewAccountBalance for the sender with account balance  = 7.800000000000000000", func() {
						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].GetFrom()))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("7.800000000000000000")))
					})

					g.It("third op should be an OpNewAccountBalance for the recipient with account balance  = 2.000000000000000000", func() {
						Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[2].Address()).To(Equal(txs[0].GetTo()))
						Expect(ops[2].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("2.000000000000000000")))
					})
				})
			})
		})

		g.Describe(".processTransactions (only TxTypeAllocCoin transactions)", func() {
			g.When("recipient account does not exist", func() {
				g.It("should successfully return one state object = OpNewAccountBalance", func() {
					var txs = []types.Transaction{
						&core.Transaction{Type: core.TxTypeAlloc, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: sender.Addr(), SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value: "10", Timestamp: 1532730724,
							Fee: "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}

					ops, err := bc.ProcessTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(1))
					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("10.000000000000000000")))
				})
			})

			g.When("recipient account already exists with account balance = 100", func() {
				g.BeforeEach(func() {
					Expect(bc.CreateAccount(1, genesisChain, &core.Account{
						Type:    core.AccountTypeBalance,
						Address: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
						Balance: "100",
					})).To(BeNil())
				})

				g.It("should successfully return one state object = OpNewAccountBalance and Balance = 110.000000000000000000", func() {
					var txs = []types.Transaction{
						&core.Transaction{Type: core.TxTypeAlloc, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: sender.Addr(), SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value: "10", Timestamp: 1532730724,
							Fee: "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}
					ops, err := bc.ProcessTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(1))
					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("110.000000000000000000")))
				})
			})
		})

		g.Describe(".ComputeTxsRoot", func() {
			g.It("should return expected root", func() {
				txs := []types.Transaction{
					&core.Transaction{Sig: []byte("b"), Hash: util.StrToHash("hash_1")},
					&core.Transaction{Sig: []byte("a"), Hash: util.StrToHash("hash_2")},
				}
				root := common.ComputeTxsRoot(txs)
				expected := util.Hash{
					0x37, 0xef, 0x76, 0x42, 0x69, 0x87, 0x67, 0xba, 0x8b, 0xfe, 0xf7, 0x5d, 0x66, 0x91, 0xda, 0x12,
					0x20, 0xb1, 0x2d, 0x11, 0x81, 0xeb, 0x85, 0x9e, 0x5a, 0x0a, 0xb3, 0xbb, 0x11, 0x7d, 0x75, 0xdb,
				}
				Expect(root).To(Equal(expected))
				Expect(root).To(HaveLen(32))
			})
		})

		g.Describe(".execBlock", func() {
			var block types.Block

			g.BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
			})

			g.Context("when block has a transaction that failed validation", func() {
				g.It("should return error", func() {
					newSender := crypto.NewKeyFromIntSeed(3)
					block.GetTransactions()[0].SetFrom(util.String(newSender.Addr()))
					_, _, err := bc.execBlock(genesisChain, block)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})

			g.Context("block does not include fee allocation transaction", func() {

				g.BeforeEach(func() {
					block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
						Transactions: []types.Transaction{
							core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730724),
						},
						Creator:    sender,
						Nonce:      util.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
				})

				g.Specify("balance must be less than initial balance because no fee is paid back. The fee is lost", func() {
					_, stateObjs, err := bc.execBlock(genesisChain, block)
					Expect(err).To(BeNil())
					Expect(stateObjs).To(HaveLen(1))

					var m map[string]interface{}
					util.BytesToObject(stateObjs[0].Value, &m)
					Expect(m["balance"]).To(Equal("97.500000000000000000"))
				})
			})

			g.When("block includes a fee allocation transaction", func() {

				g.BeforeEach(func() {
					block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
						Transactions: []types.Transaction{
							core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730724),
						},
						Creator:     sender,
						Nonce:       util.EncodeNonce(1),
						Difficulty:  new(big.Int).SetInt64(131072),
						AddFeeAlloc: true,
					})
				})

				g.Specify("balance is equal to initial balance; fee is paid back; fee is not lost", func() {
					_, stateObjs, err := bc.execBlock(genesisChain, block)
					Expect(err).To(BeNil())
					Expect(stateObjs).To(HaveLen(1))

					var m map[string]interface{}
					util.BytesToObject(stateObjs[0].Value, &m)
					Expect(m["balance"]).To(Equal("100.000000000000000000"))
				})
			})
		})

		g.Describe(".ProcessBlock", func() {

			var block types.Block

			g.BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions: []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730722),
						core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730722),
					},
					Creator:    sender,
					Nonce:      util.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			g.It("should reject the block if it has been added to the rejected cache", func() {
				bc.rejectedBlocks.Add(block.GetHashAsHex(), struct{}{})
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBlockRejected))
			})

			g.It("should return error if block already exists in one of the known chains", func() {
				err = genesisChain.append(block)
				Expect(err).To(BeNil())
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("block already exists")))
			})

			g.It("should return error if block has been added to the orphaned cache", func() {
				bc.orphanBlocks.Add(block.GetHashAsHex(), block)
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("orphan block")))
			})

			g.When("a block's parent does not exist in any chain", func() {

				g.BeforeEach(func() {
					block.GetHeader().SetParentHash(util.StrToHash("unknown"))
					block.SetHash(block.ComputeHash())
					blockSig, _ := core.BlockSign(block, sender.PrivKey().Base58())
					block.SetSignature(blockSig)
				})

				g.It("should return nil and be added to the orphan block cache", func() {
					_, err = bc.ProcessBlock(block)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Has(block.GetHashAsHex())).To(BeTrue())
				})
			})

			g.When("a block's parent exists in a chain", func() {

				g.When("block's timestamp is lesser than its parent's timestamp", func() {

					var block2 types.Block

					g.BeforeEach(func() {
						block2 = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730724),
							},
							Creator:           sender,
							Nonce:             util.EncodeNonce(1),
							Difficulty:        new(big.Int).SetInt64(131072),
							OverrideTimestamp: genesisBlock.GetHeader().GetTimestamp() - 1,
						})
					})

					g.It("should return error", func() {
						_, err = bc.ProcessBlock(block2)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("block timestamp must be greater than its parent's"))
						Expect(bc.isRejected(block2)).To(BeTrue())
					})
				})

				g.When("block number is less than the chain tip block number", func() {

					var staleBlock2 core.Block

					g.BeforeEach(func() {
						block2 := MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730724),
							},
							Creator:           sender,
							Nonce:             util.EncodeNonce(1),
							Difficulty:        new(big.Int).SetInt64(131072),
							OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
						})
						err = genesisChain.append(block2)
						Expect(err).To(BeNil())

						block3 := MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730725),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
							},
							Creator:           sender,
							Nonce:             util.EncodeNonce(2),
							Difficulty:        new(big.Int).SetInt64(131072),
							OverrideTimestamp: time.Now().Add(3 * time.Second).Unix(),
						})
						err = genesisChain.append(block3)
						Expect(err).To(BeNil())

						copier.Copy(&staleBlock2, block2)
						staleBlock2.GetHeader().SetNumber(2)
						staleBlock2.GetHeader().SetNonce(util.EncodeNonce(3))
						staleBlock2.Hash = staleBlock2.ComputeHash()
						staleBlock2.Sig, _ = core.BlockSign(&staleBlock2, sender.PrivKey().Base58())
					})

					g.It("should return nil and create a new chain", func() {
						Expect(bc.chains).To(HaveLen(1))
						_, err = bc.ProcessBlock(&staleBlock2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
					})
				})

				g.When("a block has same number as the chainTip", func() {
					var block2, block2_2 types.Block
					g.BeforeEach(func() {
						block2 = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
							},
							Creator:    sender,
							Nonce:      util.EncodeNonce(1),
							Difficulty: new(big.Int).SetInt64(131072),
						})

						block2_2 = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730725),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
							},
							Creator:    sender,
							Nonce:      util.EncodeNonce(4),
							Difficulty: new(big.Int).SetInt64(131072),
						})

						err = genesisChain.append(block2)
						Expect(err).To(BeNil())
					})

					g.It("should create a new chain tree; return nil; new tree's parent should be expected; tree must include the new block", func() {
						ch, err := bc.ProcessBlock(block2_2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
						Expect(bc.chains[ch.GetID()].parentBlock.GetHash()).To(Equal(genesisBlock.GetHash()))
						hasBlock, err := bc.chains[ch.GetID()].hasBlock(block2_2.GetHash())
						Expect(err).To(BeNil())
						Expect(hasBlock).To(BeTrue())
					})
				})

				g.When("a block number is the same as the chain height", func() {

					var block2 types.Block

					g.BeforeEach(func() {

						block2 = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
							},
							Creator:    sender,
							Nonce:      util.EncodeNonce(1),
							Difficulty: new(big.Int).SetInt64(131072),
						})

						block2.GetHeader().SetNumber(2)
						block2.SetHash(block2.ComputeHash())
						block3Sig, _ := core.BlockSign(block2, sender.PrivKey().Base58())
						block2.SetSignature(block3Sig)

						err = genesisChain.append(block)
						Expect(err).To(BeNil())
					})

					g.It("should return nil and create a new chain", func() {
						Expect(bc.chains).To(HaveLen(1))
						_, err := bc.ProcessBlock(block2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
					})
				})
			})

			g.Context("state root comparison", func() {

				var blockInvalidStateRoot, okStateRoot types.Block

				g.BeforeEach(func() {
					blockInvalidStateRoot = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
						Transactions: []types.Transaction{
							core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
							core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
						},
						Creator:    sender,
						Nonce:      util.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})

					blockInvalidStateRoot.GetHeader().SetStateRoot(util.StrToHash("incorrect"))
					blockInvalidStateRoot.SetHash(blockInvalidStateRoot.ComputeHash())
					blockInvalidStateRootSig, _ := core.BlockSign(blockInvalidStateRoot, sender.PrivKey().Base58())
					blockInvalidStateRoot.SetSignature(blockInvalidStateRootSig)

					okStateRoot = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
						Transactions: []types.Transaction{
							core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730724),
							core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
						},
						Creator:    sender,
						Nonce:      util.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
				})

				g.It("should return error when block state root does not match", func() {
					_, err = bc.ProcessBlock(blockInvalidStateRoot)
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(core.ErrBlockStateRootInvalid))
				})

				g.It("should successfully accept state root of block", func() {
					_, stateObjs, err := bc.execBlock(genesisChain, okStateRoot)
					Expect(err).To(BeNil())

					_, err = bc.ProcessBlock(okStateRoot)
					Expect(err).To(BeNil())

					g.Describe("chain should contain newly added block", func() {
						mBlock, err := genesisChain.getBlockByHash(okStateRoot.GetHash())
						Expect(err).To(BeNil())
						Expect(mBlock.GetBytes()).To(Equal(okStateRoot.GetBytes()))
					})

					g.Describe("all state objects must be persisted", func() {
						for _, so := range stateObjs {
							var result = db.GetByPrefix(so.Key)
							Expect(result).To(HaveLen(1))
						}
					})

					g.Describe("all transactions must be persisted", func() {
						for _, tx := range okStateRoot.GetTransactions() {
							txKey := common.MakeQueryKeyTransaction(genesisChain.GetID().Bytes(),
								tx.GetHash().Hex())
							var result = db.GetByPrefix(txKey)
							Expect(result).To(HaveLen(1))
						}
					})
				})
			})

			g.When("ContextBlock only is set", func() {

				g.BeforeEach(func() {
					Expect(bc.getBlockValidator(nil).has(types.ContextBlock)).To(BeTrue())
				})

				g.When("block includes a transaction that does not exist in the pool", func() {
					var block types.Block

					g.BeforeEach(func() {
						block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730723),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730723),
							},
							Creator:              sender,
							Nonce:                util.EncodeNonce(1),
							Difficulty:           new(big.Int).SetInt64(131072),
							NoPoolAdditionInTest: true,
						})
					})

					g.It("should return error", func() {
						_, err = bc.ProcessBlock(block)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("tx:0, error:transaction does not " +
							"exist in the transactions pool"))
					})
				})
			})

			g.When("ContextBlockSync is set", func() {

				g.When("block includes a transaction that does not exist in the pool", func() {
					var block types.Block

					g.BeforeEach(func() {
						block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
							Transactions: []types.Transaction{
								core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730723),
								core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730723),
							},
							Creator:              sender,
							Nonce:                util.EncodeNonce(1),
							Difficulty:           new(big.Int).SetInt64(131072),
							NoPoolAdditionInTest: true,
						})
						block.SetValidationContexts(types.ContextBlockSync)
					})

					g.It("should be successful with no error", func() {
						_, err = bc.ProcessBlock(block)
						Expect(err).To(BeNil())
					})
				})
			})
		})

		g.Describe(".ProcessBlock: Test internal call of .processOrphanBlocks", func() {

			var parent1, orphanParent, orphan types.Block
			var bc2 *Blockchain
			var db elldb.DB

			// Create a blockchain (bc2) with a main chain of 4 blocks
			// e.g: [1]-[2]-[3]-[4]
			//
			// To test .processOrphanBlocks, we will attempt to add the
			// blocks in the default blockchain (bc) in such a way
			// that some blocks are considered orphans.
			g.BeforeEach(func() {

				db = elldb.NewDB(cfg.DataDir())
				err = db.Open(util.RandString(5))

				bc2 = New(bc.txPool, cfg, log)
				bc2.SetDB(db)
				bc2.SetGenesisBlock(genesisBlock)
				err = bc2.Up()
				Expect(err).To(BeNil())

				parent1 = MakeBlock(bc2, bc2.bestChain, sender, receiver)
				_, err = bc2.ProcessBlock(parent1)
				Expect(err).To(BeNil())

				orphanParent = MakeBlockWithSingleTx(bc2, bc2.bestChain, sender, receiver, 2)
				_, err = bc2.ProcessBlock(orphanParent)
				Expect(err).To(BeNil())

				orphan = MakeBlockWithSingleTx(bc2, bc2.bestChain, sender, receiver, 3)
				_, err = bc2.ProcessBlock(orphan)
				Expect(err).To(BeNil())

				bc2CurBlock, _ := bc2.bestChain.Current()
				Expect(bc2CurBlock.GetNumber()).To(Equal(uint64(4)))
			})

			g.AfterEach(func() {
				db.Close()
			})

			g.Context("with one orphan block", func() {

				g.BeforeEach(func() {
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(1))
				})

				g.Context("when orphan parent is processed successfully", func() {
					g.BeforeEach(func() {
						_, err = bc.ProcessBlock(parent1)
						Expect(err).To(BeNil())
					})

					g.It("should recursively process all orphans when their parent exists in a chain", func() {
						Expect(genesisChain.hasBlock(parent1.GetHash())).To(BeTrue())
						Expect(genesisChain.hasBlock(orphanParent.GetHash())).To(BeTrue())
						Expect(bc.orphanBlocks.Len()).To(Equal(0))
					})

					g.Specify("chain must contain the previously orphaned block as the tip", func() {
						has, err := genesisChain.hasBlock(orphanParent.GetHash())
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := genesisChain.Current()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphanParent.GetHeader().ComputeHash()))
					})
				})
			})

			g.Context("with more than one orphan block", func() {

				g.BeforeEach(func() {
					_, err = bc.ProcessBlock(orphan)
					Expect(err).To(BeNil())
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(2))
					_, err = bc.ProcessBlock(parent1)
					Expect(err).To(BeNil())
				})

				g.It("should successfully add block and recursively process add all orphans when their parent exists in a chain", func() {
					Expect(genesisChain.hasBlock(parent1.GetHash())).To(BeTrue())
					Expect(genesisChain.hasBlock(orphanParent.GetHash())).To(BeTrue())
					Expect(genesisChain.hasBlock(orphan.GetHash())).To(BeTrue())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))
				})

				g.Specify("chain must contain the previously orphaned block as the tip", func() {
					has, err := genesisChain.hasBlock(orphan.GetHash())
					Expect(err).To(BeNil())
					Expect(has).To(BeTrue())
					tipHeader, err := genesisChain.Current()
					Expect(err).To(BeNil())
					Expect(tipHeader.ComputeHash()).To(Equal(orphan.GetHeader().ComputeHash()))
				})
			})
		})
	})
}
