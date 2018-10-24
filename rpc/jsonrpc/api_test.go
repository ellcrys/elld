package jsonrpc

import (
	"testing"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestAPI(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("API", func() {
		g.Describe("Params", func() {

			type Person struct {
				Name string
				Age  int
			}

			g.It("should decode successfully", func() {
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

		g.Describe("APISet", func() {
			g.It("should return nil if api is not in the set", func() {
				apiSet := APISet(map[string]APIInfo{})
				expected := apiSet.Get("unknown")
				Expect(expected).To(BeNil())
			})
		})
	})
}
