package histcache_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHistoryCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HistoryCache Suite")
}
