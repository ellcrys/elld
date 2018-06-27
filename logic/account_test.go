package logic

import (
	path "path/filepath"

	"github.com/ellcrys/elld/crypto"

	evbus "github.com/asaskevich/EventBus"
	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/configdir"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Account", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".AccountsGet", func() {

		var err error
		var n *node.Node
		var logic *Logic
		var bus evbus.Bus
		var errCh chan error

		BeforeEach(func() {
			errCh = make(chan error)
			n, err = node.NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
			Expect(err).To(BeNil())
			bus = evbus.New()
			logic = New(n, bus, log)
		})

		It("should successfully create 2 accounts and return 2 of them", func() {
			k := crypto.NewKeyFromIntSeed(2)
			k2 := crypto.NewKeyFromIntSeed(3)
			am := accountmgr.New(path.Join(n.Cfg().ConfigDir(), configdir.AccountDirName))
			err := am.CreateAccount(k, "my_pass")
			Expect(err).To(BeNil())
			err = am.CreateAccount(k2, "my_pass")
			Expect(err).To(BeNil())

			var result = make(chan []*accountmgr.StoredAccount, 1)
			logic.AccountsGet(result, errCh)
			err = <-errCh
			Expect(err).To(BeNil())
			Expect(<-result).To(HaveLen(2))
		})

	})
})
