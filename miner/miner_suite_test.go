package miner

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMiner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Miner Suite")
}
