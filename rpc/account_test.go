package rpc

import (
	"path"

	"github.com/asaskevich/EventBus"
	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/logic"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Accounts", func() {

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
	})

	AfterEach(func() {
		n.Host().Close()
	})

	Describe(".AccountsGet", func() {
		service := new(Service)

		BeforeEach(func() {
			event := EventBus.New()
			service.logic = logic.New(n, event, log)
		})

		It("should return 0 addresses when no accounts exists", func() {
			payload := map[string]interface{}{}
			var result Result
			err := service.AccountGetAll(payload, &result)
			Expect(err).To(BeNil())
			Expect(result.Data).To(HaveKey("accounts"))
			Expect(result.Data["accounts"]).To(HaveLen(0))
			Expect(result.Status).To(Equal(200))
		})

		It("should return 1 address", func() {
			am := accountmgr.New(path.Join(cfg.ConfigDir(), config.AccountDirName))
			address, _ := crypto.NewKey(nil)
			err := am.CreateAccount(address, "pass123")
			Expect(err).To(BeNil())

			payload := AccountGetAllPayload{}
			var result Result
			err = service.AccountGetAll(payload, &result)
			Expect(err).To(BeNil())
			Expect(result.Data).To(HaveKey("accounts"))
			Expect(result.Data["accounts"]).To(HaveLen(1))

			addresses := result.Data["accounts"].([]string)
			Expect(addresses[0]).To(Equal(address.Addr()))
		})
	})
})
