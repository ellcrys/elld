package wire

import (
	"fmt"

	"github.com/ellcrys/elld/crypto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Block & Header", func() {

	var key = crypto.NewKeyFromIntSeed(1)

	Describe("Header.validate", func() {
		It("should validate fields", func() {

			_key := key.PubKey().Base58()

			var data = map[Header]error{
				Header{}: fmt.Errorf("field:parentHash, error:expected 66 characters"),
				Header{Number: 0, ParentHash: "some_hash"}:                                                                                                                                                              fmt.Errorf("field:number, error:number must be greater or equal to 1"),
				Header{Number: 0, ParentHash: "some_hash"}:                                                                                                                                                              fmt.Errorf("field:parentHash, error:expected 66 characters"),
				Header{Number: 1, ParentHash: "some_hash"}:                                                                                                                                                              fmt.Errorf("field:parentHash, error:should be empty since block number is 1"),
				Header{Number: 1}:                                                                                                                                                                                       fmt.Errorf("field:creatorPubKey, error:creator public key is required"),
				Header{Number: 1, CreatorPubKey: _key}:                                                                                                                                                                  fmt.Errorf("field:transactionsRoot, error:expected 66 characters"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key}:                                                                                                                                                 fmt.Errorf("field:number, error:number must be greater or equal to 1"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2}:                                                                                                                                      fmt.Errorf("field:transactionsRoot, error:expected 66 characters"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, StateRoot: RandString(66)}:                                                                                                           fmt.Errorf("field:transactionsRoot, error:expected 66 characters"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, TransactionsRoot: RandString(66)}:                                                                                                    fmt.Errorf("field:stateRoot, error:expected 66 characters"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, StateRoot: RandString(66), TransactionsRoot: RandString(66)}:                                                                         fmt.Errorf("field:nonce, error:must not be zero"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, StateRoot: RandString(66), TransactionsRoot: RandString(66), Nonce: 1}:                                                               fmt.Errorf("field:mixHash, error:expected 32 characters"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, StateRoot: RandString(66), TransactionsRoot: RandString(66), Nonce: 1, MixHash: RandString(32)}:                                      fmt.Errorf("field:difficulty, error:must be non-zero and non-negative"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, StateRoot: RandString(66), TransactionsRoot: RandString(66), Nonce: 1, MixHash: RandString(32), Difficulty: "ac"}:                    fmt.Errorf("field:difficulty, error:must be numeric"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, StateRoot: RandString(66), TransactionsRoot: RandString(66), Nonce: 1, MixHash: RandString(32), Difficulty: "1"}:                     fmt.Errorf("field:timestamp, error:must not be empty or a negative value"),
				Header{ParentHash: RandString(66), CreatorPubKey: _key, Number: 2, StateRoot: RandString(66), TransactionsRoot: RandString(66), Nonce: 1, MixHash: RandString(32), Difficulty: "1", Timestamp: 1500000}: nil,
			}

			for h, e := range data {
				err := h.Validate()
				if e != nil {
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal(e.Error()))
				} else {
					Expect(err).To(BeNil())
				}
			}
		})
	})

	Describe("Header.Bytes", func() {
		It("should successfully return bytes", func() {
			h := Header{
				ParentHash:       "parent_hash",
				Number:           1,
				TransactionsRoot: "state_root_hash",
				StateRoot:        "tx_hash",
				Nonce:            1,
				MixHash:          "mixHash",
				Difficulty:       "1",
				Timestamp:        1500000,
			}
			expected := []byte{48, 62, 12, 11, 112, 97, 114, 101, 110, 116, 95, 104, 97, 115, 104, 19, 1, 49, 12, 15, 115, 116, 97, 116, 101, 95, 114, 111, 111, 116, 95, 104, 97, 115, 104, 12, 7, 116, 120, 95, 104, 97, 115, 104, 19, 1, 49, 19, 7, 109, 105, 120, 72, 97, 115, 104, 19, 1, 49, 2, 3, 22, 227, 96}
			Expect(h.Bytes()).To(Equal(expected))
		})
	})

	Describe("Header.ComputeHash", func() {
		It("should successfully return 32 bytes digest", func() {
			h := Header{
				ParentHash:       "parent_hash",
				Number:           1,
				TransactionsRoot: "state_root_hash",
				StateRoot:        "tx_hash",
				Nonce:            1,
				MixHash:          "mixHash",
				Difficulty:       "1",
				Timestamp:        1500000,
			}
			actual := h.ComputeHash()
			Expect(actual).To(Equal("0x30beb61284f96a98db1210f07c43430d63dbdd8af42f159464e0df0bf06eaa8d"))
		})
	})

	Describe(".BlockSign", func() {

		var key = crypto.NewKeyFromIntSeed(1)
		var header1 = &Header{
			ParentHash:       RandString(66),
			Number:           2,
			TransactionsRoot: RandString(66),
			StateRoot:        RandString(66),
			Nonce:            1,
			MixHash:          RandString(32),
			Difficulty:       "1",
			Timestamp:        1529670647,
		}

		It("should successfully sign a block", func() {

			b := Block{
				Transactions: []*Transaction{
					&Transaction{Type: TxTypeBalance, SenderPubKey: key.PubKey().Base58(), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
				},
				Header: header1,
				Hash:   "",
			}

			t1 := b.Transactions[0]
			t1.Hash = ToHex(t1.ComputeHash())

			t1Sig, err := TxSign(t1, key.PrivKey().Base58())
			Expect(err).To(BeNil())
			t1.Sig = ToHex(t1Sig)

			b.Hash = b.ComputeHash()
			bs, err := BlockSign(&b, key.PrivKey().Base58())
			Expect(err).To(BeNil())
			Expect(bs).ToNot(BeEmpty())
		})
	})

	Describe(".BlockVerify", func() {
		var key = crypto.NewKeyFromIntSeed(1)
		var header1 = &Header{
			ParentHash:       RandString(66),
			Number:           2,
			TransactionsRoot: RandString(66),
			StateRoot:        RandString(66),
			Nonce:            1,
			MixHash:          RandString(32),
			Difficulty:       "1",
			Timestamp:        1529670647,
		}

		It("should return err if creator pub key is not set in header", func() {

			b := Block{
				Transactions: []*Transaction{
					&Transaction{Type: TxTypeBalance, SenderPubKey: key.PubKey().Base58(), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
				},
				Header: header1,
				Hash:   "",
			}

			err := BlockVerify(&b)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field:header.creatorPubKey, error:creator public not set"))
		})

		It("should return err if creator signature is not set in header", func() {

			b := Block{
				Transactions: []*Transaction{
					&Transaction{Type: TxTypeBalance, SenderPubKey: key.PubKey().Base58(), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
				},
				Header: header1,
				Hash:   "",
			}
			b.Header.CreatorPubKey = key.PubKey().Base58()

			err := BlockVerify(&b)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field:sig, error:signature not set"))
		})

		It("should return error if signature is invalid", func() {

			b := Block{
				Transactions: []*Transaction{
					&Transaction{Type: TxTypeBalance, SenderPubKey: key.PubKey().Base58(), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
				},
				Header: header1,
				Hash:   "",
			}
			b.Header.CreatorPubKey = key.PubKey().Base58()

			t1 := b.Transactions[0]
			t1.Hash = ToHex(t1.ComputeHash())

			t1Sig, err := TxSign(t1, key.PrivKey().Base58())
			Expect(err).To(BeNil())
			t1.Sig = ToHex(t1Sig)

			b.Hash = b.ComputeHash()

			b.Sig = ToHex([]byte("invalid"))
			err = BlockVerify(&b)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("block verification failed"))
		})

		It("should successfully verify a block", func() {

			b := Block{
				Transactions: []*Transaction{
					&Transaction{Type: TxTypeBalance, SenderPubKey: key.PubKey().Base58(), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
				},
				Header: header1,
				Hash:   "",
			}
			b.Header.CreatorPubKey = key.PubKey().Base58()

			t1 := b.Transactions[0]
			t1.Hash = ToHex(t1.ComputeHash())

			t1Sig, err := TxSign(t1, key.PrivKey().Base58())
			Expect(err).To(BeNil())
			t1.Sig = ToHex(t1Sig)

			b.Hash = b.ComputeHash()
			bs, err := BlockSign(&b, key.PrivKey().Base58())
			Expect(err).To(BeNil())
			Expect(bs).ToNot(BeEmpty())

			b.Sig = ToHex(bs)
			err = BlockVerify(&b)
			Expect(err).To(BeNil())
		})
	})

	Describe("Block.ComputeHash", func() {
		It("should successfully return 32 bytes digest", func() {

			b := Block{
				Transactions: []*Transaction{
					&Transaction{
						Type:         TxTypeBalance,
						SenderPubKey: "48qgD4WR71u2fMJJNdsXmfDKNqNmiFdVo3YfnGjTA915cArTUTw",
						To:           "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
						From:         "e9L9UNbcxCrnAEc8ARUnkiVrDJjW57MKdq",
						Value:        "100.30",
						Timestamp:    156707944,
						Fee:          "0.10"},
				},
				Header: &Header{
					ParentHash:       "parent_hash",
					Number:           1,
					TransactionsRoot: "state_root_hash",
					StateRoot:        "tx_hash",
					Nonce:            1,
					MixHash:          "mixHash",
					Difficulty:       "1",
					Timestamp:        1500000,
				},
			}

			actual := b.ComputeHash()
			Expect(actual).To(Equal("0x05ec83603ef38b226994c36e4be0ebecb9b146953c20a45b240a3d3f24ad70f1"))
		})
	})

})
