package miner

import (
	"runtime"

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

	It("Begin Result & Digest length should not be Bezo using default Nthread and numCpu", func() {

		res, err := miner.Begin(b.Header)
		Nonce = res.nonce
		MixHash = res.digest

		Expect(len(res.digest)).ShouldNot(BeZero())
		Expect(len(res.result)).ShouldNot(BeZero())
		Expect(err).Should(BeNil())
	})

	It("VerifyPoW err should be nil", func() {

		b.Header.MixHash = MixHash
		b.Header.Nonce = Nonce

		errPow := miner.VerifyPoW(b.Header)
		Expect(errPow).Should(BeNil())
	})

	It("VerifyPoW return errNonPositiveBlock, if Number is block number is 0", func() {
		b.Header.Number = 0
		errPow := miner.VerifyPoW(b.Header)
		Expect(errPow).ShouldNot(BeNil())
	})

	It("VerifyPoW PowMode == ModeTest", func() {

		miner.config.PowMode = ModeTest

		b.Header.MixHash = MixHash
		b.Header.Nonce = Nonce
		b.Header.Number = 54

		errPow := miner.VerifyPoW(b.Header)
		Expect(errPow).ShouldNot(BeNil())
	})

	It("Begin Result & Digest length should not be Bezo setting Nthread and numCpu to higher value", func() {

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

	It("Begin Result & Digest length should not be Bezo setting Nthread  to 0", func() {

		miner.config.NumCPU = 0

		res, err := miner.Begin(b.Header)
		Nonce = res.nonce
		MixHash = res.digest

		Expect(len(res.digest)).ShouldNot(BeZero())
		Expect(len(res.result)).ShouldNot(BeZero())
		Expect(err).Should(BeNil())
	})

	It("Threads cannot be zero", func() {
		Expect(miner.Threads()).ShouldNot(BeZero())
	})
})
