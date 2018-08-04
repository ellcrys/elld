package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BlockValidatorTest = func() bool {
	return Describe("BlockValidator", func() {

		BeforeEach(func() {
			bc.bestChain = chain
		})

		Describe(".check", func() {
			It("should check for validation errors", func() {
				var cases = map[*wire.Block]interface{}{
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
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.mixHash, error:mix hash is required"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.difficulty, error:difficulty must be non-zero and non-negative"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.timestamp, error:timestamp must not be greater or equal to 1"),
					&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:transactions, error:at least one transaction is required"),
					&wire.Block{Header: &wire.Header{}, Transactions: []*wire.Transaction{&wire.Transaction{Type: 109}}}: fmt.Errorf("tx:0, field:type, error:unsupported transaction type"),
				}
				for b, err := range cases {
					validator := NewBlockValidator(b, nil, nil, false)
					errs := validator.check()
					Expect(errs).To(ContainElement(err))
				}
			})
		})

		Describe(".checkSignature", func() {
			It("should check for validation errors", func() {
				key := crypto.NewKeyFromIntSeed(1)
				var cases = map[*wire.Block]interface{}{
					&wire.Block{Header: &wire.Header{}}:                                     fmt.Errorf("field:header.creatorPubKey, error:empty pub key"),
					&wire.Block{Header: &wire.Header{CreatorPubKey: "invalid"}}:             fmt.Errorf("field:header.creatorPubKey, error:invalid format: version and/or checksum bytes missing"),
					&wire.Block{Header: &wire.Header{CreatorPubKey: key.PubKey().Base58()}}: fmt.Errorf("field:sig, error:signature is not valid"),
				}
				for b, err := range cases {
					validator := NewBlockValidator(b, nil, nil, false)
					errs := validator.checkSignature()
					Expect(errs).To(ContainElement(err))
				}
			})
		})

		Describe(".Validate", func() {

			BeforeEach(func() {
				_, err = bc.ProcessBlock(testdata.Block2)
				Expect(err).To(BeNil())
			})

			It("should return if block and a transaction in the block exist", func() {
				validator := NewBlockValidator(block, nil, bc, true)
				errs := validator.Validate()
				Expect(errs).To(ContainElement(fmt.Errorf("error:block found in chain")))
				Expect(errs).To(ContainElement(fmt.Errorf("tx:0, error:transaction already exist in main chain")))
			})
		})
	})
}
