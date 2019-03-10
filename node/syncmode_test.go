package node

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SyncMode", func() {
	Describe(".NewDefaultSyncMode", func() {
		It("should create an instance with the fields initialized", func() {
			s := NewDefaultSyncMode(true)
			Expect(s.disabled).To(BeTrue())
		})
	})

	Describe(".IsDisabled", func() {
		It("should return true/false if `disabled` field is set to true/false", func() {
			s := NewDefaultSyncMode(true)
			Expect(s.IsDisabled()).To(BeTrue())
			s = NewDefaultSyncMode(false)
			Expect(s.IsDisabled()).To(BeFalse())
		})
	})

	Describe(".Enable", func() {
		It("should set `default` field to false", func() {
			s := NewDefaultSyncMode(true)
			s.Enable()
			Expect(s.disabled).To(BeFalse())
		})
	})

	Describe(".Disable", func() {
		It("should set `default` field to true", func() {
			s := NewDefaultSyncMode(false)
			s.Disable()
			Expect(s.disabled).To(BeTrue())
		})
	})
})
