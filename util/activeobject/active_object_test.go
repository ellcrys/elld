package activeobject

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActiveObject", func() {
	Describe(".NewActiveObject", func() {
		It("should create an instance", func() {
			o := NewActiveObject()
			Expect(o.scheduler).ToNot(BeNil())
		})
	})

	Describe(".RegisterFunc", func() {
		o := NewActiveObject()

		BeforeEach(func() {
			o.RegisterFunc("f1", func(args ...interface{}) interface{} { return nil })
			o.RegisterFunc("f2", func(args ...interface{}) interface{} { return nil })
		})

		It("should register functions `f1` and `f2`", func() {
			Expect(o.scheduler.funcs).To(HaveKey("f1"))
			Expect(o.scheduler.funcs).To(HaveKey("f2"))
		})
	})

	Describe(".Call", func() {
		var o *ActiveObject

		BeforeEach(func() {
			o = NewActiveObject()
		})

		It("should call panic when called function was not registered", func() {
			Expect(func() {
				o.Call("f1")
			}).To(Panic())
		})

		When("called function was registered", func() {
			BeforeEach(func() {
				o.RegisterFunc("f1", func(args ...interface{}) interface{} { return nil })
				Expect(o.scheduler.queue.Size()).To(BeZero())
			})

			It("should schedule the function call request", func() {
				o.Call("f1")
				Expect(o.scheduler.queue.Size()).To(Equal(1))
			})
		})

		When("two functions are called", func() {
			BeforeEach(func() {
				o.RegisterFunc("f1", func(args ...interface{}) interface{} { return nil })
				o.RegisterFunc("f2", func(args ...interface{}) interface{} { return nil })
				Expect(o.scheduler.queue.Size()).To(BeZero())
			})

			Specify("that the call request are scheduled in FIFO order", func() {
				o.Call("f1")
				o.Call("f2")
				Expect(o.scheduler.queue.Size()).To(Equal(2))
				item0 := o.scheduler.queue.Shift().(*CallRequest)
				item1 := o.scheduler.queue.Shift().(*CallRequest)
				Expect(item0.fName).To(Equal("f1"))
				Expect(item1.fName).To(Equal("f2"))
			})
		})

		Context("with a running scheduler", func() {
			var o *ActiveObject

			BeforeEach(func() {
				o = NewActiveObject()
				o.Start()
			})

			It("should call the registered function and receive result", func() {
				o.RegisterFunc("add", func(args ...interface{}) interface{} {
					return args[0].(int) + args[1].(int)
				})
				result := o.Call("add", 2, 2)
				Expect(<-result).To(Equal(4))
			})
		})
	})
})
