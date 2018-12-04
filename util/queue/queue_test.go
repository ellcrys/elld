package queue

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestStruct struct {
	Name string
}

func (ts *TestStruct) ID() interface{} {
	return ts.Name
}

var _ = Describe("Queue", func() {

	var queue *Queue

	BeforeEach(func() {
		queue = New()
	})

	Describe(".Append && Head", func() {

		It("should append 2 items", func() {
			item := &TestStruct{Name: "ben"}
			item2 := &TestStruct{Name: "glen"}
			queue.Append(item)
			queue.Append(item2)
			Expect(queue.Head()).To(Equal(item))
			Expect(queue.Head()).To(Equal(item2))
			Expect(queue.Head()).To(BeNil())
		})

		It("should not add duplicate item", func() {
			item := &TestStruct{Name: "ben"}
			item2 := &TestStruct{Name: "ben"}
			queue.Append(item)
			queue.Append(item2)
			Expect(queue.Head()).To(Equal(item))
			Expect(queue.Head()).To(BeNil())
		})
	})

	Describe(".Empty", func() {
		It("should return true when empty", func() {
			Expect(queue.Empty()).To(BeTrue())
			queue.Append(&TestStruct{Name: "ken"})
			Expect(queue.Empty()).To(BeFalse())
		})
	})

	Describe(".Has", func() {
		It("should true if item is in the queue", func() {
			item := &TestStruct{Name: "ben"}
			item2 := &TestStruct{Name: "glen"}
			queue.Append(item)
			queue.Append(item2)
			Expect(queue.Has(item)).To(BeTrue())
			Expect(queue.Has(item2)).To(BeTrue())
		})
	})

	Describe(".Size", func() {
		It("should correct size", func() {
			item := &TestStruct{Name: "ben"}
			item2 := &TestStruct{Name: "glen"}
			queue.Append(item)
			queue.Append(item2)
			Expect(queue.Size()).To(Equal(2))
		})
	})

})
