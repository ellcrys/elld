package couchdb

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCouchdb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Couchdb Suite")
}
