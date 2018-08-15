package blockchain

import (
	"github.com/ellcrys/elld/types/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var MetadataTest = func() bool {
	return Describe("Metadata", func() {

		Context("Metadata", func() {

			var meta = core.BlockchainMeta{}

			Describe(".UpdateMeta", func() {
				It("should successfully save metadata", func() {
					err = bc.updateMeta(&meta)
					Expect(err).To(BeNil())
				})
			})

			Describe(".GetMeta", func() {

				BeforeEach(func() {
					err = bc.updateMeta(&meta)
					Expect(err).To(BeNil())
				})

				It("should return metadata", func() {
					result := bc.GetMeta()
					Expect(result).To(Equal(&meta))
				})
			})
		})

	})
}
