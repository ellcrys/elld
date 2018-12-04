package common

import (
	"os"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common", func() {
	var cfg *config.EngineConfig
	var err error
	var db elldb.DB

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
		db = elldb.NewDB(cfg.DataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe(".HasTxOp", func() {

		It("should return true when TxOp instance is found", func() {
			tx, err := db.NewTx()
			Expect(err).To(BeNil())
			defer tx.Rollback()
			txOp := &OpTx{Tx: tx, CanFinish: true}
			Expect(HasTxOp(txOp)).To(BeTrue())
		})

		It("should return true when TxOp instance is found", func() {
			Expect(HasTxOp(nil)).To(BeFalse())
		})
	})

	Describe(".GetTxOp", func() {

		It("should return transaction option passed to it", func() {
			tx, err := db.NewTx()
			Expect(err).To(BeNil())
			defer tx.Rollback()
			txOp := &OpTx{Tx: tx, CanFinish: true}
			result := GetTxOp(db, txOp)
			Expect(txOp).To(Equal(result))
		})

		It("should create new transaction option if call options does not include a TxOp", func() {
			result := GetTxOp(db)
			Expect(result).ToNot(BeNil())
			result.Finishable().Rollback()
		})

		It("should a finished TxOp when database is closed", func() {
			db.Close()
			txOp := GetTxOp(db)
			Expect(txOp.finished).To(BeTrue())
		})
	})

	Describe(".GetBlockQueryRangeOp", func() {
		It("should get the block range passed to it", func() {
			br := &OpBlockQueryRange{Min: 2, Max: 10}
			result := GetBlockQueryRangeOp(br)
			Expect(result).To(Equal(br))
		})

		It("should return empty BlockQueryRange if call options does not include a BlockQueryRange option", func() {
			result := GetBlockQueryRangeOp(nil)
			Expect(result).ToNot(BeNil())
			Expect(result).To(Equal(&OpBlockQueryRange{}))
		})
	})

	Describe(".GetTransitions", func() {
		It("should get the transitions passed to it", func() {
			var transitions = []Transition{
				&OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}},
			}
			opt := OpTransitions(transitions)
			result := GetTransitions(&opt)
			Expect(result).To(Equal(transitions))
		})

		It("should return an empty slice if no transition option was found", func() {
			result := GetTransitions()
			Expect(result).ToNot(BeNil())
			Expect(result).To(Equal([]Transition{}))
		})
	})

	Describe(".GetChainerOp", func() {
		It("should get the chainer passed to it", func() {
			var chain = OpChainer{name: "chain1"}
			result := GetChainerOp(&chain)
			Expect(result).To(Equal(&chain))
		})

		It("should return empty ChainOp if no chain option was found", func() {
			result := GetChainerOp()
			Expect(result).To(Equal(&OpChainer{}))
		})
	})

	Describe(".ExecAllowed", func() {
		It("should get the value", func() {
			var op = OpAllowExec(true)
			result := ExecAllowed(op)
			Expect(result).To(Equal(true))
			result = ExecAllowed()
			Expect(result).To(Equal(false))
		})
	})

	Describe(".ComputeTxsRoot", func() {
		It("should return expected root", func() {
			txs := []types.Transaction{
				core.NewTransaction(1, 1, "abc", "xyz", "10", "0.01", time.Now().Unix()),
				core.NewTransaction(1, 1, "abc", "xyz", "10", "0.01", time.Now().Unix()),
			}
			root := ComputeTxsRoot(txs)
			Expect(root).To(Equal(util.Hash{
				0x3b, 0x65, 0xc7, 0x5f, 0x8f, 0x61, 0xdd, 0xef, 0x7d, 0x49, 0x67, 0x1f, 0x52, 0x26, 0x76, 0xbb,
				0x7a, 0x46, 0xcc, 0xc0, 0x77, 0x8e, 0x28, 0x78, 0x3e, 0x6e, 0xea, 0x72, 0x90, 0xd9, 0xa8, 0xe3,
			}))
		})
	})

})
