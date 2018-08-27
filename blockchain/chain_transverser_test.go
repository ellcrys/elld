package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var ChainTransverserTest = func() bool {
	return Describe("ChainTransverser", func() {

		Describe(".Query", func() {

			var trv *ChainTransverser

			BeforeEach(func() {
				trv = bc.NewChainTransverser()
			})

			It("should return error if query function returned an error", func() {
				chain := NewChain("abc", db, cfg, log)
				err := trv.Start(chain).Query(func(c core.Chainer) (bool, error) {
					return false, fmt.Errorf("something bad")
				})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("something bad"))
			})

			It("should return nil when true is returned", func() {
				chain := NewChain("abc", db, cfg, log)
				err := trv.Start(chain).Query(func(c core.Chainer) (bool, error) {
					return true, nil
				})
				Expect(err).To(BeNil())
			})

			When("object is stored 3 generations chain", func() {

				var chain, parent, grandParent *Chain

				BeforeEach(func() {
					chain = NewChain("abc", db, cfg, log)
					parent = NewChain("parent_abc", db, cfg, log)
					grandParent = NewChain("grand_parent_abc", db, cfg, log)

					err := bc.saveChain(chain, parent.GetID(), 2)
					Expect(err).To(BeNil())
					err = bc.saveChain(parent, grandParent.GetID(), 2)
					Expect(err).To(BeNil())
					err = bc.saveChain(grandParent, "", 0)
					Expect(err).To(BeNil())

					Expect(bc.chains).To(HaveLen(4))
				})

				It("should return object store in parent chain", func() {

					key := elldb.MakeKey(nil, []byte(grandParent.id), []byte("stuff"))
					db.Put([]*elldb.KVObject{
						elldb.NewKVObject(key, []byte("123")),
					})

					var result []*elldb.KVObject
					err = trv.Start(chain).Query(func(c core.Chainer) (bool, error) {
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
}
