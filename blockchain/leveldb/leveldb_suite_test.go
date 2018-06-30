package leveldb

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLeveldb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Leveldb Suite")
}
