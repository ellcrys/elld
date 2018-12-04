package jsonrpc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestJsonrpc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jsonrpc Suite")
}
