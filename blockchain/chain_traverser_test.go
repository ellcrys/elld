package blockchain

import (
	"fmt"
	"os"

	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ChainTraverser", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var log = logger.NewLogrusNoOp()

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(log)
		err = db.Open(cfg.NetDataDir())
		Expect(err).To(BeNil())

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
		bc.SetNodeKey(crypto.NewKeyFromIntSeed(1234))
	})

	BeforeEach(func() {
		genesisBlock, err := LoadBlockFromFile("genesis-test.json")
		Expect(err).To(BeNil())
		bc.SetGenesisBlock(genesisBlock)
		err = bc.Up()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe(".Query", func() {

		var trv *ChainTraverser
		var chain *Chain

		BeforeEach(func() {
			trv = bc.NewChainTraverser()
			chain = NewChain("chain_x", db, cfg, log)
		})

		It("should return error if query function returned an error", func() {
			err := trv.Start(chain).Query(func(c types.Chainer) (bool, error) {
				return false, fmt.Errorf("something bad")
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("something bad"))
		})

		It("should return nil when true is returned", func() {
			err := trv.Start(chain).Query(func(c types.Chainer) (bool, error) {
				return true, nil
			})
			Expect(err).To(BeNil())
		})

		When("object is stored/associated to a chain's grand parent", func() {

			var parent, grandParent *Chain

			// Target shape:
			// [1]-[2]..[n] - grand parent chain (grand_parent_x)
			//      |___[n] - parent chain (parent_x)
			//      |___[n] - chain (chain_x)
			BeforeEach(func() {
				parent = NewChain("parent_x", db, cfg, log)
				grandParent = NewChain("grand_parent_x", db, cfg, log)

				err := bc.saveChain(chain, parent.GetID(), 2)
				Expect(err).To(BeNil())
				err = bc.saveChain(parent, grandParent.GetID(), 2)
				Expect(err).To(BeNil())
				err = bc.saveChain(grandParent, "", 0)
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(4))
			})

			Context("when starting chain has 2 ancestors (parent and grand parent)", func() {
				var ancestors []string

				BeforeEach(func() {
					ancestors = []string{}
					err = trv.Start(chain).Query(func(c types.Chainer) (bool, error) {
						ancestors = append(ancestors, c.GetID().String())
						return false, nil
					})
					Expect(err).To(BeNil())
				})

				It("should go up the chain ancestry returning 3 chains (itself and the ancestors)", func() {
					Expect(ancestors).To(HaveLen(3))
				})

				Specify("the returned ancestor order must start from the start chain and up to the last ancestor", func() {
					Expect(ancestors[0]).To(Equal(chain.GetID().String()))
					Expect(ancestors[1]).To(Equal(parent.GetID().String()))
					Expect(ancestors[2]).To(Equal(grandParent.GetID().String()))
				})
			})

			It("should find the object stored at the ancestor (grand parent)", func() {

				key := elldb.MakeKey(nil, []byte(grandParent.id), []byte("stuff"))
				db.Put([]*elldb.KVObject{
					elldb.NewKVObject(key, []byte("123")),
				})

				var result []*elldb.KVObject
				err = trv.Start(chain).Query(func(c types.Chainer) (bool, error) {
					key := elldb.MakeKey(nil, []byte(c.GetID()), []byte("stuff"))
					if result = db.GetByPrefix(key); len(result) > 0 {
						return true, nil
					}
					return false, nil
				})

				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(trv.chain.GetID()).To(Equal(grandParent.GetID()))
			})
		})
	})

})
