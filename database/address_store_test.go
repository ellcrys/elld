package database

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddressStore", func() {

	var db DB
	var testCfgDir = "/Users/ncodes/.ellcrys_test"

	BeforeEach(func() {
		err := os.Mkdir(testCfgDir, 0700)
		Expect(err).To(BeNil())
		db = NewGeneralDB(testCfgDir)
		err = db.Open()
		Expect(err).To(BeNil())
	})

	Describe(".makeKey", func() {

		It("when key is 'key_a', return 'address-6b65795f61a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a'", func() {
			k := makeKey("key_a")
			Expect(k).ToNot(BeNil())
			Expect(string(k)).To(Equal("address-6b65795f61a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"))
		})

	})

	Describe(".SaveAll", func() {
		It("should successfully save all addresses with no error returned", func() {
			addresses := []string{"address_1", "address_2"}
			err := db.Address().SaveAll(addresses)
			Expect(err).To(BeNil())
		})
	})

	Describe(".GetAll", func() {

		It("should successfully return 2 addresses - address_1 and address_2", func() {
			addresses := []string{"address_1", "address_2"}
			err := db.Address().SaveAll(addresses)
			Expect(err).To(BeNil())
			addrs, err := db.Address().GetAll()
			Expect(err).To(BeNil())
			Expect(addrs).To(HaveLen(2))
			Expect(addrs).To(ContainElement("address_1"))
			Expect(addrs).To(ContainElement("address_2"))
		})
	})

	Describe(".DeleteByPrefix", func() {
		It("should successfully delete all addresses", func() {
			addresses := []string{"address_1", "address_2"}
			err := db.Address().SaveAll(addresses)
			Expect(err).To(BeNil())

			err = db.Address().ClearAll()
			Expect(err).To(BeNil())

			addrs, err := db.Address().GetAll()
			Expect(err).To(BeNil())
			Expect(addrs).To(HaveLen(0))
		})
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).To(BeNil())
		err = os.RemoveAll(testCfgDir)
		Expect(err).To(BeNil())
	})
})
