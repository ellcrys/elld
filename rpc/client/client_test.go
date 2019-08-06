package client_test

import (
	"testing"

	"github.com/ellcrys/partnertracker/rpcclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}

var _ = Describe("Client", func() {
	var client = rpcclient.NewClient()

	Describe(".New", func() {
		It("should panic when opts.host is unset", func() {
			Expect(func() {
				client.New(&rpcclient.Options{})
			}).To(Panic())
		})

		It("should set options.port to 8999 if it is unset", func() {
			c := client.New(&rpcclient.Options{Host: "host"})
			Expect(c.GetOptions().Port).To(Equal(8999))
		})
	})

	Describe(".Call", func() {
		It("should return error when options haven't been set", func() {
			_, err := rpcclient.NewClient().Call("", nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("http client and options not set"))
		})
	})

	Describe(".GetOptions", func() {
		It("should return options", func() {
			opts := &rpcclient.Options{Host: "hostA"}
			Expect(client.New(opts).GetOptions()).To(Equal(opts))
		})
	})
})
