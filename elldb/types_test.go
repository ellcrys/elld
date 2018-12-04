package elldb

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types", func() {
	Describe("NewKVObject.GetKey", func() {
		It("should create key", func() {
			kv := NewKVObject([]byte("age"), []byte("20"), []byte("prefix"))
			key := kv.GetKey()
			Expect(key).ToNot(BeEmpty())
			Expect(key).To(Equal([]byte("prefix@@age")))
		})
	})

	Describe(".FromKeyValue", func() {
		When("key has no KeyPrefixSeparator", func() {
			It("should return object that contains key=age, prefix= value=20 and GetKey=age@@", func() {
				o := FromKeyValue([]byte("age"), []byte("20"))
				Expect(o.Prefix).To(BeEmpty())
				Expect(o.Key).To(Equal([]byte("age")))
				Expect(o.Value).To(Equal([]byte("20")))
				Expect(o.GetKey()).To(Equal([]byte("age")))
			})
		})

		When("key has KeyPrefixSeparator", func() {
			It("should return object that contains key=age, value=20, prefix=prefixA", func() {
				o := FromKeyValue([]byte("prefixA@@age"), []byte("20"))
				Expect(o.Key).To(Equal([]byte("age")))
				Expect(o.Value).To(Equal([]byte("20")))
				Expect(o.GetKey()).To(Equal([]byte("prefixA@@age")))
				Expect(o.Prefix).To(Equal([]byte("prefixA")))
			})
		})
	})

	Describe(".MakePrefix", func() {
		It("should return 'prefixA:prefixB'", func() {
			actual := MakePrefix([]byte("prefixA"), []byte("prefixB"))
			Expect(string(actual)).To(Equal("prefixA:prefixB"))
		})

		It("should return 'prefixA'", func() {
			actual := MakePrefix([]byte("prefixA"))
			Expect(string(actual)).To(Equal("prefixA"))
		})

		It("should return empty string when prefixes are not provided", func() {
			actual := MakePrefix()
			Expect(string(actual)).To(Equal(""))
		})
	})

	Describe(".MakeKey", func() {
		It("should return 'prefixA:prefixB@@age' when key and prefixes are provided", func() {
			actual := MakeKey([]byte("age"), []byte("prefixA"), []byte("prefixB"))
			Expect(string(actual)).To(Equal("prefixA:prefixB@@age"))
		})

		It("should return only concatenated prefixes 'prefixA:prefixB' with no KeyPrefixSeparator when key is not provided", func() {
			actual := MakeKey(nil, []byte("prefixA"), []byte("prefixB"))
			Expect(string(actual)).To(Equal("prefixA:prefixB"))
		})

		It("should return only key with no KeyPrefixSeparator when prefixes are not provided", func() {
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
			kv := NewKVObject([]byte("age"), []byte("value"), []byte("prefixA"), []byte("prefixB"))
			key := MakeKey([]byte("age"), []byte("prefixA"), []byte("prefixB"))
			Expect(kv.GetKey()).To(Equal(key))
		})

		It("should return same result when key is not provided", func() {
			kv := NewKVObject(nil, []byte("value"), []byte("prefixA"), []byte("prefixB"))
			key := MakeKey(nil, []byte("prefixA"), []byte("prefixB"))
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
