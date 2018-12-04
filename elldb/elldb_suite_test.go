package elldb_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestElldb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Elldb Suite")
}
