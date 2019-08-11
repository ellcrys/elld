package client

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}

var _ = Describe("Client", func() {

	Describe(".NewClient", func() {
		It("should panic when option.host is not set", func() {
			Expect(func() {
				NewClient(nil)
			}).To(Panic())
		})

		It("should set default option.port to 8999 when option.port is not set", func() {
			c := NewClient(&Options{Host: "127.0.0.1"})
			Expect(c.GetOptions().Port).To(Equal(8999))
		})
	})

	Describe(".Call", func() {
		It("should return error when options haven't been set", func() {
			c := Client{opts: &Options{Host: "127.0.0.1"}}
			_, err := c.Call("", nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("http client and options not set"))
		})
	})

	Describe(".GetOptions", func() {
		It("should return options", func() {
			opts := &Options{Host: "hostA"}
			Expect(NewClient(opts).GetOptions()).To(Equal(opts))
		})
	})
})
