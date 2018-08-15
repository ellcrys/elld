package stub_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoStub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Go.Stub Suite")
}
