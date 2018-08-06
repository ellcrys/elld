package rpc

import (
	"time"

	"github.com/ellcrys/elld/logic"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transactions", func() {

	var n *node.Node
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
		n, err = node.NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
		Expect(err).To(BeNil())
		gossip := node.NewGossip(n, log)
		n.SetGossipProtocol(gossip)
	})

	AfterEach(func() {
		n.Host().Close()
	})

	Describe(".Send", func() {
		service := new(Service)

		BeforeEach(func() {
			service.logic, _ = logic.New(n, log)
		})

		It("should return 0 addresses when no accounts exists", func() {
			payload := map[string]interface{}{}
			var result Result
			err := service.TransactionAdd(payload, &result)
			Expect(err).To(BeNil())
			Expect(result.Error).ToNot(BeEmpty())
			Expect(result.Error).To(Equal("unknown transaction type"))
			Expect(result.ErrCode).To(Equal(errCodeUnknownTransactionType))
			Expect(result.Status).To(Equal(400))
		})

		It("should return error when transaction is invalid", func() {

			addr, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, addr.Addr(), sender.PubKey().Base58(), "10", "10", time.Now().Unix())
			tx.From = sender.Addr()
			sig, _ := wire.TxSign(tx, sender.PrivKey().Base58())
			tx.Hash = util.StrToHash("invalid_hash")

			payload := map[string]interface{}{
				"type":         tx.Type,
				"nonce":        tx.Nonce,
				"from":         tx.From,
				"senderPubKey": tx.SenderPubKey,
				"to":           tx.To,
				"value":        tx.Value,
				"fee":          tx.Fee,
				"timestamp":    tx.Timestamp,
				"sig":          sig,
				"hash":         tx.Hash,
			}

			var result Result
			err = service.TransactionAdd(payload, &result)
			Expect(result.Error).ToNot(BeEmpty())
			Expect(result.Error).To(Equal("index:0, field:hash, error:hash is not correct"))
		})

		It("should successfully send transaction", func() {

			addr, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, addr.Addr(), sender.PubKey().Base58(), "10", "10", time.Now().Unix())
			tx.From = sender.Addr()
			sig, _ := wire.TxSign(tx, sender.PrivKey().Base58())
			tx.Hash = tx.ComputeHash()

			payload := map[string]interface{}{
				"type":         tx.Type,
				"nonce":        tx.Nonce,
				"from":         tx.From,
				"senderPubKey": tx.SenderPubKey,
				"to":           tx.To,
				"value":        tx.Value,
				"fee":          tx.Fee,
				"timestamp":    tx.Timestamp,
				"sig":          sig,
				"hash":         tx.Hash,
			}

			var result Result
			service.TransactionAdd(payload, &result)
			Expect(result.Error).To(BeEmpty())
			Expect(result.Status).To(Equal(200))
			Expect(result.Data).To(HaveKey("txId"))
		})

	})
})
