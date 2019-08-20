package burner

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/params"

	"github.com/ellcrys/elld/crypto"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/ltcd/btcjson"
	"github.com/ellcrys/ltcd/chaincfg/chainhash"
	"github.com/gojuno/minimock"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UTXOIndexer", func() {

	var cfg *config.EngineConfig
	var log logger.Logger
	var bus *emitter.Emitter
	var interrupt <-chan struct{}
	var db elldb.DB
	var minimumBurnAmt = params.MinimumBurnAmt

	BeforeEach(func() {
		var err error
		log = logger.NewLogrusNoOp()
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
		bus = &emitter.Emitter{}
		db = elldb.NewDB(log)
		Expect(db.Open(cfg.NetDataDir())).To(BeNil())
		cfg.G().Bus = bus
		params.MinimumBurnAmt = decimal.NewFromFloat(1)
	})

	AfterEach(func() {
		db.Close()
		err := os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
		params.MinimumBurnAmt = minimumBurnAmt
	})

	Describe(".openDB", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var tm *TicketManager

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			tm = NewTicketManager(cfg, client, interrupt)
			_, err := os.Stat(cfg.GetTicketDBPath())
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})

		It("should open the database without error", func() {
			err := tm.openDB()
			Expect(err).To(BeNil())
			_, err = os.Stat(cfg.GetTicketDBPath())
			Expect(err).To(BeNil())
		})
	})

	Describe(".getBestBlockHeight", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var tm *TicketManager

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			tm = NewTicketManager(cfg, client, interrupt)
		})

		It("should return err if it failed", func() {
			e := fmt.Errorf("error")
			client.GetBestBlockMock.Return(nil, 0, e)
			height, err := tm.getBestBlockHeight()
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(e))
			Expect(height).To(BeZero())
		})

		It("should return expected height", func() {
			client.GetBestBlockMock.Return(nil, 10, nil)
			height, err := tm.getBestBlockHeight()
			Expect(err).To(BeNil())
			Expect(height).To(Equal(int32(10)))
		})
	})

	Describe(".getBlock", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var tm *TicketManager

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			tm = NewTicketManager(cfg, client, interrupt)
		})

		AfterEach(func() {
			mc.Finish()
		})

		It("should return error if no block at the given height is unknown", func() {
			e := fmt.Errorf("not found")
			client.GetBlockHashMock.Return(nil, e)
			h, err := tm.getBlock(1000)
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

			actual, err := tm.getBlock(1000)
			Expect(err).To(BeNil())
			Expect(actual.Hash).To(Equal(hash.String()))
		})
	})

	Describe(".getLastIndexedHeight", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var tm *TicketManager

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			tm = NewTicketManager(cfg, client, interrupt)
			err := tm.openDB()
			Expect(err).To(BeNil())
		})

		It("should return zero(0) when last_height meta key is unset", func() {
			h := tm.getLastIndexedHeight()
			Expect(h).To(Equal(int64(0)))
		})

		When("last_height meta key is set to 10", func() {
			BeforeEach(func() {
				err := tm.db.Create(&ticketManagerMeta{Key: "last_height", Value: 10}).Error
				Expect(err).To(BeNil())
			})

			It("should return ten(10)", func() {
				h := tm.getLastIndexedHeight()
				Expect(h).To(Equal(int64(10)))
			})
		})
	})

	Describe(".deleteTicketFromHeight", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var tm *TicketManager

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			tm = NewTicketManager(cfg, client, interrupt)
			err := tm.openDB()
			Expect(err).To(BeNil())
		})

		When("the specified height is 25 and there are ticket1 at height 20 and ticket2 at height 30", func() {
			BeforeEach(func() {
				err := tm.db.Create(&Ticket{TxID: "ticket1", BlockHeight: 20}).Error
				Expect(err).To(BeNil())
				err = tm.db.Create(&Ticket{TxID: "ticket2", BlockHeight: 30}).Error
				Expect(err).To(BeNil())
				err = tm.deleteTicketFromHeight(25)
				Expect(err).To(BeNil())
			})

			It("should delete ticket2 since it is after the given height", func() {
				var count int
				err := tm.db.Model(&Ticket{}).Where(&Ticket{TxID: "ticket2"}).Count(&count).Error
				Expect(err).To(BeNil())
				Expect(count).To(Equal(0))
			})

			It("should not delete ticket1 since it is before the given height", func() {
				var count int
				err := tm.db.Model(&Ticket{}).Where(&Ticket{TxID: "ticket1"}).Count(&count).Error
				Expect(err).To(BeNil())
				Expect(count).To(Equal(1))
			})
		})
	})

	Describe(".isValidOpReturn", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var tm *TicketManager

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			tm = NewTicketManager(cfg, client, interrupt)
			err := tm.openDB()
			Expect(err).To(BeNil())
		})

		Describe(".isValidOpReturn", func() {
			It("should return error when .ASM value does not include op_return string", func() {
				out := btcjson.Vout{ScriptPubKey: btcjson.ScriptPubKeyResult{Asm: "something"}}
				_, err := isValidOpReturn(out)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("missing op_return keyword"))
			})

			It("should return error when .ASM value has one part", func() {
				out := btcjson.Vout{ScriptPubKey: btcjson.ScriptPubKeyResult{Asm: "op_return"}}
				_, err := isValidOpReturn(out)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid Asm format"))
			})

			It("should return error when .ASM op_return value is not hex encoded", func() {
				out := btcjson.Vout{ScriptPubKey: btcjson.ScriptPubKeyResult{Asm: "op_return &*^HH"}}
				_, err := isValidOpReturn(out)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to decoded from hex"))
			})

			It("should return error when op_return hex-encoded value is not 22 bytes in size", func() {
				val := hex.EncodeToString([]byte{2, 5, 6})
				out := btcjson.Vout{ScriptPubKey: btcjson.ScriptPubKeyResult{Asm: "op_return " + val}}
				_, err := isValidOpReturn(out)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("value must be 22 byte in size"))
			})

			It("should return error when op_return hex-encoded value has invalid prefix", func() {
				dec, _ := crypto.DecodeAddrOnly("eDHeYqxu3B93QW62fxuxfCFE66jeAeVTkT")
				val := hex.EncodeToString(append([]byte{2, 5}, dec[:]...))
				out := btcjson.Vout{ScriptPubKey: btcjson.ScriptPubKeyResult{Asm: "op_return " + val}}
				_, err := isValidOpReturn(out)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unexpected prefix"))
			})

			It("should return error when value is less than minimum burn amount", func() {
				dec, _ := crypto.DecodeAddrOnly("eDHeYqxu3B93QW62fxuxfCFE66jeAeVTkT")
				val := hex.EncodeToString(append(ticketOpReturnPrefix, dec[:]...))
				out := btcjson.Vout{ScriptPubKey: btcjson.ScriptPubKeyResult{Asm: "op_return " + val}, Value: 0}
				_, err := isValidOpReturn(out)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("output value is insufficient"))
			})

			It("should return no error when valid", func() {
				dec, _ := crypto.DecodeAddrOnly("eDHeYqxu3B93QW62fxuxfCFE66jeAeVTkT")
				val := hex.EncodeToString(append(ticketOpReturnPrefix, dec[:]...))
				out := btcjson.Vout{ScriptPubKey: btcjson.ScriptPubKeyResult{Asm: "op_return " + val}, Value: 100}
				data, err := isValidOpReturn(out)
				Expect(err).To(BeNil())
				Expect(data).To(Equal(dec[:]))
			})
		})
	})
})
