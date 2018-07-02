package node

import (
	evbus "github.com/asaskevich/EventBus"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransactionSession", func() {

	var err error
	var gossip *Gossip
	var log = logger.NewLogrusNoOp()
	var n *Node

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		n, err = NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
		Expect(err).To(BeNil())
		gossip = NewGossip(n, log)
		bus := evbus.New()
		n.SetLogicBus(bus)
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".HasTxSession", func() {
		It("should return false when txId is not in the session map", func() {
			Expect(gossip.HasTxSession("some_id")).To(BeFalse())
		})
	})

	Describe(".AddTxSession", func() {
		It("should successfully add txId to the session map", func() {
			gossip.AddTxSession("my_id")
			Expect(gossip.openTransactionsSession).To(HaveKey("my_id"))
			Expect(gossip.HasTxSession("my_id")).To(BeTrue())
		})
	})

	Describe(".RemoveTxSession", func() {
		It("should successfully remove txId from the session map", func() {
			gossip.AddTxSession("my_id")
			Expect(gossip.openTransactionsSession).To(HaveKey("my_id"))
			gossip.RemoveTxSession("my_id")
			Expect(gossip.openTransactionsSession).ToNot(HaveKey("my_id"))
			Expect(gossip.HasTxSession("my_id")).To(BeFalse())
		})
	})

	Describe(".CountTxSession", func() {
		It("should return 2", func() {
			gossip.AddTxSession("my_id")
			gossip.AddTxSession("my_id_2")
			Expect(gossip.CountTxSession()).To(Equal(2))
		})
	})
})
