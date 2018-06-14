package node

import (
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransactionSession", func() {

	var protoc *Inception
	var log = logger.NewLogrusNoOp()

	BeforeEach(func() {
		protoc = NewInception(nil, log)
	})

	Describe(".HasTxSession", func() {
		It("should return false when txId is not in the session map", func() {
			Expect(protoc.HasTxSession("some_id")).To(BeFalse())
		})
	})

	Describe(".AddTxSession", func() {
		It("should successfully add txId to the session map", func() {
			protoc.AddTxSession("my_id")
			Expect(protoc.openTxSessions).To(HaveKey("my_id"))
			Expect(protoc.HasTxSession("my_id")).To(BeTrue())
		})
	})

	Describe(".RemoveTxSession", func() {
		It("should successfully remove txId from the session map", func() {
			protoc.AddTxSession("my_id")
			Expect(protoc.openTxSessions).To(HaveKey("my_id"))
			protoc.RemoveTxSession("my_id")
			Expect(protoc.openTxSessions).ToNot(HaveKey("my_id"))
			Expect(protoc.HasTxSession("my_id")).To(BeFalse())
		})
	})

	Describe(".CountTxSession", func() {
		It("should return 2", func() {
			protoc.AddTxSession("my_id")
			protoc.AddTxSession("my_id_2")
			Expect(protoc.CountTxSession()).To(Equal(2))
		})
	})
})
