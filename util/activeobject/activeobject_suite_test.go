package activeobject_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestActiveobject(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ActiveObject Suite")
}
