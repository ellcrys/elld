package ticket_test

import (
	"math/big"
	"testing"

	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

var log logger.Logger

func init() {
	log = logger.NewLogrusNoOp()
	params.FeePerByte = decimal.NewFromFloat(0.01)
	params.MinimumDifficulty = new(big.Int).SetInt64(100000)
}

func TestTicket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ticket Suite")
}
