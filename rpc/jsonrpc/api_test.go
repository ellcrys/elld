package jsonrpc

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("API", func() {

	Describe("Params", func() {

		type Person struct {
			Name string
			Age  int
		}

		It("should decode successfully", func() {
			var param = Params(map[string]interface{}{
				"name": "Ben",
				"age":  10,
			})
			var person Person
			err := param.Scan(&person)
			Expect(err).To(BeNil())
			Expect(person.Name).To(Equal("Ben"))
			Expect(person.Age).To(Equal(10))
		})
	})

	Describe("APISet", func() {
		It("should return nil if api is not in the set", func() {
			apiSet := APISet(map[string]APIInfo{})
			expected := apiSet.Get("unknown")
			Expect(expected).To(BeNil())
		})
	})

})
