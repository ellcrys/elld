package tick

import (
	"math"
	"sync"
	"time"
)

var WindowCap = 10000
var CleanPct = 0.1

type window struct {
	tickCount uint64
}

// MovingAverage collates ticks and determines the moving
// average of the collated ticks. It collects ticks within
// fixed time window, then uses the count of ticks in the
// windows to calculate the moving average.
type MovingAverage struct {
	sync.Mutex
	currentWindow        *window
	currentWindowEndTime time.Time
	concludeWindows      []*window
	windowDur            time.Duration
}

// NewMovingAverage creates an instance of MovingAverage.
// windowDur is the lifespan of a window
func NewMovingAverage(windowDur time.Duration) *MovingAverage {
	return &MovingAverage{
		windowDur: windowDur,
	}
}

// Tick increments the tick count of the current window
// and creates a new window if the current window has expired
func (m *MovingAverage) Tick() {
	m.Lock()
	defer m.Unlock()

	now := time.Now()

	// Create a new window if no current window is active
	if m.currentWindow == nil {
		m.currentWindow = new(window)
		m.currentWindowEndTime = time.Now().Add(m.windowDur)
	}

	// When the current window expires, add it to the
	// list of concluded windows and create a new window
	if m.currentWindowEndTime.Before(now) {
		m.concludeWindows = append(m.concludeWindows, m.currentWindow)
		m.currentWindow = new(window)
		m.currentWindowEndTime = time.Now().Add(m.windowDur)
	}

	// Increment window's tick count
	m.currentWindow.tickCount++

	m.clean()
}

// clean removes older windows from the slice of
// concluded windows. When the number of concluded
// windows equal the cap, 10% of the older windows
// are dropped.
// Note: Not thread-safe. Must be called with lock acquired
func (m *MovingAverage) clean() {
	if curLen := len(m.concludeWindows); curLen == WindowCap {
		var tenPct = math.Round(CleanPct * float64(WindowCap))
		m.concludeWindows = m.concludeWindows[int(tenPct):]
	}
}

// Averages calculates the dur moving averages.
// dur must be divisible by the initial window duration
// without remainder. E.g if initial window duration is
// 10 seconds, then 10, 20, 30, 40 are valid values
func (m *MovingAverage) Averages(dur time.Duration) []float64 {
	m.Lock()
	defer m.Unlock()

	if math.Mod(dur.Seconds(), m.windowDur.Seconds()) != 0 {
		panic("dur must be divisible by initial window duration without remainder")
	}

	slideCount := dur.Seconds() / m.windowDur.Seconds()
	nConcludedWin := len(m.concludeWindows)
	averages := []float64{}
	for i, w := range m.concludeWindows {
		curSlideWindows := []*window{w}
		slideEnd := i + int(slideCount) - 1
		next := i + 1
		for next < nConcludedWin {
			curSlideWindows = append(curSlideWindows, m.concludeWindows[next])
			if next == slideEnd {
				break
			}
			next++
		}
		slideSum := uint64(0)
		for _, _w := range curSlideWindows {
			slideSum += _w.tickCount
		}
		averages = append(averages, float64(slideSum)/float64(len(curSlideWindows)))
	}

	return averages
}

// Average calculates the average of all derived
// moving averages of dur. dur must be divisible by the
// initial window duration without remainder. E.g if
// initial window duration is 10 seconds, then 10, 20,
// 30, 40 are valid values
func (m *MovingAverage) Average(dur time.Duration) float64 {
	averages := m.Averages(dur)
	if len(averages) == 0 {
		return 0
	}
	sum := 0.0
	for _, avg := range averages {
		sum += avg
	}
	return sum / float64(len(averages))
}
