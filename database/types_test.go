package database

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types", func() {

	Describe("NewKVObject.GetKey", func() {
		It("should create key", func() {
			kv := NewKVObject([]byte("age"), []byte("20"), "prefix")
			key := kv.GetKey()
			Expect(key).ToNot(BeEmpty())
			Expect(key).To(Equal([]byte{112, 114, 101, 102, 105, 120, 58, 97, 103, 101}))
		})
	})

	Describe(".FromKeyValue", func() {
		When("key has no prefix", func() {
			It("should return object that contains key=age, value=20 and prefixes=nil", func() {
				o := FromKeyValue([]byte("age"), []byte("20"))
				Expect(o.Key).To(Equal([]byte("age")))
				Expect(o.Value).To(Equal([]byte("20")))
				Expect(o.Prefixes).To(BeEmpty())
				Expect(o.GetKey()).To(Equal([]byte("age")))
			})
		})

		When("key has prefix", func() {
			It("should return object that contains key=age, value=20, ", func() {
				o := FromKeyValue([]byte("prefixA~prefixB:age"), []byte("20"))
				Expect(o.Key).To(Equal([]byte("age")))
				Expect(o.Value).To(Equal([]byte("20")))
				Expect(o.GetKey()).To(Equal([]byte("prefixA~prefixB:age")))
				Expect(o.Prefixes).To(Equal([]string{"prefixA", "prefixB"}))
			})
		})
	})
})
