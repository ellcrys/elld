package rpc

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transactions", func() {

	var p *node.Node
	var err error

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	BeforeEach(func() {
		p, err = node.NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		p.Host().Close()
	})

	Describe(".Send", func() {
		service := new(Service)

		BeforeEach(func() {
			service.node = p
		})

		It("should return 0 addresses when no accounts exists", func() {
			payload := SendTxPayload{Args: SendTxArgs{}}
			var result Result
			err := service.Send(payload, &result)
			Expect(err).To(BeNil())
			Expect(result.Error).ToNot(BeEmpty())
			Expect(result.Error).To(Equal("unknown transaction type"))
			Expect(result.ErrCode).To(Equal(errCodeUnknownTransactionType))
			Expect(result.Status).To(Equal(400))
		})

		It("should successfully add transaction to tx pool", func() {

			addr, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, addr.Addr(), sender.PubKey().Base58(), "10", "10", time.Now().Unix())
			payload := SendTxPayload{Args: SendTxArgs{
				Type:         tx.Type,
				Nonce:        tx.Nonce,
				SenderPubKey: tx.SenderPubKey,
				To:           tx.To,
				Value:        tx.Value,
				Fee:          tx.Fee,
				Timestamp:    tx.Timestamp,
			}}

			payload.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())

			var result Result
			service.Send(payload, &result)
			Expect(result.Error).To(BeEmpty())

			Expect(service.node.GetTxPool().Has(tx)).To(BeTrue())
		})

	})
})
