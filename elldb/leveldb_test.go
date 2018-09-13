package elldb

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database", func() {

	var testCfgDir string
	var db DB
	var err error

	BeforeEach(func() {
		home, _ := homedir.Dir()
		testCfgDir = filepath.Join(home, ".ellcrys_test")
		err = os.Mkdir(testCfgDir, 0700)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = NewDB(testCfgDir)
		err = db.Open("")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		err = os.RemoveAll(testCfgDir)
		Expect(err).To(BeNil())
	})

	Describe(".Open", func() {
		It("should return error if unable to open database", func() {
			db = NewDB(testCfgDir)
			err = db.Open("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to create database. resource temporarily unavailable"))
		})
	})

	Describe(".WriteBatch", func() {
		It("should successfully write several objects", func() {
			err = db.Put([]*KVObject{
				NewKVObject([]byte("object_1"), []byte("value1")),
				NewKVObject([]byte("object_2"), []byte("value2")),
			})
			Expect(err).To(BeNil())
		})
	})

	Describe(".Get", func() {
		It("should successfully get objects", func() {
			objs := []*KVObject{
				NewKVObject([]byte("object_1"), []byte("value1")),
				NewKVObject([]byte("object_2"), []byte("value2")),
			}
			err = db.Put(objs)
			Expect(err).To(BeNil())
			results := db.GetByPrefix(MakeKey([]byte("obj")))
			Expect(results).To(HaveLen(2))
			Expect(results[0].Equal(objs[0])).To(BeTrue())
			Expect(results[1].Equal(objs[1])).To(BeTrue())
		})
	})

	Describe(".DeleteByPrefix", func() {
		It("should successfully delete objects", func() {
			err := db.Put([]*KVObject{
				&KVObject{Key: []byte("object_1"), Value: []byte("value1")},
				&KVObject{Key: []byte("object_2"), Value: []byte("value2")},
				&KVObject{Key: []byte("another_object_3"), Value: []byte("value3")},
			})
			Expect(err).To(BeNil())

			err = db.DeleteByPrefix(MakeKey([]byte("object")))
			Expect(err).To(BeNil())

			objs := db.GetByPrefix(MakeKey([]byte("obj")))
			Expect(objs).To(HaveLen(0))

			objs = db.GetByPrefix(MakeKey([]byte("an")))
			Expect(objs).To(HaveLen(1))
		})
	})

	Describe(".Truncate", func() {
		It("should successfully get objects", func() {
			objs := []*KVObject{
				NewKVObject([]byte("object_1"), []byte("value1")),
				NewKVObject([]byte("object_2"), []byte("value2")),
			}
			err = db.Put(objs)
			Expect(err).To(BeNil())

			err = db.Truncate()
			Expect(err).To(BeNil())

			results := db.GetByPrefix(nil)
			Expect(results).To(HaveLen(0))
		})
	})

	Describe(".GetFirstOrLast", func() {

		var key, val, key2, val2 []byte

		BeforeEach(func() {
			key, val = []byte("age"), []byte("20")
			key2, val2 = []byte("age"), []byte("20")
			err = db.Put([]*KVObject{
				NewKVObject(key, val, []byte("namespace.1")),
				NewKVObject(key2, val2, []byte("namespace.2")),
			})
			Expect(err).To(BeNil())
		})

		It("should get the first item when first arg is set to true", func() {
			obj := db.GetFirstOrLast(MakeKey(nil, []byte("namespace")), true)
			Expect(obj.Key).To(Equal(key))
			Expect(obj.Value).To(Equal(val))
		})

		It("should get the last item when first arg is set to false", func() {
			obj := db.GetFirstOrLast(MakeKey(nil, []byte("namespace")), false)
			Expect(obj.Key).To(Equal(key2))
			Expect(obj.Value).To(Equal(val2))
		})
	})

	Context("using a transaction", func() {

		var err error
		var dbTx Tx

		BeforeEach(func() {
			dbTx, err = db.NewTx()
			Expect(err).To(BeNil())
		})

		Describe(".Put", func() {
			It("should not put object if commit was not called", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val)})
				Expect(err).To(BeNil())

				result := db.GetByPrefix(key)
				Expect(result).To(BeEmpty())
			})

			It("should put object if commit was called", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val)})
				Expect(err).To(BeNil())
				dbTx.Commit()

				result := db.GetByPrefix(MakeKey(key))
				Expect(result).NotTo(BeEmpty())
			})

			It("should not put object if rollback was called", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val)})
				Expect(err).To(BeNil())
				dbTx.Rollback()

				result := db.GetByPrefix(MakeKey(key))
				Expect(result).To(BeEmpty())
			})
		})

		Describe(".GetByPrefix", func() {
			It("should get object by prefix", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val, []byte("namespace"))})
				Expect(err).To(BeNil())

				objs := dbTx.GetByPrefix(MakeKey(nil, []byte("namespace")))
				Expect(objs).To(HaveLen(1))
				dbTx.Commit()

				objs = db.GetByPrefix(MakeKey(nil, []byte("namespace")))
				Expect(objs).To(HaveLen(1))
			})

			It("should not get object by prefix if rollback was called", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val, []byte("namespace"))})
				Expect(err).To(BeNil())
				dbTx.Rollback()

				objs := dbTx.GetByPrefix(MakeKey(nil, []byte("namespace")))
				Expect(objs).To(BeEmpty())

				objs = db.GetByPrefix(MakeKey(nil, []byte("namespace")))
				Expect(objs).To(BeEmpty())
			})
		})

		Describe(".DeleteByPrefix", func() {
			It("should successfully delete objects", func() {
				err := dbTx.Put([]*KVObject{
					&KVObject{Key: []byte("object_1"), Value: []byte("value1")},
					&KVObject{Key: []byte("object_2"), Value: []byte("value2")},
					&KVObject{Key: []byte("another_object_3"), Value: []byte("value3")},
				})
				Expect(err).To(BeNil())

				err = dbTx.DeleteByPrefix(MakeKey([]byte("object")))
				Expect(err).To(BeNil())

				objs := dbTx.GetByPrefix(MakeKey([]byte("obj")))
				Expect(objs).To(HaveLen(0))

				objs = dbTx.GetByPrefix(MakeKey([]byte("an")))
				Expect(objs).To(HaveLen(1))
			})

			It("should not successfully delete if rollback is called", func() {
				err := dbTx.Put([]*KVObject{
					&KVObject{Key: []byte("object_1"), Value: []byte("value1")},
					&KVObject{Key: []byte("object_2"), Value: []byte("value2")},
				})
				Expect(err).To(BeNil())
				dbTx.Commit()

				dbTx, _ = db.NewTx()
				err = dbTx.DeleteByPrefix(MakeKey([]byte("object")))
				Expect(err).To(BeNil())
				dbTx.Rollback()

				objs := db.GetByPrefix(MakeKey([]byte("obj")))
				Expect(objs).To(HaveLen(2))
			})
		})
	})

	Describe(".Iterate", func() {

		BeforeEach(func() {
			err = db.Put([]*KVObject{
				NewKVObject([]byte("some_key"), []byte("a"), []byte("namespace.1")),
				NewKVObject([]byte("some_key"), []byte("b"), []byte("namespace.2")),
				NewKVObject([]byte("some_key"), []byte("c"), []byte("namespace.3")),
			})
			Expect(err).To(BeNil())
		})

		It("should find items in this order namespace.1, namespace.2, namespace.3", func() {
			var itemsKey [][]byte
			db.Iterate([]byte("namespace"), true, func(kv *KVObject) bool {
				itemsKey = append(itemsKey, kv.Prefix)
				return false
			})
			Expect(itemsKey).To(Equal([][]byte{[]byte("namespace.1"), []byte("namespace.2"), []byte("namespace.3")}))
		})

		It("should find items in this order namespace.3, namespace.2, namespace.1", func() {
			var itemsKey [][]byte
			db.Iterate([]byte("namespace"), false, func(kv *KVObject) bool {
				itemsKey = append(itemsKey, kv.Prefix)
				return false
			})
			Expect(itemsKey).To(Equal([][]byte{[]byte("namespace.3"), []byte("namespace.2"), []byte("namespace.1")}))
		})

		It("should find item namespace.2 only", func() {
			var itemsKey [][]byte
			db.Iterate([]byte("namespace"), true, func(kv *KVObject) bool {
				if bytes.Equal(kv.Prefix, []byte("namespace.2")) {
					itemsKey = append(itemsKey, kv.Prefix)
					return true
				}
				return false
			})
			Expect(itemsKey).To(Not(BeEmpty()))
			Expect(itemsKey[0]).To(Equal([]byte("namespace.2")))
		})
	})
})
