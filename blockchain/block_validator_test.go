package blockchain

import (
	"fmt"
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

var _ = Describe("BlockValidator", func() {

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

		db = elldb.NewDB(cfg.ConfigDir())
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
		err = os.RemoveAll(cfg.ConfigDir())
		Expect(err).To(BeNil())
	})

	Describe(".checkFields", func() {

		Context("when block is nil", func() {
			It("should return error", func() {
				errs := NewBlockValidator(nil, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("nil block"))
			})
		})

		Context("Header: when it is invalid", func() {

			var block *objects.Block

			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver).(*objects.Block)
			})

			It("should return nil when header is not provided", func() {
				block.Header = nil
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header, error:header is required")))
			})

			It("should return error when number is 0", func() {
				block.Header.Number = 0
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.number, error:number must be greater or equal to 1")))
			})

			When("header number is not equal to 1", func() {
				It("should return error when parent hash is missing", func() {
					block.Header.Number = 2
					block.Header.ParentHash = util.Hash{}
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
				block.Header.CreatorPubKey = ""
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.creatorPubKey, error:creator's public key is required")))
			})

			It("should return error when creator pub key is not valid", func() {
				block.Header.CreatorPubKey = "invalid"
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.creatorPubKey, error:invalid format: version and/or checksum bytes missing")))
			})

			It("should return error when transactions root is not provided", func() {
				block.Header.TransactionsRoot = util.Hash{}
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.transactionsRoot, error:transaction root is required")))
			})

			It("should return error when transactions root is invalid", func() {
				block.Header.TransactionsRoot = util.Hash{1, 2, 3}
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.transactionsRoot, error:transactions root is not valid")))
			})

			It("should return error when state root is not provided", func() {
				block.Header.StateRoot = util.Hash{}
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.stateRoot, error:state root is required")))
			})

			It("should return error when difficulty is lesser than 1", func() {
				block.Header.Difficulty = new(big.Int).SetInt64(0)
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.difficulty, error:difficulty must be greater than zero")))
			})

			It("should return error when timestamp is not provided", func() {
				block.Header.Timestamp = 0
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.timestamp, error:timestamp is required")))
			})

			It("should return error when timestamp is over 15 seconds in the future", func() {
				block.Header.Timestamp = time.Now().Add(16 * time.Second).Unix()
				errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
				Expect(errs).To(ContainElement(fmt.Errorf("field:header.timestamp, error:timestamp is too far in the future")))
			})
		})

		Context("Header: when it is valid", func() {

			var block core.Block

			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver)
			})

			Context("Transactions", func() {
				When("only allocation transactions are in the block", func() {
					BeforeEach(func() {
						block = MakeBlockWithOnlyAllocTx(bc, genesisChain, sender, receiver)
					})

					It("should return error about the lack of transactions", func() {
						errs := NewBlockValidator(block, nil, nil, cfg, log).CheckFields()
						Expect(errs).To(HaveLen(1))
						Expect(errs).To(ContainElement(fmt.Errorf("field:transactions, error:at least one transaction is required")))
					})
				})
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

	Describe(".checkPow", func() {
		var block core.Block

		Context("when a block's parent is unknown", func() {
			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver).(*objects.Block)
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
				block = MakeBlock(bc, genesisChain, sender, receiver).(*objects.Block)
				block.GetHeader().SetDifficulty(new(big.Int).SetInt64(131))
			})

			It("should return error when total difficulty is invalid", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckPoW()
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("field:parentHash, error:invalid difficulty: have 131, want 9995117188")))
			})
		})

		Context("when block that has an invalid total difficulty", func() {

			BeforeEach(func() {
				block = MakeBlock(bc, genesisChain, sender, receiver).(*objects.Block)
				block.GetHeader().SetDifficulty(new(big.Int).SetInt64(9995117188))
			})

			It("should return error when total difficulty is invalid", func() {
				errs := NewBlockValidator(block, nil, bc, cfg, log).CheckPoW()
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("field:parentHash, error:invalid total difficulty: have 0, want 19995117188")))
			})
		})
	})

	Describe(".checkSignature", func() {
		When("block creator's public key is not valid", func() {
			It("should return error", func() {
				genesisBlock.(*objects.Block).Header.CreatorPubKey = "invalid"
				errs := NewBlockValidator(genesisBlock, nil, bc, cfg, log).checkSignature()
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("field:header.creatorPubKey, error:invalid format: version and/or checksum bytes missing"))
			})
		})

		When("signature is not valid", func() {
			It("should return error", func() {
				genesisBlock.(*objects.Block).Sig = []byte("invalid")
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

		When("block has not fee allocation", func() {
			var block core.Block
			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
					},
					Creator:           sender,
					Nonce:             core.EncodeNonce(1),
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
			var block core.Block
			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "1", "0", 1532730722),
					},
					Creator:           sender,
					Nonce:             core.EncodeNonce(1),
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
			var block core.Block
			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "7.080000000000000000", "0", 1532730722),
					},
					Creator:           sender,
					Nonce:             core.EncodeNonce(1),
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
