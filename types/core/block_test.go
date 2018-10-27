package core

import (
	"math/big"
	"testing"

	. "github.com/ncodes/goblin"

	"github.com/vmihailenco/msgpack"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/gomega"
)

func TestBlockAndHeader(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Block & Header", func() {

		var key = crypto.NewKeyFromIntSeed(1)

		g.Describe("Header.Bytes", func() {
			g.It("should successfully return bytes", func() {
				h := Header{
					ParentHash:       util.BytesToHash([]byte("parent_hash")),
					Number:           1,
					TransactionsRoot: util.BytesToHash([]byte("state_root_hash")),
					StateRoot:        util.BytesToHash([]byte("tx_hash")),
					Nonce:            util.EncodeNonce(1),
					Difficulty:       new(big.Int).SetUint64(100),
					Timestamp:        1500000,
				}
				expected := []uint8{154, 196, 32, 112, 97, 114, 101, 110, 116, 95, 104, 97, 115, 104, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 160, 196, 32, 115, 116, 97, 116, 101, 95, 114, 111, 111, 116, 95, 104, 97, 115, 104, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 196, 32, 116, 120, 95, 104, 97, 115, 104, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 196, 1, 100, 196, 0, 206, 0, 22, 227, 96, 196, 8, 0, 0, 0, 0, 0, 0, 0, 1, 192}
				Expect(h.GetBytes()).To(Equal(expected))
			})
		})

		g.Describe("Header.Copy", func() {
			g.It("should successfully copy", func() {
				h := Header{
					ParentHash:       util.BytesToHash([]byte("parent_hash")),
					Number:           1,
					TransactionsRoot: util.BytesToHash([]byte("state_root_hash")),
					StateRoot:        util.BytesToHash([]byte("tx_hash")),
					Nonce:            util.EncodeNonce(1),
					Difficulty:       new(big.Int).SetUint64(100),
					Timestamp:        1500000,
				}

				h2 := h.Copy()
				Expect(h2).ToNot(Equal(h))
				Expect(h.ParentHash).To(Equal(h2.GetParentHash()))
				h2.SetParentHash(util.StrToHash("xyz"))

				Expect(h.ParentHash).ToNot(Equal(h2.GetParentHash()))
				h2.SetNumber(10)
				Expect(h.Number).ToNot(Equal(h2.GetNumber()))

				h2.SetTransactionsRoot(util.StrToHash("abc"))
				Expect(h.TransactionsRoot).ToNot(Equal(h2.GetTransactionsRoot()))

				h2.SetNonce(util.EncodeNonce(10))
				Expect(h.Nonce).ToNot(Equal(h2.GetNonce()))
			})
		})

		g.Describe("Header.ComputeHash", func() {
			g.It("should successfully return 32 bytes digest", func() {
				h := Header{
					ParentHash:       util.BytesToHash([]byte("parent_hash")),
					Number:           1,
					TransactionsRoot: util.BytesToHash([]byte("state_root_hash")),
					StateRoot:        util.BytesToHash([]byte("tx_hash")),
					Nonce:            util.EncodeNonce(1),
					Difficulty:       new(big.Int).SetUint64(100),
					Timestamp:        1500000,
				}
				actual := h.ComputeHash()
				expected := util.Hash([32]byte{207, 192, 165, 248, 237, 83, 165, 106, 166, 54, 3, 115, 19, 178, 25, 184, 128, 80, 228, 253, 152, 55, 132, 74, 228, 163, 198, 218, 47, 218, 75, 94})
				Expect(actual).To(HaveLen(32))
				Expect(actual).To(Equal(expected))
			})
		})

		g.Describe(".BlockSign", func() {

			var header1 = &Header{
				ParentHash:       util.BytesToHash([]byte("parent_hash")),
				Number:           1,
				TransactionsRoot: util.BytesToHash([]byte("state_root_hash")),
				StateRoot:        util.BytesToHash([]byte("tx_hash")),
				Nonce:            util.EncodeNonce(1),
				Difficulty:       new(big.Int).SetUint64(100),
				Timestamp:        1500000,
			}

			g.It("should successfully sign a block", func() {

				b := Block{
					Transactions: []*Transaction{
						{Type: TxTypeBalance, SenderPubKey: util.String(key.PubKey().Base58()), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
					},
					Header: header1,
				}

				t1 := b.Transactions[0]
				t1.Hash = t1.ComputeHash()

				t1Sig, err := TxSign(t1, key.PrivKey().Base58())
				Expect(err).To(BeNil())
				t1.Sig = t1Sig

				b.Hash = b.ComputeHash()
				bs, err := BlockSign(&b, key.PrivKey().Base58())
				Expect(err).To(BeNil())
				Expect(bs).ToNot(BeEmpty())
			})
		})

		g.Describe(".BlockVerify", func() {

			var header1 = &Header{
				ParentHash:       util.BytesToHash([]byte("parent_hash")),
				Number:           1,
				TransactionsRoot: util.BytesToHash([]byte("state_root_hash")),
				StateRoot:        util.BytesToHash([]byte("tx_hash")),
				Nonce:            util.EncodeNonce(1),
				Difficulty:       new(big.Int).SetUint64(100),
				Timestamp:        1500000,
			}

			g.It("should return err if creator pub key is not set in header", func() {

				b := Block{
					Transactions: []*Transaction{
						{Type: TxTypeBalance, SenderPubKey: util.String(key.PubKey().Base58()), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
					},
					Header: header1,
				}

				err := BlockVerify(&b)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("field:header.creatorPubKey, error:creator public not set"))
			})

			g.It("should return err if creator signature is not set in header", func() {

				b := Block{
					Transactions: []*Transaction{
						{Type: TxTypeBalance, SenderPubKey: util.String(key.PubKey().Base58()), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
					},
					Header: header1,
				}
				b.GetHeader().SetCreatorPubKey(util.String(key.PubKey().Base58()))

				err := BlockVerify(&b)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("field:sig, error:signature not set"))
			})

			g.It("should return error if signature is invalid", func() {

				b := Block{
					Transactions: []*Transaction{
						{Type: TxTypeBalance, SenderPubKey: util.String(key.PubKey().Base58()), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
					},
					Header: header1,
				}
				b.Header.CreatorPubKey = util.String(key.PubKey().Base58())

				t1 := b.Transactions[0]
				t1.Hash = t1.ComputeHash()

				t1Sig, err := TxSign(t1, key.PrivKey().Base58())
				Expect(err).To(BeNil())
				t1.Sig = t1Sig

				b.Hash = b.ComputeHash()

				b.Sig = []byte("invalid")
				err = BlockVerify(&b)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("block verification failed"))
			})

			g.It("should successfully verify a block", func() {

				b := Block{
					Transactions: []*Transaction{
						{Type: TxTypeBalance, SenderPubKey: util.String(key.PubKey().Base58()), To: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", Value: "100.30", Timestamp: 1529670647, Fee: "0.10"},
					},
					Header: header1,
				}
				b.Header.CreatorPubKey = util.String(key.PubKey().Base58())

				t1 := b.Transactions[0]
				t1.Hash = t1.ComputeHash()

				t1Sig, err := TxSign(t1, key.PrivKey().Base58())
				Expect(err).To(BeNil())
				t1.Sig = t1Sig

				b.Hash = b.ComputeHash()
				bs, err := BlockSign(&b, key.PrivKey().Base58())
				Expect(err).To(BeNil())
				Expect(bs).ToNot(BeEmpty())

				b.Sig = bs
				err = BlockVerify(&b)
				Expect(err).To(BeNil())
			})
		})

		g.Describe("Block.ComputeHash", func() {
			g.It("should successfully return 32 bytes digest", func() {
				b := Block{
					Transactions: []*Transaction{
						{
							Type:         TxTypeBalance,
							SenderPubKey: "48qgD4WR71u2fMJJNdsXmfDKNqNmiFdVo3YfnGjTA915cArTUTw",
							To:           "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
							From:         "e9L9UNbcxCrnAEc8ARUnkiVrDJjW57MKdq",
							Value:        "100.30",
							Timestamp:    156707944,
							Fee:          "0.10"},
					},
					Header: &Header{
						ParentHash:       util.BytesToHash([]byte("parent_hash")),
						Number:           1,
						TransactionsRoot: util.BytesToHash([]byte("state_root_hash")),
						StateRoot:        util.BytesToHash([]byte("tx_hash")),
						Nonce:            util.EncodeNonce(1),
						Difficulty:       new(big.Int).SetUint64(100),
						Timestamp:        1500000,
					},
				}

				actual := b.ComputeHash()
				expected := util.BytesToHash([]byte{204, 90, 37, 106, 233, 100, 37, 67, 190, 189, 127, 23, 160, 227, 74, 47, 5, 7, 232, 123, 218, 160, 255, 28, 38, 1, 181, 247, 121, 20, 246, 226})
				Expect(actual).To(Equal(expected))
			})
		})

		g.Describe("Header serialization", func() {
			g.It("should serialize and deserialize as expected", func() {
				h := &Header{
					Number:           10,
					Nonce:            util.EncodeNonce(10),
					Timestamp:        100,
					CreatorPubKey:    "abc",
					ParentHash:       util.StrToHash("xyz"),
					StateRoot:        util.StrToHash("abc"),
					TransactionsRoot: util.StrToHash("abc"),
					Extra:            []byte("abc"),
					Difficulty:       new(big.Int).SetInt64(30000),
					TotalDifficulty:  new(big.Int).SetInt64(40000),
				}
				bs, err := msgpack.Marshal(h)
				Expect(err).To(BeNil())
				Expect(bs).ToNot(BeEmpty())

				var h2 Header
				err = msgpack.Unmarshal(bs, &h2)
				Expect(err).To(BeNil())
				Expect(&h2).To(Equal(h))
			})
		})

		g.Describe("Block serialization", func() {
			g.It("should serialize and deserialize as expected", func() {
				b := &Block{
					Header: &Header{
						Number:           10,
						Nonce:            util.EncodeNonce(10),
						Timestamp:        100,
						CreatorPubKey:    "abc",
						ParentHash:       util.StrToHash("xyz"),
						StateRoot:        util.StrToHash("abc"),
						TransactionsRoot: util.StrToHash("abc"),
						Extra:            []byte("abc"),
						Difficulty:       new(big.Int).SetInt64(30000),
						TotalDifficulty:  new(big.Int).SetInt64(40000),
					},
					Transactions: []*Transaction{
						{
							Type:         1,
							Nonce:        1,
							To:           "some_address",
							SenderPubKey: "abc",
							Value:        "120",
							Timestamp:    12345,
							Fee:          "0.22",
							InvokeArgs: &InvokeArgs{
								Func: "doStuff",
								Params: map[string][]byte{
									"age": []byte("1000"),
								},
							},
							Sig:  []byte("xyz"),
							Hash: util.StrToHash("abcdef"),
						},
					},
				}
				bs, err := msgpack.Marshal(b)
				Expect(err).To(BeNil())
				Expect(bs).ToNot(BeEmpty())

				var b2 Block
				err = msgpack.Unmarshal(bs, &b2)
				Expect(err).To(BeNil())
				Expect(&b2).To(Equal(b))
			})
		})
	})
}