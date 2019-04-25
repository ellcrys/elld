package activeobject

// ActiveObject provides an API for implements Active Object Pattern
type ActiveObject struct {
	scheduler *Scheduler
}

// NewActiveObject creates an instance of ActiveObject
func NewActiveObject() *ActiveObject {
	return &ActiveObject{scheduler: NewScheduler()}
}

// Start the active object engine.
// It will start the scheduler on a goroutine and return.
func (a *ActiveObject) Start() {
	go a.scheduler.start()
}

// RegisterFunc registers a function to execute against
// a call request that matches fName.
func (a *ActiveObject) RegisterFunc(fName string, f SchedulerFunc) {
	a.scheduler.addFunc(fName, f)
}

// Call schedules a function call.
// It panics if no function has been registered under the fName.
// It returns a channel which will be sent the result of the call.
func (a *ActiveObject) Call(fName string, args ...interface{}) chan interface{} {
	return a.scheduler.callFunc(fName, args...)
}
