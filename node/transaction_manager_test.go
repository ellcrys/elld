package node

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node", func() {

	var lp *Node
	// var sender = crypto.NewKeyFromIntSeed(int(1))
	var lpPort int

	BeforeEach(func() {
		lpPort = getPort()
		lp = makeTestNode(lpPort)
		Expect(lp.GetBlockchain().Up()).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
	})

	Describe(".AddRaw", func() {

	})
})
