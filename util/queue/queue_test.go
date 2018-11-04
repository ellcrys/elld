package queue

import (
	"testing"

	"github.com/k0kubun/pp"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

type TestStruct struct {
	Name string
}

func (ts *TestStruct) ID() interface{} {
	return ts.Name
}

func TestCache(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Queue", func() {

		var queue *Queue

		g.BeforeEach(func() {
			queue = New()
		})

		g.Describe(".Append && Head", func() {

			g.It("should append 2 items", func() {
				item := &TestStruct{Name: "ben"}
				item2 := &TestStruct{Name: "glen"}
				queue.Append(item)
				queue.Append(item2)
				Expect(queue.Head()).To(Equal(item))
				Expect(queue.Head()).To(Equal(item2))
				Expect(queue.Head()).To(BeNil())
			})

			g.It("should not add duplicate item", func() {
				item := &TestStruct{Name: "ben"}
				item2 := &TestStruct{Name: "ben"}
				queue.Append(item)
				queue.Append(item2)
				pp.Println(queue.Size())
				Expect(queue.Head()).To(Equal(item))
				Expect(queue.Head()).To(BeNil())
			})
		})

		g.Describe(".Empty", func() {
			g.It("should return true when empty", func() {
				Expect(queue.Empty()).To(BeTrue())
				queue.Append(&TestStruct{Name: "ken"})
				Expect(queue.Empty()).To(BeFalse())
			})
		})

		g.Describe(".Has", func() {
			g.It("should true if item is in the queue", func() {
				item := &TestStruct{Name: "ben"}
				item2 := &TestStruct{Name: "glen"}
				queue.Append(item)
				queue.Append(item2)
				Expect(queue.Has(item)).To(BeTrue())
				Expect(queue.Has(item2)).To(BeTrue())
			})
		})

		g.Describe(".Size", func() {
			g.It("should correct size", func() {
				item := &TestStruct{Name: "ben"}
				item2 := &TestStruct{Name: "glen"}
				queue.Append(item)
				queue.Append(item2)
				Expect(queue.Size()).To(Equal(2))
			})
		})
	})
}
