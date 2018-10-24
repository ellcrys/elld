package blockcode

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBlockcode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockcode Suite")
}
