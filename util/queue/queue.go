package queue

import (
	"sync"

	"gopkg.in/oleiade/lane.v1"
)

// Item represents a queue item
type Item interface {
	ID() interface{}
}

// Queue wraps lane.Deque which is a head-tail
// linked list. It maintains an index of items
// contained in the queue
type Queue struct {
	sync.RWMutex
	q     *lane.Deque
	index map[interface{}]struct{}
}

// New creates an instance of Queue
func New() *Queue {
	return &Queue{
		q:     lane.NewDeque(),
		index: make(map[interface{}]struct{}),
	}
}

// Head get an item from the head of the queue
func (q *Queue) Head() Item {
	q.Lock()
	defer q.Unlock()
	el := q.q.Shift()
	if el != nil {
		delete(q.index, el)
	} else {
		return nil
	}
	return el.(Item)
}

// Append appends an item to the queue
func (q *Queue) Append(i Item) {
	if q.Has(i) {
		return
	}
	q.Lock()
	defer q.Unlock()
	q.q.Append(i)
	q.index[i.ID()] = struct{}{}
}

// Empty checks whether the queue is empty
func (q *Queue) Empty() bool {
	return q.q.Empty()
}

// Has checks whether a item exist in the queue
func (q *Queue) Has(i Item) bool {
	q.RLock()
	defer q.RUnlock()
	_, ok := q.index[i.ID()]
	return ok
}

// Size returns the size of the queue
func (q *Queue) Size() int {
	q.RLock()
	defer q.RUnlock()
	return q.q.Size()
}
