package accountmgr

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccountmgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Accountmgr Suite")
}
