package activeobject

import (
	"fmt"
	"sync"

	"gopkg.in/oleiade/lane.v1"
)

// SchedulerFunc describes the signature of a scheduled method
type SchedulerFunc func(args ...interface{}) interface{}

// CallRequest represents a request for a function call.
type CallRequest struct {

	// fName is the name of the function requested
	fName string

	// args includes the arguments to pass to the function
	args []interface{}

	// result is the channel where the result of the call is sent.
	result chan interface{}
}

// Scheduler manages method calls and execute already scheduled calls in a FIFO order.
type Scheduler struct {
	*sync.Mutex

	// queue stores function call requests
	queue *lane.Deque

	// funcs store all functions known to the scheduler
	funcs map[string]SchedulerFunc
}

// NewScheduler creates an instance of Scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		Mutex: &sync.Mutex{},
		queue: lane.NewDeque(),
		funcs: make(map[string]SchedulerFunc),
	}
}

// addFunc adds a function that is executed whenever a given function name is called.
func (s *Scheduler) addFunc(fName string, f SchedulerFunc) {
	s.Lock()
	defer s.Unlock()
	s.funcs[fName] = f
}

// callFunc schedules a request to call a function.
// It panics if no function has been registered under the fName.
// It returns a channel which will be sent the result of the call.
func (s *Scheduler) callFunc(fName string, args ...interface{}) chan interface{} {
	s.Lock()
	_, ok := s.funcs[fName]
	s.Unlock()
	if !ok {
		panic(fmt.Errorf("no registered function with name: %s", fName))
	}

	result := make(chan interface{})
	s.queue.Append(&CallRequest{fName, args, result})
	return result
}

// start begins the executing scheduled functions.
// It should be called on a different thread.
func (s *Scheduler) start() {
	for {
		req := s.queue.Shift()
		if req == nil {
			continue
		}

		_req := req.(*CallRequest)
		fName := _req.fName
		args := _req.args
		result := _req.result
		schFunc := s.funcs[fName]
		result <- schFunc(args...)
	}
}
