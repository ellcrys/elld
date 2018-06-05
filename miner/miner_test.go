package miner

import (
	"github.com/ellcrys/elld/wire"
	"github.com/kr/pretty"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Miner", func() {

	BeforeEach(func() {
		log.SetToDebug()
	})

	Describe(".", func() {
		It("play", func() {

			miner := New(Config{
				NumCPU:         1,
				CachesOnDisk:   1,
				CacheDir:       "CacheFile",
				CachesInMem:    0,
				DatasetDir:     "DagFile",
				DatasetsInMem:  0,
				DatasetsOnDisk: 1,
				PowMode:        ModeNormal,
			})

			b := &wire.Block{
				Header: &wire.Header{
					Number:     1,
					Difficulty: "50000",
				},
			}

			res, err := miner.Begin(b.Header)
			b.Header.MixHash = res.digest
			b.Header.Nonce = res.nonce
			err = miner.VerifyPoW(b.Header)
			pretty.Println("PoW Verification Error:", err)
		})
	})
})
