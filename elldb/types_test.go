package elldb

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

	Describe(".MakePrefix", func() {
		It("should return 'prefixA~prefixB'", func() {
			actual := MakePrefix([]string{"prefixA", "prefixB"})
			Expect(string(actual)).To(Equal("prefixA~prefixB"))
		})

		It("should return 'prefixA'", func() {
			actual := MakePrefix([]string{"prefixA"})
			Expect(string(actual)).To(Equal("prefixA"))
		})

		It("should return empty string when prefixes are not provided", func() {
			actual := MakePrefix([]string{})
			Expect(string(actual)).To(Equal(""))
		})
	})

	Describe(".MakeKey", func() {
		It("should return 'prefixA~prefixB:age' when key and prefixes are provided", func() {
			actual := MakeKey([]byte("age"), []string{"prefixA", "prefixB"})
			Expect(string(actual)).To(Equal("prefixA~prefixB:age"))
		})

		It("should return only concatenated prefixes 'prefixA~prefixB' when key is not provided", func() {
			actual := MakeKey(nil, []string{"prefixA", "prefixB"})
			Expect(string(actual)).To(Equal("prefixA~prefixB"))
		})

		It("should return only key when prefixes are not provided", func() {
			actual := MakeKey([]byte("age"), nil)
			Expect(string(actual)).To(Equal("age"))
		})

		It("should return empty string when key and prefixes are not provided", func() {
			actual := MakeKey(nil, nil)
			Expect(string(actual)).To(Equal(""))
		})
	})

	Describe("NewKVObject.GetKey vs .MakeKey", func() {
		It("should return same result when key and prefixes are provided", func() {
			kv := NewKVObject([]byte("age"), []byte("value"), "prefixA", "prefixB")
			key := MakeKey([]byte("age"), []string{"prefixA", "prefixB"})
			Expect(kv.GetKey()).To(Equal(key))
		})

		It("should return same result when key is not provided", func() {
			kv := NewKVObject(nil, []byte("value"), "prefixA", "prefixB")
			key := MakeKey(nil, []string{"prefixA", "prefixB"})
			Expect(kv.GetKey()).To(Equal(key))
		})

		It("should return same result when prefixes are not provided", func() {
			kv := NewKVObject([]byte("age"), nil)
			key := MakeKey([]byte("age"), nil)
			Expect(kv.GetKey()).To(Equal(key))
		})

		It("should return empty string when key and prefixes are not provided", func() {
			kv := NewKVObject(nil, nil)
			key := MakeKey(nil, nil)
			Expect(kv.GetKey()).To(Equal(key))
		})
	})
})
