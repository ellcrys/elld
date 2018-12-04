package txpool_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTxpool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Txpool Suite")
}
