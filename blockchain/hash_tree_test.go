package blockchain

import (
	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockchain", func() {

	var err error
	var store types.Store
	var db database.DB
	var chainID = "chain1"
	var hashTree *HashTree

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = database.NewLevelDB(cfg.ConfigDir())
		err = db.Open("")
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store, err = leveldb.New(db)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		hashTree = NewHashTree(chainID, store)
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".Upsert", func() {
		It("should successfully add object without error", func() {
			key := []byte("my_key")
			err = hashTree.Upsert(key, []byte("value"))
			Expect(err).To(BeNil())
		})
	})

	Describe(".Find", func() {

		It("should successfully find object without error", func() {
			key := []byte("my_key")
			value := ByteVal([]byte("value"))
			err = hashTree.Upsert(key, []byte(value))
			Expect(err).To(BeNil())

			obj, err := hashTree.Find(key)
			Expect(err).To(BeNil())
			Expect(obj).ToNot(BeNil())
			Expect(obj).To(Equal(value))
		})

		It("should return error when key is not found", func() {
			key := []byte("my_key")
			_, err := hashTree.Find(key)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("not found"))
		})
	})

	Describe(".Root", func() {

		It("should successfully return expected root", func() {
			key := []byte("my_key")
			value := ByteVal([]byte("value"))
			err = hashTree.Upsert(key, []byte(value))
			Expect(err).To(BeNil())

			root, err := hashTree.Root()
			expected := "0xcf7929487dad938befb80fd788e48952a6500d07c9e442c6e3534f065461e5abc5380dd8ad7a81bffe5408b9355f2d227d69ae2702adec9f0ce05e9e48b84851"
			Expect(err).To(BeNil())
			Expect(util.ToHex(root)).To(Equal(expected))
		})
	})

	Describe(".NewHashTree", func() {

		var root []byte

		BeforeEach(func() {
			err = hashTree.Upsert([]byte("k1"), []byte("a"))
			Expect(err).To(BeNil())
			err = hashTree.Upsert([]byte("k2"), []byte("b"))
			Expect(err).To(BeNil())
			err = hashTree.Upsert([]byte("k6"), []byte("cs"))
			Expect(err).To(BeNil())
			root, err = hashTree.Root()
			Expect(err).To(BeNil())
		})

		Context("with chain ID of existing tree", func() {

			It("should load existing tree with matching root", func() {
				existingTree := NewHashTree(chainID, store)
				existingTreeRoot, err := existingTree.Root()
				Expect(err).To(BeNil())
				Expect(existingTreeRoot).To(Equal(root))
			})
		})
	})
})
