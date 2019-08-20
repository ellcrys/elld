package blockchain

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ellcrys/elld/util/logger"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlockValidator", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock types.Block
	var genesisChain *Chain
	var sender, receiver *crypto.Key
	var log = logger.NewLogrusNoOp()

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(log)
		err = db.Open(cfg.NetDataDir())
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
		bc.SetNodeKey(crypto.NewKeyFromIntSeed(1234))
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

	Describe(".CheckSize", func() {

		var curMaxBlockNonTxsSize, curMaxBlockTxsSize int64

		BeforeEach(func() {
			curMaxBlockNonTxsSize = params.MaxBlockNonTxsSize
			curMaxBlockTxsSize = params.MaxBlockTxsSize
		})

		AfterEach(func() {
			params.MaxBlockNonTxsSize = curMaxBlockNonTxsSize
			params.MaxBlockTxsSize = curMaxBlockTxsSize
		})

		It("should return error if block size is exceeded", func() {
			params.MaxBlockNonTxsSize = 1
			params.MaxBlockTxsSize = 1
			block := MakeBlock(bc, genesisChain, sender, receiver)
			errs := NewBlockValidator(block, nil, nil, cfg, log).CheckSize()
			Expect(errs).To(ContainElement(fmt.Errorf("block size exceeded")))
		})
	})

	Describe(".CheckFields", func() {

		Context("when block is nil", func() {
			It("should return error", func() {
				errs := NewBlockValidator(nil, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("nil block"))
			})
		})

		Context("Header: when it is invalid", func() {

			var block types.Block

			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
			})

			It("should return nil when header is not provided", func() {
				block.SetHeader(nil)
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header, error:header is required")))
			})

			It("should return error when number is 0", func() {
				block.GetHeader().SetNumber(0)
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.number, error:number must be greater or equal to 1")))
			})

			When("header number is not equal to 1", func() {
				It("should return error when parent hash is missing", func() {
					block.GetHeader().SetNumber(2)
					block.GetHeader().SetParentHash(util.Hash{})
					errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
					Expect(errs).To(ContainElement(fmt.Errorf("field:header.parentHash, error:parent hash is required")))
				})
			})

			It("should return error when header number is equal to 1", func() {
				genesisBlock.GetHeader().SetParentHash(util.StrToHash("unexpected_abc"))
				errs := NewBlockValidator(genesisBlock, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("field:header.parentHash, error:parent hash is not expected in a genesis block"))
			})

			It("should return error when creator pub key is not provided", func() {
				block.GetHeader().SetCreatorPubKey("")
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.creatorPubKey, error:creator's public key is required")))
			})

			It("should return error when creator pub key is not valid", func() {
				block.GetHeader().SetCreatorPubKey("invalid")
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.creatorPubKey, error:invalid format: version and/or checksum bytes missing")))
			})

			It("should return error when transactions root is invalid", func() {
				block.GetHeader().SetTransactionsRoot(util.Hash{1, 2, 3})
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.transactionsRoot, error:transactions root is not valid")))
			})

			It("should return error when state root is not provided", func() {
				block.GetHeader().SetStateRoot(util.Hash{})
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.stateRoot, error:state root is required")))
			})

			It("should return error when difficulty is lesser than 1", func() {
				block.GetHeader().SetDifficulty(new(big.Int).SetInt64(0))
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.difficulty, error:difficulty must be greater than zero")))
			})

			It("should return error when timestamp is not provided", func() {
				block.GetHeader().SetTimestamp(0)
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.timestamp, error:timestamp is required")))
			})

			It("should return error when timestamp is over 15 seconds in the future", func() {
				block.GetHeader().SetTimestamp(time.Now().Add(16 * time.Second).Unix())
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.timestamp, error:timestamp is too far in the future")))
			})
		})

		Context("Header: when it is valid", func() {

			var block types.Block

			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
			})

			Context("Hash", func() {

				It("should return error when hash is not provided", func() {
					block.SetHash(util.Hash{})
					errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
					Expect(errs).To(HaveLen(1))
					Expect(errs).To(ContainElement(fmt.Errorf("field:hash, error:hash is required")))
				})

				It("should return error when hash is not valid", func() {
					block.SetHash(util.Hash{1, 2, 3})
					errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
					Expect(errs).To(HaveLen(1))
					Expect(errs).To(ContainElement(fmt.Errorf("field:hash, error:hash is not correct")))
				})
			})

			Context("Signature", func() {
				It("should return error when signature is not provided", func() {
					block.SetSignature(nil)
					errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
					Expect(errs).To(HaveLen(2))
					Expect(errs).To(ContainElement(fmt.Errorf("field:sig, error:signature is required")))
				})

				It("should return error when signature is not valid", func() {
					block.SetSignature([]byte{1, 2, 3})
					block.SetHash(block.ComputeHash())
					errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
					Expect(errs).To(HaveLen(1))
					Expect(errs).To(ContainElement(fmt.Errorf("field:sig, error:signature is not valid")))
				})
			})
		})
	})

	Describe(".CheckTransactions", func() {
		Context("types.ContextBlock is set", func() {
			Context("when transaction does not exist in pool", func() {
				var block types.Block
				BeforeEach(func() {
					block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
						Transactions: []types.Transaction{
							core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.4", 1532730722),
						},
						Creator:           sender,
						Nonce:             util.EncodeNonce(1),
						Difficulty:        new(big.Int).SetInt64(131136),
						OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
					})
				})

				It("should return no error", func() {
					tp := txpool.New(1)
					validator := NewBlockValidator(block, tp, bc, cfg, log)
					validator.setContext(types.ContextBlock)
					errs := validator.CheckTransactions()
					Expect(errs).To(HaveLen(0))
				})
			})

			Context("when a sender X's current nonce is 1", func() {

				var txs []types.Transaction
				var block types.Block

				BeforeEach(func() {
					now := time.Now().Unix()
					txs = []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.4", now),
						core.NewTx(core.TxTypeBalance, 2, receiver.Addr(), sender, "1", "2.4", now),
					}
					for _, tx := range txs {
						bc.txPool.Put(tx)
					}

					block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
						Transactions:      txs,
						Creator:           sender,
						Nonce:             util.EncodeNonce(1),
						Difficulty:        new(big.Int).SetInt64(131136),
						OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
					})
				})

				Context("and X has two transactions with nonce 2 and 3", func() {
					It("should return error no error", func() {
						validator := NewBlockValidator(block, bc.txPool, bc, cfg, log)
						validator.setContext(types.ContextBlock)
						errs := validator.CheckTransactions()
						Expect(errs).To(HaveLen(0))
					})
				})
			})
		})
	})

	Describe(".checkPow", func() {
		var block types.Block

		Context("when a block's parent is unknown", func() {
			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
				block.GetHeader().SetParentHash(util.StrToHash("unknown"))
			})

			It("should return error about missing parent block", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckPoW()
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("field:parentHash, error:block not found")))
			})
		})

		Context("when block has an invalid difficulty", func() {
			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
				block.GetHeader().SetDifficulty(new(big.Int).SetInt64(131))
			})

			It("should return error when total difficulty is invalid", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckPoW()
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("field:header, error:invalid difficulty: have 131, want 100000")))
			})
		})

		// Context("when block has a invalid total difficulty", func() {

		// 	var block types.Block

		// 	BeforeEach(func() {
		// 		block = MakeBlock(bc, genesisChain, sender, receiver)
		// 		diff := bkm.CalcDifficulty(block.GetHeader(), genesisBlock.GetHeader())
		// 		block.GetHeader().SetDifficulty(diff)
		// 		block.GetHeader().SetTotalDifficulty(new(big.Int).SetInt64(10222))
		// 	})

		// 	It("should return no error", func() {
		// 		errs := NewBlockValidator(block, nil, bc, cfg, log).CheckPoW()
		// 		Expect(errs).To(HaveLen(1))
		// 		Expect(errs).To(ContainElement(fmt.Errorf("field:header, error:invalid total difficulty: have 10222, want 200000")))
		// 	})
		// })

		// Context("when block has a valid difficulty and total difficulty", func() {

		// 	var block types.Block

		// 	BeforeEach(func() {
		// 		block = MakeBlock(bc, genesisChain, sender, receiver)
		// 		diff := bkm.CalcDifficulty(block.GetHeader(), genesisBlock.GetHeader())
		// 		block.GetHeader().SetDifficulty(diff)
		// 		block.GetHeader().SetTotalDifficulty(new(big.Int).Add(diff, genesisBlock.GetHeader().GetDifficulty()))
		// 	})

		// 	It("should return invalid proof-of-work error", func() {
		// 		errs := NewBlockValidator(block, nil, bc, cfg, log).CheckPoW()
		// 		Expect(errs).To(HaveLen(1))
		// 		Expect(errs).To(ContainElement(fmt.Errorf("field:header, error:invalid proof-of-work")))
		// 	})
		// })
	})

	Describe(".checkSignature", func() {
		When("block creator's public key is not valid", func() {
			It("should return error", func() {
				genesisBlock.GetHeader().SetCreatorPubKey("invalid")
				errs := NewBlockValidator(genesisBlock, nil, bc, cfg, log).checkSignature()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("field:header.creatorPubKey, error:invalid format: version and/or checksum bytes missing"))
			})
		})

		When("signature is not valid", func() {
			It("should return error", func() {
				genesisBlock.SetSig([]byte("invalid"))
				errs := NewBlockValidator(genesisBlock, nil, bc, cfg, log).checkSignature()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("field:sig, error:signature is not valid"))
			})
		})
	})

	Describe(".checkAllocs", func() {

		When("block is a genesis block", func() {
			It("should return no error", func() {
				errs := NewBlockValidator(genesisBlock, nil, bc, cfg, log).CheckAllocs()
				Expect(errs).To(BeEmpty())
			})
		})

		When("block has no transactions", func() {
			var block types.Block
			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions:      []types.Transaction{},
					Creator:           sender,
					Nonce:             util.EncodeNonce(1),
					Difficulty:        new(big.Int).SetInt64(131136),
					OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
				})
			})

			It("should not return error", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckAllocs()
				Expect(errs).To(HaveLen(0))
			})
		})

		When("block has no fee allocation", func() {
			var block types.Block
			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions: []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
					},
					Creator:           sender,
					Nonce:             util.EncodeNonce(1),
					Difficulty:        new(big.Int).SetInt64(131136),
					OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
				})
			})

			It("should return error when block does not include the fee allocation", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckAllocs()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("field:transactions, error:block allocations and expected allocations do not match"))
			})
		})

		When("block has invalid/unexpected fee allocation", func() {
			var block types.Block
			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions: []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeAlloc, 1, sender.Addr(), sender, "1", "0", 1532730722),
					},
					Creator:           sender,
					Nonce:             util.EncodeNonce(1),
					Difficulty:        new(big.Int).SetInt64(131136),
					OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
				})
			})

			It("should return error when block does include a fee allocation with expected values", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckAllocs()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("field:transactions, error:block allocations and expected allocations do not match"))
			})
		})

		When("block has valid fee allocation", func() {
			var block types.Block
			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions: []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.36", 1532730722),
						core.NewTx(core.TxTypeAlloc, 1, sender.Addr(), sender, "7.080000000000000000", "0", 1532730722),
					},
					Creator:           sender,
					Nonce:             util.EncodeNonce(1),
					Difficulty:        new(big.Int).SetInt64(131136),
					OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
				})
			})

			It("should return no error", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckAllocs()
				Expect(errs).To(HaveLen(0))
			})
		})
	})

})
