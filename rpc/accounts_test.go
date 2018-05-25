package rpc

import (
	"path"

	"github.com/ellcrys/druid/accountmgr"
	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/node"
	"github.com/ellcrys/druid/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Accounts", func() {

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

	Describe(".AccountsGet", func() {
		service := new(Service)

		BeforeEach(func() {
			service.node = p
		})

		It("should return 0 addresses when no accounts exists", func() {
			payload := GetAccountsPayload{}
			var result Result
			err := service.GetAccounts(payload, &result)
			Expect(err).To(BeNil())
			Expect(result.Data).To(HaveKey("accounts"))
			Expect(result.Data["accounts"]).To(HaveLen(0))
			Expect(result.Status).To(Equal(200))
		})

		It("should return 1 address", func() {
			am := accountmgr.New(path.Join(cfg.ConfigDir(), configdir.AccountDirName))
			address, _ := crypto.NewKey(nil)
			err := am.CreateAccount(address, "pass123")
			Expect(err).To(BeNil())

			payload := GetAccountsPayload{}
			var result Result
			err = service.GetAccounts(payload, &result)
			Expect(err).To(BeNil())
			Expect(result.Data).To(HaveKey("accounts"))
			Expect(result.Data["accounts"]).To(HaveLen(1))

			addresses := result.Data["accounts"].([]string)
			Expect(addresses[0]).To(Equal(address.Addr()))
		})
	})
})
