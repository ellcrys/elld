package blockchain

import (
	"testing"

	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

var log logger.Logger

func TestBlockchainSuite(t *testing.T) {
	log = logger.NewLogrusNoOp()
	params.FeePerByte = decimal.NewFromFloat(0.01)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockchain Suite")
}
