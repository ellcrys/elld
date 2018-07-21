package database

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database", func() {

	var testCfgDir = "/Users/ncodes/.ellcrys_test"
	var db DB
	var err error

	BeforeEach(func() {
		err = os.Mkdir(testCfgDir, 0700)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = NewLevelDB(testCfgDir)
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
			db = NewLevelDB("/*^&^")
			err = db.Open("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to create database. mkdir /*^&^: permission denied"))
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
			results := db.GetByPrefix([]byte("obj"))
			Expect(results).To(HaveLen(2))
			Expect(results[0]).To(Equal(objs[0]))
			Expect(results[1]).To(Equal(objs[1]))
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

			err = db.DeleteByPrefix([]byte("object"))
			Expect(err).To(BeNil())

			objs := db.GetByPrefix([]byte("obj"))
			Expect(objs).To(HaveLen(0))

			objs = db.GetByPrefix([]byte("an"))
			Expect(objs).To(HaveLen(1))
		})
	})

	Describe(".GetFirstOrLast", func() {

		var key, val, key2, val2 []byte

		BeforeEach(func() {
			key, val = []byte("age"), []byte("20")
			key2, val2 = []byte("age"), []byte("20")
			err = db.Put([]*KVObject{NewKVObject(key, val, "namespace.1")})
			Expect(err).To(BeNil())
			err = db.Put([]*KVObject{NewKVObject(key2, val2, "namespace.2")})
			Expect(err).To(BeNil())
		})

		It("should get the first item when first arg is set to true", func() {
			obj := db.GetFirstOrLast([]byte("namespace"), true)
			Expect(obj.Key).To(Equal(key))
			Expect(obj.Value).To(Equal(val))
		})

		It("should get the last item when first arg is set to false", func() {
			obj := db.GetFirstOrLast([]byte("namespace"), false)
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

				result := db.GetByPrefix(key)
				Expect(result).NotTo(BeEmpty())
			})

			It("should not put object if rollback was called", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val)})
				Expect(err).To(BeNil())
				dbTx.Rollback()

				result := db.GetByPrefix(key)
				Expect(result).To(BeEmpty())
			})
		})

		Describe(".GetByPrefix", func() {
			It("should get object by prefix", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val, "namespace")})
				Expect(err).To(BeNil())

				objs := dbTx.GetByPrefix([]byte("namespace"))
				Expect(objs).To(HaveLen(1))
				dbTx.Commit()

				objs = db.GetByPrefix([]byte("namespace"))
				Expect(objs).To(HaveLen(1))
			})

			It("should not get object by prefix if rollback was called", func() {
				key := []byte("age")
				val := []byte("20")
				err = dbTx.Put([]*KVObject{NewKVObject(key, val, "namespace")})
				Expect(err).To(BeNil())
				dbTx.Rollback()

				objs := dbTx.GetByPrefix([]byte("namespace"))
				Expect(objs).To(BeEmpty())

				objs = db.GetByPrefix([]byte("namespace"))
				Expect(objs).To(BeEmpty())
			})
		})
	})

	Describe(".Iterate", func() {

		BeforeEach(func() {
			err = db.Put([]*KVObject{NewKVObject([]byte("some_key"), []byte("a"), "namespace.1")})
			Expect(err).To(BeNil())
			err = db.Put([]*KVObject{NewKVObject([]byte("some_key"), []byte("b"), "namespace.2")})
			Expect(err).To(BeNil())
			err = db.Put([]*KVObject{NewKVObject([]byte("some_key"), []byte("c"), "namespace.3")})
			Expect(err).To(BeNil())
		})

		It("should find items in this order namespace.1, namespace.2, namespace.3", func() {
			var itemsKey []string
			db.Iterate([]byte("namespace"), true, func(kv *KVObject) bool {
				itemsKey = append(itemsKey, kv.Prefixes[0])
				return false
			})
			Expect(itemsKey).To(Equal([]string{"namespace.1", "namespace.2", "namespace.3"}))
		})

		It("should find items in this order namespace.3, namespace.2, namespace.1", func() {
			var itemsKey []string
			db.Iterate([]byte("namespace"), false, func(kv *KVObject) bool {
				itemsKey = append(itemsKey, kv.Prefixes[0])
				return false
			})
			Expect(itemsKey).To(Equal([]string{"namespace.3", "namespace.2", "namespace.1"}))
		})

		It("should find item namespace.2 only", func() {
			var itemsKey []string
			db.Iterate([]byte("namespace"), true, func(kv *KVObject) bool {
				if kv.Prefixes[0] == "namespace.2" {
					itemsKey = append(itemsKey, kv.Prefixes[0])
					return true
				}
				return false
			})
			Expect(itemsKey).To(Equal([]string{"namespace.2"}))
		})
	})
})
