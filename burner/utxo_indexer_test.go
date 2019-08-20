package burner

import (
	"fmt"
	"os"
	"time"

	"github.com/ellcrys/elld/elldb/mocks"

	"github.com/ellcrys/ltcd/btcjson"

	"github.com/ellcrys/ltcd/chaincfg/chainhash"

	"github.com/gojuno/minimock"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util/logger"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/testutil"
)

var _ = Describe("UTXOIndexer", func() {

	var cfg *config.EngineConfig
	var log logger.Logger
	var bus *emitter.Emitter
	var interrupt <-chan struct{}
	var netVersion = "test"
	var db elldb.DB

	BeforeEach(func() {
		var err error
		log = logger.NewLogrusNoOp()
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
		bus = &emitter.Emitter{}
		db = elldb.NewDB(log)
		Expect(db.Open(cfg.NetDataDir())).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
		err := os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe(".SetClient", func() {
		var client = NewRPCClientMock(GinkgoT())

		It("should set client", func() {
			utxoIndexer := NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			utxoIndexer.SetClient(client)
			Expect(utxoIndexer.client).ToNot(BeNil())
		})
	})

	Describe(".getLastIndexedHeight", func() {
		var client *RPCClientMock
		var indexer *UTXOIndexer
		var addr = "addr"

		BeforeEach(func() {
			client = NewRPCClientMock(GinkgoT())
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.SetClient(client)
		})

		It("should return 0 when last scanned height has not been recorded", func() {
			actual := indexer.getLastIndexedHeight(addr)
			Expect(actual).To(Equal(int32(0)))
		})

		When("last scanned height was previously record with value = 10", func() {
			BeforeEach(func() {
				key := MakeKeyLastScannedBlock(addr)
				kv := elldb.NewKVObject(key, util.EncodeNumber(uint64(10)))
				err := db.Put([]*elldb.KVObject{kv})
				Expect(err).To(BeNil())
			})

			It("should return 10", func() {
				actual := indexer.getLastIndexedHeight(addr)
				Expect(actual).To(Equal(int32(10)))
			})
		})
	})

	Describe(".setLastScannedHeight", func() {
		var client *RPCClientMock
		var indexer *UTXOIndexer
		var addr = "addr"

		BeforeEach(func() {
			client = NewRPCClientMock(GinkgoT())
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.SetClient(client)
		})

		BeforeEach(func() {
			tx, err := db.NewTx()
			Expect(err).To(BeNil())
			defer tx.Commit()
			err = indexer.setLastScannedHeight(tx, addr, int64(20))
			Expect(err).To(BeNil())
		})

		It("should successfully set/get the last scanned height", func() {
			actual := indexer.getLastIndexedHeight(addr)
			Expect(actual).To(Equal(int32(20)))
		})
	})

	Describe(".resetLastScannedHeightTo", func() {
		var client *RPCClientMock
		var indexer *UTXOIndexer
		var addr = "addr"
		var addr2 = "addr2"

		BeforeEach(func() {
			client = NewRPCClientMock(GinkgoT())
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.SetClient(client)
		})

		When("last scanned height for two keys are set to 20 (addr) and 30 (addr2)", func() {
			BeforeEach(func() {
				tx, err := db.NewTx()
				Expect(err).To(BeNil())
				defer tx.Commit()
				err = indexer.setLastScannedHeight(tx, addr, int64(20))
				Expect(err).To(BeNil())
				err = indexer.setLastScannedHeight(tx, addr2, int64(30))
				Expect(err).To(BeNil())
			})

			When("the reset height is 25", func() {

				BeforeEach(func() {
					err := indexer.resetLastScannedHeightTo(25)
					Expect(err).To(BeNil())
				})

				It("should leave addr unchanged", func() {
					kv := db.GetFirstOrLast(MakeKeyLastScannedBlock(addr), true)
					Expect(kv).ToNot(BeNil())
					Expect(util.DecodeNumber(kv.Value)).To(Equal(uint64(20)))
				})

				It("should change addr2 to 25", func() {
					kv := db.GetFirstOrLast(MakeKeyLastScannedBlock(addr2), true)
					Expect(kv).ToNot(BeNil())
					Expect(util.DecodeNumber(kv.Value)).To(Equal(uint64(25)))
				})
			})
		})
	})

	Describe(".getBestBlockHeight", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var indexer *UTXOIndexer

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.SetClient(client)
		})

		AfterEach(func() {
			mc.Finish()
		})

		Context("when internal RPC call to burner server returns error", func() {
			It("should return error", func() {
				e := fmt.Errorf("something bad")
				client.GetBestBlockMock.Return(nil, 0, e)
				_, err := indexer.getBestBlockHeight()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(e))
			})
		})

		Context("when best block is = 20", func() {
			It("should return 20", func() {
				client.GetBestBlockMock.Return(nil, 20, nil)
				actual, err := indexer.getBestBlockHeight()
				Expect(err).To(BeNil())
				Expect(actual).To(Equal(int32(20)))
			})
		})
	})

	Describe(".getBlock", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var indexer *UTXOIndexer

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.SetClient(client)
		})

		AfterEach(func() {
			mc.Finish()
		})

		It("should return error if no block at the given height is unknown", func() {
			e := fmt.Errorf("not found")
			client.GetBlockHashMock.Return(nil, e)
			h, err := indexer.getBlock(1000)
			Expect(err).To(Equal(e))
			Expect(h).To(BeNil())
		})

		It("should return hash if block exists", func() {
			hashStr := "c694ce78fe9555168d66c503f0709a1ff715e196a50a4931a8afb1ee388d1370"
			hash, _ := chainhash.NewHashFromStr(hashStr)
			client.GetBlockHashMock.When(1000).Then(hash, nil)
			client.GetBlockVerboseTxMock.When(hash).Then(&btcjson.GetBlockVerboseResult{
				Hash: hash.String(),
			}, nil)

			actual, err := indexer.getBlock(1000)
			Expect(err).To(BeNil())
			Expect(actual.Hash).To(Equal(hash.String()))
		})
	})

	Describe(".getUTXOs", func() {
		When("an address has 2 utxo entries", func() {
			var addr = "addr1"

			BeforeEach(func() {
				utxo1Key := MakeKeyAddressUTXO(1, addr, "txHash1", 0)
				utxoDoc1 := DocUTXO{TxID: "txHash1", Index: 0, Value: 10}
				o1 := elldb.NewKVObject(utxo1Key, util.ObjectToBytes(utxoDoc1))
				err := db.Put([]*elldb.KVObject{o1})
				Expect(err).To(BeNil())

				utxo2Key := MakeKeyAddressUTXO(1, addr, "txHash2", 0)
				utxoDoc2 := DocUTXO{TxID: "txHash2", Index: 0, Value: 20}
				o2 := elldb.NewKVObject(utxo2Key, util.ObjectToBytes(utxoDoc2))
				err = db.Put([]*elldb.KVObject{o2})
				Expect(err).To(BeNil())
			})

			It("should return all UTXOs", func() {
				utxos := getUTXOs(db, addr)
				Expect(utxos).To(HaveLen(2))
			})
		})

		When("an address has no utxo entries", func() {
			var addr = "addr1"
			It("should return zero (0) UTXOs", func() {
				utxos := getUTXOs(db, addr)
				Expect(utxos).To(HaveLen(0))
			})
		})
	})

	Describe(".BalanceOf", func() {
		When("an address has 2 utxo entries with value 10 and 20 respectively", func() {
			var addr = "addr1"

			BeforeEach(func() {
				utxo1Key := MakeKeyAddressUTXO(1, addr, "txHash1", 0)
				utxoDoc1 := DocUTXO{TxID: "txHash1", Index: 0, Value: 10}
				o1 := elldb.NewKVObject(utxo1Key, util.ObjectToBytes(utxoDoc1))
				err := db.Put([]*elldb.KVObject{o1})
				Expect(err).To(BeNil())

				utxo2Key := MakeKeyAddressUTXO(1, addr, "txHash2", 0)
				utxoDoc2 := DocUTXO{TxID: "txHash2", Index: 0, Value: 20}
				o2 := elldb.NewKVObject(utxo2Key, util.ObjectToBytes(utxoDoc2))
				err = db.Put([]*elldb.KVObject{o2})
				Expect(err).To(BeNil())
			})

			It("should return balance equal to 30", func() {
				balance := balanceOf(db, addr)
				Expect(balance).To(Equal("30"))
			})
		})

		When("an address has no utxo entries", func() {
			var addr = "addr1"
			It("should return balance equal to zero (0)", func() {
				balance := balanceOf(db, addr)
				Expect(balance).To(Equal("0"))
			})
		})
	})

	Describe(".HasStopped", func() {
		var indexer *UTXOIndexer
		It("should return false/true if 'stop' flag is false/true", func() {
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.stop = false
			Expect(indexer.HasStopped()).To(BeFalse())
			indexer.stop = true
			Expect(indexer.HasStopped()).To(BeTrue())
		})
	})

	Describe(".Stop", func() {
		var indexer *UTXOIndexer

		It("should set stop to true", func() {
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.Stop()
			Expect(indexer.stop).To(BeTrue())
		})

		When("an unstopped indexer ticker is set", func() {
			It("should set the ticker to nil", func() {
				indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
				indexer.indexerTicker = time.NewTicker(10 * time.Minute)
				indexer.Stop()
				Expect(indexer.indexerTicker).To(BeNil())
			})
		})
	})

	Describe(".index", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var indexer *UTXOIndexer
		var dbMock *mocks.DBMock
		var addr = "QaPhoSV8VBgs2VCQQuxKCQdRFukM9th2T4"
		var block *btcjson.GetBlockVerboseResult

		BeforeEach(func() {
			block = &btcjson.GetBlockVerboseResult{
				RawTx: []btcjson.TxRawResult{
					btcjson.TxRawResult{
						Txid: "tx_id",
						Hash: "tx_hash",
						Vout: []btcjson.Vout{{Value: 10, N: 0, ScriptPubKey: btcjson.ScriptPubKeyResult{Hex: "hex1234", Addresses: []string{addr}}}},
					},
				},
			}
		})

		Context("when unable to create a db transaction", func() {
			BeforeEach(func() {
				dbMock = mocks.NewDBMock(mc)
				client = NewRPCClientMock(mc)
				indexer = NewUTXOIndexer(cfg, log, dbMock, netVersion, bus, interrupt)
				indexer.SetClient(client)
			})

			AfterEach(func() {
				mc.Finish()
			})

			It("should return error if unable to create db transaction", func() {
				e := fmt.Errorf("failed to create db transaction")
				dbMock.NewTxMock.Return(nil, e)
				err := indexer.index("", nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(e.Error()))
			})
		})

		Context("when finished", func() {
			var height = int32(12)

			BeforeEach(func() {
				indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
				block = &btcjson.GetBlockVerboseResult{
					Height: int64(height),
				}
			})

			It("should set the block's height as the last scanned height", func() {
				err := indexer.index(addr, block, nil)
				Expect(err).To(BeNil())
				res := indexer.getLastIndexedHeight(addr)
				Expect(res).To(Equal(height))
			})
		})

		Context("block with one tx that has one output", func() {
			var height = int32(12)

			BeforeEach(func() {
				block.Height = int64(height)
				block.RawTx[0].Vout[0].ScriptPubKey.Addresses = []string{addr}
				indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
				err := indexer.index(addr, block, nil)
				Expect(err).To(BeNil())
			})

			Specify("that the output was indexed", func() {
				txID := block.RawTx[0].Txid
				outIndex := block.RawTx[0].Vout[0].N
				key := MakeQueryKeyAddressUTXO(addr, txID, outIndex)
				res := db.GetFirstOrLast(key, true)
				Expect(res).ToNot(BeNil())
			})
		})

		Context("block with one tx whose single output has an unknown address", func() {
			var height = int32(12)

			BeforeEach(func() {
				indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
				block.Height = int64(height)
				block.RawTx[0].Vout[0].ScriptPubKey.Addresses = []string{"unknown"}
				err := indexer.index(addr, block, nil)
				Expect(err).To(BeNil())
			})

			Specify("that the output was not indexed", func() {
				txID := block.RawTx[0].Txid
				outIndex := block.RawTx[0].Vout[0].N
				key := MakeQueryKeyAddressUTXO(addr, txID, outIndex)
				res := db.GetFirstOrLast(key, true)
				Expect(res).To(BeNil())
			})

		})

		Context("when output has already been indexed", func() {
			var height = int32(12)
			var mc = minimock.NewController(GinkgoT())
			var indexer *UTXOIndexer
			var dbMock *mocks.DBMock

			// Goal: Mock the result of finding the already indexed output
			// to return a hit.
			BeforeEach(func() {
				txMock := mocks.NewTxMock(mc)
				outputKV := elldb.NewKVObject([]byte{}, []byte{})
				txMock.GetByPrefixMock.Return([]*elldb.KVObject{outputKV})
				txMock.PutMock.Return(nil)
				txMock.CommitMock.Return(nil)

				dbMock = mocks.NewDBMock(mc)
				dbMock.NewTxMock.Return(txMock, nil)
				indexer = NewUTXOIndexer(cfg, log, dbMock, netVersion, bus, interrupt)
			})

			BeforeEach(func() {
				block.Height = int64(height)
				block.RawTx[0].Vout[0].ScriptPubKey.Addresses = []string{addr}
				err := indexer.index(addr, block, nil)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				mc.Finish()
			})

			It("should not re-index known output", func() {
				txID := block.RawTx[0].Txid
				outIndex := block.RawTx[0].Vout[0].N
				key := MakeQueryKeyAddressUTXO(addr, txID, outIndex)
				res := db.GetFirstOrLast(key, true)
				Expect(res).To(BeNil())
			})
		})

		Context("when an address has an indexed output", func() {
			var height = int32(12)

			BeforeEach(func() {
				block.Height = int64(height)
				block.RawTx[0].Vout[0].ScriptPubKey.Addresses = []string{addr}
				indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
				err := indexer.index(addr, block, nil)
				Expect(err).To(BeNil())

				txID := block.RawTx[0].Txid
				outIndex := block.RawTx[0].Vout[0].N
				key := MakeQueryKeyAddressUTXO(addr, txID, outIndex)
				res := db.GetFirstOrLast(key, true)
				Expect(res).ToNot(BeNil())
			})

			Context("then another block spends the output", func() {
				var txID string
				var outIndex uint32

				BeforeEach(func() {
					txID = block.RawTx[0].Txid
					outIndex = block.RawTx[0].Vout[0].N

					block.Height = int64(height)
					block.RawTx[0].Vin = []btcjson.Vin{{
						Txid: txID,
						Vout: outIndex,
					}}

					block.RawTx[0].Vout = []btcjson.Vout{}
				})

				Specify("that the output is removed", func() {
					indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
					err := indexer.index(addr, block, nil)
					Expect(err).To(BeNil())

					key := MakeQueryKeyAddressUTXO(addr, txID, outIndex)
					res := db.GetFirstOrLast(key, true)
					Expect(res).To(BeNil())
				})
			})
		})
	})

	Describe(".begin", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var indexer *UTXOIndexer
		var addr = "QaPhoSV8VBgs2VCQQuxKCQdRFukM9th2T4"
		var interrupt chan struct{}
		var dbMock *mocks.DBMock

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)
			indexer.SetClient(client)
			interrupt = make(chan struct{})
			dbMock = mocks.NewDBMock(mc)
		})

		It("should return error if unable to determine the burner chain best block", func() {
			client.GetBestBlockMock.Return(nil, 0, fmt.Errorf("error"))
			r := indexer.begin(1, addr, 0, false, interrupt)
			Expect(r.err).ToNot(BeNil())
			Expect(r.err.Error()).To(Equal("Failed to get best block height of the burner chain: error"))
		})

		Context("best block height is 10", func() {

			BeforeEach(func() {
				indexer = NewUTXOIndexer(cfg, log, dbMock, netVersion, bus, interrupt)
				indexer.SetClient(client)
				client.GetBestBlockMock.Return(nil, 10, nil)
			})

			When("last indexed height is greater than best block height", func() {

				BeforeEach(func() {
					key := MakeKeyLastScannedBlock(addr)
					kv := elldb.NewKVObject(key, util.EncodeNumber(11))
					dbMock.GetFirstOrLastMock.Return(kv)
				})

				It("should return status = indexUpToDate", func() {
					r := indexer.begin(1, addr, 0, false, interrupt)
					Expect(r.err).To(BeNil())
					Expect(r.status).To(Equal(indexUpToDate))
				})
			})

			When("the indexer is been stopped", func() {

				BeforeEach(func() {
					key := MakeKeyLastScannedBlock(addr)
					kv := elldb.NewKVObject(key, util.EncodeNumber(1))
					dbMock.GetFirstOrLastMock.Return(kv)
					indexer.stop = true
				})

				It("should return status = indexUpToDate", func() {
					r := indexer.begin(1, addr, 0, false, interrupt)
					Expect(r.err).To(BeNil())
					Expect(r.status).To(Equal(stoppedDueToShutdown))
				})
			})
		})

		When("the `interrupt` chan is closed", func() {

			BeforeEach(func() {
				close(interrupt)
			})

			It("should return status = indexUpToDate", func() {
				r := indexer.begin(1, addr, 0, false, interrupt)
				Expect(r.err).To(BeNil())
				Expect(r.status).To(Equal(stoppedDueToInterrupt))
			})
		})

		AfterEach(func() {
			mc.Finish()
		})
	})

	Describe(".deleteIndexFrom", func() {
		var indexer *UTXOIndexer

		Context("with keys representing block 10,11,12,13, deleting from 11", func() {
			BeforeEach(func() {
				indexer = NewUTXOIndexer(cfg, log, db, netVersion, bus, interrupt)

				oKey := MakeKeyAddressUTXO(10, "addr1", "abc", 0)
				o := elldb.NewKVObject(oKey, []byte{})
				err := db.Put([]*elldb.KVObject{o})
				Expect(err).To(BeNil())

				oKey2 := MakeKeyAddressUTXO(11, "addr2", "abc", 0)
				o2 := elldb.NewKVObject(oKey2, []byte{})
				err = db.Put([]*elldb.KVObject{o2})
				Expect(err).To(BeNil())

				oKey3 := MakeKeyAddressUTXO(12, "addr3", "abc", 0)
				o3 := elldb.NewKVObject(oKey3, []byte{})
				err = db.Put([]*elldb.KVObject{o3})
				Expect(err).To(BeNil())

				oKey4 := MakeKeyAddressUTXO(13, "addr4", "abc", 0)
				o4 := elldb.NewKVObject(oKey4, []byte{})
				err = db.Put([]*elldb.KVObject{o4})
				Expect(err).To(BeNil())

				err = indexer.deleteIndexFrom(11)
				Expect(err).To(BeNil())
			})

			It("should delete keys representing block 11,12,13", func() {
				result := db.GetByPrefix(nil)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Key).To(Equal(util.EncodeNumber(10)))
			})
		})

	})
})
