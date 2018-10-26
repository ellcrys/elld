package common

import (
	"testing"

	"github.com/ellcrys/elld/types/core"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestTransitions(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Transitions", func() {

		g.Describe("OpBase.Equal", func() {

			g.It("it should return true", func() {
				o := &OpBase{Addr: "abc"}
				o2 := &OpBase{Addr: "abc"}
				Expect(o.Equal(o2)).To(BeTrue())
			})

			g.It("it should return false", func() {
				o := &OpBase{Addr: "abc"}
				o2 := &OpBase{Addr: "xyz"}
				Expect(o.Equal(o2)).To(BeFalse())
			})
		})

		g.Describe("OpNewAccountBalance.Equal", func() {

			g.It("it should return true", func() {
				o := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &core.Account{Balance: "300"}}
				o2 := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &core.Account{Balance: "300"}}
				Expect(o.Equal(o2)).To(BeTrue())
			})

			g.It("it should return false if addresses are not the same", func() {
				o := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &core.Account{Balance: "100"}}
				o2 := &OpNewAccountBalance{OpBase: &OpBase{Addr: "xyz"}, Account: &core.Account{Balance: "300"}}
				Expect(o.Equal(o2)).To(BeFalse())
			})

			g.It("it should return false if types are different", func() {
				o := &OpNewAccountBalance{OpBase: &OpBase{Addr: "abc"}, Account: &core.Account{Balance: "100"}}
				o2 := &OpBase{Addr: "xyz"}
				Expect(o.Equal(o2)).To(BeFalse())
			})
		})

		g.Describe("OpCreateAccount.Equal", func() {

			g.It("it should return true", func() {
				o := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
				o2 := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
				Expect(o.Equal(o2)).To(BeTrue())
			})

			g.It("it should return false if addresses are not the same", func() {
				o := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
				o2 := &OpCreateAccount{OpBase: &OpBase{Addr: "xyz"}}
				Expect(o.Equal(o2)).To(BeFalse())
			})

			g.It("it should return false if types are different", func() {
				o := &OpCreateAccount{OpBase: &OpBase{Addr: "abc"}}
				o2 := &OpBase{Addr: "xyz"}
				Expect(o.Equal(o2)).To(BeFalse())
			})
		})
	})
}
