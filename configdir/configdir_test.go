package configdir

import (
	"fmt"

	"github.com/ellcrys/druid/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configdir", func() {

	Describe(".NewHomeDir", func() {
		It("should return error when the passed in directory does not exist", func() {
			_, err := NewConfigDir("~/not_existing")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("config directory is not ok; may not exist or we don't have enough permission"))
		})

		It("should return nil if the passed in directory exists", func() {

			dirName := fmt.Sprintf("~/%s", util.RandString(10))
			fmt.Println(dirName)
			// os.Remove(fmt)

			// _, err := NewConfigDir("~/not_existing")
			// Expect(err).NotTo(BeNil())
			// Expect(err.Error()).To(Equal("config directory is not ok; may not exist or we don't have enough permission"))
		})
	})
})
