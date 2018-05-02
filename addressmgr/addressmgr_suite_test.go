package addressmgr

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAddressmgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Addressmgr Suite")
}
