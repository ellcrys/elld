package logic

import (
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	var err error
	var n *node.Node
	var logic *Logic
	var errCh chan error

	BeforeEach(func() {
		errCh = make(chan error)
		n, err = node.NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
		Expect(err).To(BeNil())

		err = n.OpenDB()
		Expect(err).To(BeNil())

		logic, _ = New(n, log)
	})

	AfterEach(func() {
		n.Stop()
	})

	Describe(".ObjectsPut", func() {

		It("should successfully add object without error", func() {
			addresses := []*elldb.KVObject{elldb.NewKVObject([]byte("age"), []byte("20"))}
			errCh = make(chan error, 1)
			logic.ObjectsPut(addresses, errCh)
			err := <-errCh
			Expect(err).To(BeNil())
		})
	})

	Describe(".ObjectsGet", func() {

		It("should successfully get objects without error", func() {
			addresses := []*elldb.KVObject{
				elldb.NewKVObject([]byte("age"), []byte("20"), "ns"),
				elldb.NewKVObject([]byte("sex"), []byte("unknown"), "ns"),
			}
			errCh = make(chan error, 1)
			logic.ObjectsPut(addresses, errCh)
			err := <-errCh
			Expect(err).To(BeNil())

			var result = make(chan []*elldb.KVObject, 1)
			logic.ObjectsGet([]byte("ns"), result)
			objs := <-result
			Expect(objs).To(HaveLen(2))

		})
	})
})
