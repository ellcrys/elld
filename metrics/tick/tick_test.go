package tick

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tick", func() {
	Describe("#MovingAverage", func() {
		Describe(".NewMovingAverage", func() {
			It("should have fields set to nil values", func() {
				tick := NewMovingAverage(1 * time.Second)
				Expect(tick.concludeWindows).To(BeEmpty())
				Expect(tick.currentWindow).To(BeNil())
				Expect(tick.currentWindowEndTime.IsZero()).To(BeTrue())
				Expect(tick.windowDur).To(Equal(1 * time.Second))
			})
		})

		Describe(".Tick", func() {

			When("no current window is set", func() {
				var tick *MovingAverage

				BeforeEach(func() {
					tick = NewMovingAverage(100 * time.Second)
					Expect(tick.currentWindow).To(BeNil())
					tick.Tick()
				})

				Specify("that current window should be set", func() {
					Expect(tick.currentWindow).ToNot(BeNil())
				})

				Specify("that current window count should be 1", func() {
					Expect(tick.currentWindow.tickCount).To(Equal(uint64(1)))
				})

				Specify("that current window end time is a future time", func() {
					Expect(tick.currentWindowEndTime.After(time.Now())).To(BeTrue())
				})
			})

			When("there is an existing window", func() {
				var tick *MovingAverage

				BeforeEach(func() {
					tick = NewMovingAverage(100 * time.Second)
					Expect(tick.currentWindow).To(BeNil())
					tick.Tick()
					Expect(tick.currentWindow.tickCount).To(Equal(uint64(1)))
				})

				It("should increment the window tick count", func() {
					tick.Tick()
					Expect(tick.currentWindow.tickCount).To(Equal(uint64(2)))
				})
			})

			When("the current window has expired", func() {
				var tick *MovingAverage

				BeforeEach(func() {
					tick = NewMovingAverage(100 * time.Second)
					Expect(tick.currentWindow).To(BeNil())
					tick.Tick()
					tick.currentWindowEndTime = time.Now().Add(-10 * time.Second)
				})

				It("should add the expired window to the slice of concluded windows", func() {
					Expect(tick.concludeWindows).To(BeEmpty())
					curWindow := tick.currentWindow
					tick.Tick()
					Expect(tick.concludeWindows).ToNot(BeEmpty())
					Expect(tick.concludeWindows[0]).To(Equal(curWindow))
					Expect(tick.currentWindow).ToNot(BeNil())
					Expect(tick.currentWindow.tickCount).To(Equal(uint64(1)))
				})
			})
		})

		Describe(".clean", func() {
			var tick *MovingAverage
			window1 := &window{tickCount: 1}
			window2 := &window{tickCount: 2}
			window3 := &window{tickCount: 3}
			window4 := &window{tickCount: 4}
			window5 := &window{tickCount: 5}

			BeforeEach(func() {
				tick = NewMovingAverage(1 * time.Second)
				tick.concludeWindows = append(tick.concludeWindows, window1, window2, window3, window4, window5)
				Expect(tick.concludeWindows).To(HaveLen(5))
				Expect(tick.currentWindow).To(BeNil())
			})

			It("should remove older windows when cap is reached", func() {
				WindowCap = 5
				tick.clean()
				Expect(tick.concludeWindows).To(HaveLen(4))
				Expect(tick.concludeWindows[0]).To(Equal(window2))
			})
		})

		Describe(".Averages", func() {
			var tick *MovingAverage

			BeforeEach(func() {
				tick = NewMovingAverage(2 * time.Second)
			})

			It("should panic if dur is not divisible by the window duration and has a remainder", func() {
				Expect(func() {
					tick.Averages(3 * time.Second)
				}).To(Panic())
			})

			It("shouldn't panic if dur is divisible by the window duration without remainder", func() {
				Expect(func() {
					tick.Averages(2 * time.Second)
				}).ToNot(Panic())
			})

			Context("with windows", func() {
				BeforeEach(func() {
					tick = NewMovingAverage(3 * time.Second)
					tick.concludeWindows = append(tick.concludeWindows,
						&window{tickCount: 30},
						&window{tickCount: 32},
						&window{tickCount: 30},
						&window{tickCount: 33},
						&window{tickCount: 29},
						&window{tickCount: 30},
						&window{tickCount: 33})
				})

				It("should averages", func() {
					Expect(tick.Averages(9 * time.Second)).To(Equal([]float64{
						30.666666666666668,
						31.666666666666668,
						30.666666666666668,
						30.666666666666668,
						30.666666666666668,
						31.5,
						33,
					}))
				})
			})
		})

		Describe(".Average", func() {
			var tick *MovingAverage

			BeforeEach(func() {
				tick = NewMovingAverage(3 * time.Second)
				tick.concludeWindows = append(tick.concludeWindows,
					&window{tickCount: 30},
					&window{tickCount: 32},
					&window{tickCount: 30},
					&window{tickCount: 33},
					&window{tickCount: 29},
					&window{tickCount: 30},
					&window{tickCount: 33})
			})

			It("should return 31.261904761904763", func() {
				Expect(func() {
					Expect(tick.Average(9 * time.Second)).To(Equal(31.261904761904763))
				}).ToNot(Panic())
			})

			When("no averages", func() {
				It("should return 0", func() {
					tick.concludeWindows = []*window{}
					Expect(tick.Average(9 * time.Second)).To(Equal(float64(0)))
				})
			})
		})
	})
})
