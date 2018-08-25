package common

import (
	"github.com/ellcrys/elld/types/core/objects"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transitions", func() {

	Describe("OpBase.Equal", func() {

		It("it should return true", func() {
			o := &OpBase{Addr: "abc"}
			o2 := &OpBase{Addr: "abc"}
			Expect(o.Equal(o2)).To(BeTrue())
		})

		It("it should return false", func() {
			o := &OpBase{Addr: "abc"}
			o2 := &OpBase{Addr: "xyz"}
			Expect(o.Equal(o2)).To(BeFalse())
		})
	})

	Describe("OpNewAccountBalance.Equal", func() {

		It("it should return true", func() {
			o := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &objects.Account{Balance: "300"}}
			o2 := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &objects.Account{Balance: "300"}}
			Expect(o.Equal(o2)).To(BeTrue())
		})

		It("it should return false if addresses are not the same", func() {
			o := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &objects.Account{Balance: "100"}}
			o2 := &OpNewAccountBalance{OpBase: &OpBase{Addr: "xyz"}, Account: &objects.Account{Balance: "300"}}
			Expect(o.Equal(o2)).To(BeFalse())
		})

		It("it should return false if types are different", func() {
			o := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &objects.Account{Balance: "100"}}
			o2 := &OpBase{Addr: "xyz"}
			Expect(o.Equal(o2)).To(BeFalse())
		})
	})

	Describe("OpCreateAccount.Equal", func() {

		It("it should return true", func() {
			o := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
			o2 := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
			Expect(o.Equal(o2)).To(BeTrue())
		})

		It("it should return false if addresses are not the same", func() {
			o := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
			o2 := &OpCreateAccount{OpBase: &OpBase{Addr: "xyz"}}
			Expect(o.Equal(o2)).To(BeFalse())
		})

		It("it should return false if types are different", func() {
			o := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
			o2 := &OpBase{Addr: "xyz"}
			Expect(o.Equal(o2)).To(BeFalse())
		})
	})
})
