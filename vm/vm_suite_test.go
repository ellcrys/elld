package vm

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vm Suite")
}

type TestEllBlock struct {
}

const contractID = "376518103780890"

func (eb *TestEllBlock) GetContracts() []Contract {
	contracts := []Contract{{
		ID:      contractID,
		Archive: "../test_contract.zip",
	}}

	return contracts
}

var _ = Describe("Container Tests", func() {
	var contractBlock *Contract
	var v *VM

	type MyData struct {
		Amount float64
	}

	BeforeEach(func() {
		contractBlock = &Contract{
			ID: contractID,
			Transactions: []Transaction{{
				Function: "DoSomething",
				Data: MyData{
					Amount: 10.50,
				},
			}},
			Archive: "../test_contract.zip",
		}
	})

	AfterEach(func() {
		time.Sleep(3 * time.Second)
	})

	It("should start a new instance of VM and initialize a full block of contracts", func() {
		v = New()

		done := v.Init(new(TestEllBlock))
		Expect(done).To(BeTrue())
	})

	It("should execute a function on contract", func() {

		done := v.Exec(contractBlock)
		if !done {
			v.Stop()
		}
		Expect(done).To(BeTrue())
	})

})
