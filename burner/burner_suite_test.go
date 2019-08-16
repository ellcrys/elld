package burner

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBurner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Burner Suite")
}
