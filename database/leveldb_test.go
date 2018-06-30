package database

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database", func() {

	var testCfgDir = "/Users/ncodes/.ellcrys_test"
	var db DB

	BeforeEach(func() {
		err := os.Mkdir(testCfgDir, 0700)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = NewLevelDB(testCfgDir)
		err := db.Open("")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		err := os.RemoveAll(testCfgDir)
		Expect(err).To(BeNil())
	})

	Describe(".Open", func() {
		It("should return error if unable to open database", func() {
			db = NewLevelDB("/*^&^")
			err := db.Open("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to create database. mkdir /*^&^: permission denied"))
		})
	})

	Describe(".WriteBatch", func() {
		It("should successfully write several objects", func() {
			err := db.WriteBatch([]*KVObject{
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
			err := db.WriteBatch(objs)
			Expect(err).To(BeNil())
			results := db.GetByPrefix([]byte("obj"))
			Expect(results).To(HaveLen(2))
			Expect(results[0]).To(Equal(objs[0]))
			Expect(results[1]).To(Equal(objs[1]))
		})
	})

	Describe(".DeleteByPrefix", func() {
		It("should successfully delete objects", func() {
			err := db.WriteBatch([]*KVObject{
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

})
