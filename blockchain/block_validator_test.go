package blockchain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ellcrys/elld/miner/blakimoto"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BlockValidatorTest = func() bool {
	return Describe("BlockValidator", func() {

		BeforeEach(func() {
			bc.bestChain = genesisChain
		})

		Describe(".check", func() {
			It("should check for validation errors", func() {
				var cases = map[core.Block]interface{}{
					nil:           fmt.Errorf("nil block"),
					&wire.Block{}: fmt.Errorf("field:header, error:header is required"),
					&wire.Block{}: fmt.Errorf("field:hash, error:hash is required"),
					&wire.Block{Hash: util.StrToHash("invalid"), Header: &wire.Header{}}: fmt.Errorf("field:hash, error:hash is not correct"),
					&wire.Block{}:                                                                                        fmt.Errorf("field:sig, error:signature is required"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.parentHash, error:parent hash is required"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.number, error:number must be greater or equal to 1"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.number, error:number must be greater or equal to 1"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.creatorPubKey, error:creator's public key is required"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.transactionsRoot, error:transaction root is required"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.stateRoot, error:state root is required"),
					&wire.Block{Header: &wire.Header{ParentHash: util.StrToHash("abc")}}:                                 fmt.Errorf("field:header.difficulty, error:difficulty must be non-zero and non-negative"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.timestamp, error:timestamp must not be greater or equal to 1"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:transactions, error:at least one transaction is required"),
					&wire.Block{Header: &wire.Header{}, Transactions: []*wire.Transaction{&wire.Transaction{Type: 109}}}: fmt.Errorf("tx:0, field:type, error:unsupported transaction type"),
				}
				for b, err := range cases {
					validator := NewBlockValidator(b, nil, nil, false, cfg, log)
					errs := validator.check()
					Expect(errs).To(ContainElement(err))
				}
			})
		})

		Describe(".checkSignature", func() {
			It("should check for validation errors", func() {
				key := crypto.NewKeyFromIntSeed(1)
				var cases = map[core.Block]interface{}{
					&wire.Block{Header: &wire.Header{}}:                                                  fmt.Errorf("field:header.creatorPubKey, error:empty pub key"),
					&wire.Block{Header: &wire.Header{CreatorPubKey: "invalid"}}:                          fmt.Errorf("field:header.creatorPubKey, error:invalid format: version and/or checksum bytes missing"),
					&wire.Block{Header: &wire.Header{CreatorPubKey: util.String(key.PubKey().Base58())}}: fmt.Errorf("field:sig, error:signature is not valid"),
				}
				for b, err := range cases {
					validator := NewBlockValidator(b, nil, nil, false, cfg, log)
					errs := validator.checkSignature()
					Expect(errs).To(ContainElement(err))
				}
			})
		})

		Describe(".Validate", func() {

			var block core.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730722),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			BeforeEach(func() {
				_, err = bc.ProcessBlock(block)
				Expect(err).To(BeNil())
			})

			It("should return if block and a transaction in the block exist", func() {
				validator := NewBlockValidator(block, bc.txPool, bc, true, cfg, log)
				errs := validator.Validate()
				Expect(errs).To(ContainElement(fmt.Errorf("error:block found in chain")))
				Expect(errs).To(ContainElement(fmt.Errorf("tx:0, error:transaction already exist in main chain")))
			})
		})

		Describe(".checkPow", func() {
			var block core.Block

			Context("with block that has an invalid difficulty", func() {
				BeforeEach(func() {
					block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:           sender,
						Nonce:             core.EncodeNonce(1),
						Difficulty:        new(big.Int).SetInt64(131072),
						OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
					})
				})

				It("should return error if difficulty is not valid", func() {
					validator := NewBlockValidator(block, nil, bc, true, cfg, log)
					errs := validator.checkPoW()
					Expect(errs).To(HaveLen(1))
					Expect(errs).To(ContainElement(fmt.Errorf("field:parentHash, error:invalid difficulty: have 131072, want 131136")))
				})
			})

			Context("with block that has a valid difficulty", func() {
				BeforeEach(func() {
					block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:           sender,
						Nonce:             core.EncodeNonce(1),
						Difficulty:        new(big.Int).SetInt64(1),
						OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
					})
					block.GetHeader().SetDifficulty(blakimoto.CalcDifficulty(uint64(block.GetHeader().GetTimestamp()), genesisBlock.GetHeader()))
					block.SetHash(block.ComputeHash())
					blockSig, _ := wire.BlockSign(block, sender.PrivKey().Base58())
					block.SetSignature(blockSig)
				})

				It("should return error if total difficulty is invalid", func() {
					validator := NewBlockValidator(block, nil, bc, true, cfg, log)
					errs := validator.checkPoW()
					Expect(errs).To(HaveLen(1))
					Expect(errs[0].Error()).To(ContainSubstring("field:parentHash, error:invalid total difficulty"))
				})
			})

			Context("with valid difficulty and total difficulty", func() {
				BeforeEach(func() {
					block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:           sender,
						Nonce:             core.EncodeNonce(1),
						Difficulty:        new(big.Int).SetInt64(1),
						OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
					})
					diff := blakimoto.CalcDifficulty(uint64(block.GetHeader().GetTimestamp()), genesisBlock.GetHeader())
					block.GetHeader().SetDifficulty(diff)
					tDiff := new(big.Int).Add(genesisBlock.GetHeader().GetTotalDifficulty(), diff)
					block.GetHeader().SetTotalDifficulty(tDiff)
					block.SetHash(block.ComputeHash())
					blockSig, _ := wire.BlockSign(block, sender.PrivKey().Base58())
					block.SetSignature(blockSig)
				})

				It("should return nil; No error", func() {
					validator := NewBlockValidator(block, nil, bc, true, cfg, log)
					errs := validator.checkPoW()
					Expect(errs).To(BeNil())
				})
			})
		})
	})
}
