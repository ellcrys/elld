package burner

import (
	"os"

	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/ltcd/chaincfg/chainhash"
	"github.com/ellcrys/ltcd/wire"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util/logger"
	"github.com/gojuno/minimock"
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
	// var netVersion = "test"
	var db elldb.DB

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
		log = logger.NewLogrusNoOp()
		bus = &emitter.Emitter{}
		db = elldb.NewDB(cfg.NetDataDir())
		Expect(db.Open("")).To(BeNil())
		minStartHeight = 0
	})

	AfterEach(func() {
		db.Close()
		err := os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe(".getLatestLocalBlock", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var bi *BlockIndexer

		Context("with 3 blocks indexed", func() {
			BeforeEach(func() {
				client = NewRPCClientMock(mc)
				bi = NewBlockIndexer(cfg, log, db, bus, client, interrupt)
				k1 := MakeKeyIndexerBlock(1)
				o1 := elldb.NewKVObject(k1, util.ObjectToBytes(LocalBlockHeader{Number: 1}))
				k2 := MakeKeyIndexerBlock(10)
				o2 := elldb.NewKVObject(k2, util.ObjectToBytes(LocalBlockHeader{Number: 10}))
				k3 := MakeKeyIndexerBlock(2)
				o3 := elldb.NewKVObject(k3, util.ObjectToBytes(LocalBlockHeader{Number: 2}))
				err := db.Put([]*elldb.KVObject{o1, o2, o3})
				Expect(err).To(BeNil())
			})

			It("should return the block with the highest height", func() {
				lb, err := bi.getLatestLocalBlock()
				Expect(err).To(BeNil())
				Expect(lb.Number).To(Equal(int64(10)))
			})
		})

		Context("when no block header is found", func() {

			BeforeEach(func() {
				client = NewRPCClientMock(mc)
				bi = NewBlockIndexer(cfg, log, db, bus, client, interrupt)
			})

			It("should return errHeaderNotFound ", func() {
				_, err := bi.getLatestLocalBlock()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(errHeaderNotFound))
			})
		})

	})

	Describe(".getLocalBlock", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var bi *BlockIndexer

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			bi = NewBlockIndexer(cfg, log, db, bus, client, interrupt)
		})

		It("should return err if block at the given height is not found", func() {
			_, err := bi.getLocalBlock(10)
			Expect(err).To(Equal(errHeaderNotFound))
		})

		Context("with block at height 1", func() {
			BeforeEach(func() {
				key := MakeKeyIndexerBlock(1)
				b := LocalBlockHeader{Number: 1}
				kv := elldb.NewKVObject(key, util.ObjectToBytes(b))
				err := db.Put([]*elldb.KVObject{kv})
				Expect(err).To(BeNil())
			})

			It("should get block at height 1", func() {
				header, err := bi.getLocalBlock(1)
				Expect(err).To(BeNil())
				Expect(header.Number).To(Equal(int64(1)))
			})
		})
	})

	Describe(".getStartHeight", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var bi *BlockIndexer

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			bi = NewBlockIndexer(cfg, log, db, bus, client, interrupt)
		})

		AfterEach(func() {
			mc.Finish()
		})

		Context("with no local block", func() {
			BeforeEach(func() {
				client.GetBestBlockMock.Return(nil, 0, nil)
			})

			It("should return zero (0)", func() {
				h, err := bi.getStartHeight()
				Expect(err).To(BeNil())
				Expect(h).To(Equal(int64(0)))
			})
		})

		Context("when a minimum start height is 10", func() {
			BeforeEach(func() {
				client.GetBestBlockMock.Return(nil, 0, nil)
				minStartHeight = 10
			})

			It("should return minimum start height if local height is lesser", func() {
				h, err := bi.getStartHeight()
				Expect(err).To(BeNil())
				Expect(h).To(Equal(int64(10)))
			})
		})

		Context("with 3 blocks of height 1,2,3 indexed", func() {
			BeforeEach(func() {
				bi = NewBlockIndexer(cfg, log, db, bus, client, interrupt)
				k1 := MakeKeyIndexerBlock(1)
				o1 := elldb.NewKVObject(k1, util.ObjectToBytes(LocalBlockHeader{Number: 1}))
				k2 := MakeKeyIndexerBlock(3)
				o2 := elldb.NewKVObject(k2, util.ObjectToBytes(LocalBlockHeader{Number: 3}))
				k3 := MakeKeyIndexerBlock(2)
				o3 := elldb.NewKVObject(k3, util.ObjectToBytes(LocalBlockHeader{Number: 2}))
				err := db.Put([]*elldb.KVObject{o1, o2, o3})
				Expect(err).To(BeNil())
			})

			Context("when upstream best block is greater than local height", func() {
				BeforeEach(func() {
					client.GetBestBlockMock.Return(nil, 4, nil)
				})

				It("should return the the height of the highest local block", func() {
					h, err := bi.getStartHeight()
					Expect(err).To(BeNil())
					Expect(h).To(Equal(int64(3)))
				})
			})

			Context("when upstream best block is less than local block height", func() {
				BeforeEach(func() {
					client.GetBestBlockMock.Return(nil, 2, nil)
				})

				It("should return the the height of the upstream block - 1", func() {
					h, err := bi.getStartHeight()
					Expect(err).To(BeNil())
					Expect(h).To(Equal(int64(1)))
				})
			})
		})
	})

	Describe(".detectReorg", func() {

		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var bi *BlockIndexer

		BeforeEach(func() {
			client = NewRPCClientMock(mc)
			bi = NewBlockIndexer(cfg, log, db, bus, client, interrupt)
		})

		It("should return false, if new block height is 1", func() {
			newBlock := &wire.BlockHeader{}
			reorg, err := bi.detectReorg(newBlock, 1)
			Expect(err).To(BeNil())
			Expect(reorg).To(BeFalse())
		})

		Context("when the new block parent's block height has no equivalent on the local index", func() {

			BeforeEach(func() {
				key := MakeKeyIndexerBlock(2)
				b := LocalBlockHeader{Number: 3, Hash: []byte("hash")}
				kv := elldb.NewKVObject(key, util.ObjectToBytes(b))
				err := db.Put([]*elldb.KVObject{kv})
				Expect(err).To(BeNil())
			})

			It("should return expected error", func() {
				hashStr := "c694ce78fe9555168d66c503f0709a1ff715e196a50a4931a8afb1ee388d1370"
				hash, _ := chainhash.NewHashFromStr(hashStr)
				newBlock := &wire.BlockHeader{PrevBlock: *hash}
				_, err := bi.detectReorg(newBlock, 4)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("local block index is missing a common block (3)"))
			})
		})

		Context("when new block parent hash does not match local equivalent", func() {

			BeforeEach(func() {
				key := MakeKeyIndexerBlock(3)
				b := LocalBlockHeader{Number: 3, Hash: []byte("hash")}
				kv := elldb.NewKVObject(key, util.ObjectToBytes(b))
				err := db.Put([]*elldb.KVObject{kv})
				Expect(err).To(BeNil())
			})

			It("should return", func() {
				hashStr := "c694ce78fe9555168d66c503f0709a1ff715e196a50a4931a8afb1ee388d1370"
				hash, _ := chainhash.NewHashFromStr(hashStr)
				newBlock := &wire.BlockHeader{PrevBlock: *hash}
				reorg, err := bi.detectReorg(newBlock, 4)
				Expect(err).To(BeNil())
				Expect(reorg).To(BeTrue())
			})
		})

		Context("when new block parent hash match local equivalent", func() {
			hashStr := "c694ce78fe9555168d66c503f0709a1ff715e196a50a4931a8afb1ee388d1370"

			BeforeEach(func() {
				key := MakeKeyIndexerBlock(3)
				b := LocalBlockHeader{Number: 3, Hash: []byte(hashStr)}
				kv := elldb.NewKVObject(key, util.ObjectToBytes(b))
				err := db.Put([]*elldb.KVObject{kv})
				Expect(err).To(BeNil())
			})

			It("should return", func() {
				hash, _ := chainhash.NewHashFromStr(hashStr)
				newBlock := &wire.BlockHeader{PrevBlock: *hash}
				reorg, err := bi.detectReorg(newBlock, 4)
				Expect(err).To(BeNil())
				Expect(reorg).To(BeTrue())
			})
		})
	})

	Describe(".deleteBlocksFrom", func() {
		var mc = minimock.NewController(GinkgoT())
		var client *RPCClientMock
		var bi *BlockIndexer

		Context("with block 1, 2, 3 indexed", func() {
			BeforeEach(func() {
				client = NewRPCClientMock(mc)
				bi = NewBlockIndexer(cfg, log, db, bus, client, interrupt)

				k1 := MakeKeyIndexerBlock(1)
				o := elldb.NewKVObject(k1, []byte{})
				k2 := MakeKeyIndexerBlock(2)
				o2 := elldb.NewKVObject(k2, []byte{})
				k3 := MakeKeyIndexerBlock(3)
				o3 := elldb.NewKVObject(k3, []byte{})

				err := db.Put([]*elldb.KVObject{o, o2, o3})
				Expect(err).To(BeNil())
			})

			When("block 2 is passed", func() {
				It("delete block 2,3", func() {
					err := bi.deleteBlocksFrom(2)
					Expect(err).To(BeNil())
					res := db.GetByPrefix(MakeQueryKeyIndexerBlock())
					Expect(res).To(HaveLen(1))
					Expect(res[0].Key).To(Equal(util.EncodeNumber(1)))
				})
			})
		})
	})
})
