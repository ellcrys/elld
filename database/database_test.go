package database

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database", func() {

	var testCfgDir = "/Users/ncodes/.ellcrys_test"

	BeforeEach(func() {
		err := os.Mkdir(testCfgDir, 0700)
		Expect(err).To(BeNil())
	})

	Describe(".Open & .Close", func() {

		It("should successfully open and close database", func() {
			db := NewGeneralDB(testCfgDir)
			err := db.Open()
			Expect(err).To(BeNil())
			err = db.Close()
			Expect(err).To(BeNil())
		})

	})

	AfterEach(func() {
		err := os.RemoveAll(testCfgDir)
		Expect(err).To(BeNil())
	})
})
