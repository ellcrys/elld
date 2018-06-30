package miner

import (
	"runtime"

	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Miner", func() {
	miner := New(Config{
		NumCPU:         1,
		CachesOnDisk:   1,
		CacheDir:       "CacheFile",
		CachesInMem:    0,
		DatasetDir:     "DagFile",
		DatasetsInMem:  0,
		DatasetsOnDisk: 1,
		PowMode:        ModeTest,
	})

	b := &wire.Block{
		Header: &wire.Header{
			Number:     1,
			Difficulty: "50000",
		},
	}

	var MixHash []byte
	var Nonce uint64

	Describe(".Begin", func() {
		Context("Begin function using default Nthread and numCpu", func() {

			It("err should be nil", func() {

				res, err := miner.Begin(b.Header)
				Nonce = res.nonce
				MixHash = res.digest

				Expect(len(res.digest)).ShouldNot(BeZero())
				Expect(len(res.result)).ShouldNot(BeZero())
				Expect(err).Should(BeNil())
			})

		})

		Context("Setting NumCpu to 0", func() {

			It("err Should be Nil", func() {

				miner.config.NumCPU = 0

				res, err := miner.Begin(b.Header)
				Nonce = res.nonce
				MixHash = res.digest

				Expect(len(res.digest)).ShouldNot(BeZero())
				Expect(len(res.result)).ShouldNot(BeZero())
				Expect(err).Should(BeNil())
			})

		})

		Context("Setting Nthread and numCpu to higher value", func() {

			It("err must be nil", func() {

				numCPU := runtime.NumCPU()
				nThreads := miner.config.NumCPU + numCPU
				miner.config.NumCPU = nThreads

				res, err := miner.Begin(b.Header)
				Nonce = res.nonce
				MixHash = res.digest

				Expect(len(res.digest)).ShouldNot(BeZero())
				Expect(len(res.result)).ShouldNot(BeZero())
				Expect(err).Should(BeNil())
			})

		})
	})

	Describe(".VerifyPoW", func() {

		Context("VerifyPoW err should be nil if all parameters are valid", func() {
			It("It should be nil", func() {
				b.Header.MixHash = util.ToHex(MixHash)
				b.Header.Nonce = Nonce
				errPow := miner.VerifyPoW(b.Header)
				Expect(errPow).Should(BeNil())
			})
		})

		Context("VerifyPoW should generate error when block number is 0", func() {

			It("It should not be nil", func() {
				b.Header.Number = 0
				errPow := miner.VerifyPoW(b.Header)
				Expect(errPow).ShouldNot(BeNil())
			})

		})

		Context("If PowMode == ModeTest", func() {
			It("It should not be Nil", func() {
				miner.config.PowMode = ModeTest
				b.Header.MixHash = util.ToHex(MixHash)
				b.Header.Nonce = Nonce
				b.Header.Number = 54

				errPow := miner.VerifyPoW(b.Header)
				Expect(errPow).ShouldNot(BeNil())
			})
		})

	})

	Describe(".Threads", func() {
		It("It Should not be zero", func() {
			Expect(miner.Threads()).ShouldNot(BeZero())
		})
	})
})
