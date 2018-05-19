package vm

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Vm", func() {

	// var contractBlock *Contract
	// var v *VM

	// type MyData struct {
	// 	Amount float64
	// }

	// BeforeEach(func() {
	// 	v = New()
	// })

	// AfterEach(func() {
	// 	err := v.Stop()
	// 	Expect(err).To(BeNil())
	// })

	// Describe(".Init", func() {

	// 	It("should start a new instance of VM and initialize a full block of contracts", func() {
	// 		done := v.Init(new(TestEllBlock))
	// 		Expect(done).To(BeTrue())
	// 	})

	// })

	// It("should execute a function on contract", func() {
	// 	contractBlock = &Contract{
	// 		ID: contractID,
	// 		Transactions: []Transaction{{
	// 			Function: "DoSomething",
	// 			Data: MyData{
	// 				Amount: 10.50,
	// 			},
	// 		}},
	// 		Archive: "../test_contract.zip",
	// 	}

	// 	done := v.Exec(contractBlock, func(err error) {
	// 		//fmt.Printf("%v\n", err)
	// 		Expect(err).To(BeNil())
	// 	})
	// 	Expect(done).To(BeTrue())
	// })

})
