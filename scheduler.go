package htask

import (
	"errors"
	"sync"
	"time"
)

// errors
var (
	ErrClosed         = errors.New("scheduler is already closed")
	ErrInvalidWorkers = errors.New("workers must be more than 0")
	ErrInvalidTime    = errors.New("time is invalid zero time")
	ErrInvalidTask    = errors.New("task must not be nil")
	ErrTaskCancelled  = errors.New("task cancelled")
)

type job struct {
	chCancel <-chan struct{}
	t        time.Time
	task     func(time.Time)
}

// Scheduler is used to schedule tasks.
type Scheduler struct {
	chClose   chan struct{}
	wg        *sync.WaitGroup
	chJob     chan job
	chWork    chan job
	chFin     chan struct{}
	chWorkers chan int
	wNum      int
}

// NewScheduler creates Scheduler and start scheduler and workers.
// number of created goroutines is counted to sync.WaitGroup.
func NewScheduler(wg *sync.WaitGroup, workers int) *Scheduler {
	if workers < 0 {
		workers = 0
	}
	c := &Scheduler{
		chClose:   make(chan struct{}),
		wg:        wg,
		chJob:     make(chan job),
		chWork:    make(chan job),
		chFin:     make(chan struct{}),
		chWorkers: make(chan int),
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go c.worker(wg)
		c.wNum++
	}
	wg.Add(1)
	go c.scheduler(wg, workers)
	return c
}

// Set enqueue new task to scheduler heap queue.
// task will be cancelled by closing chCancel. chCancel == nil is acceptable.
func (c *Scheduler) Set(chCancel <-chan struct{}, t time.Time, task func(time.Time)) error {
	if t.IsZero() {
		return ErrInvalidTime
	} else if task == nil {
		return ErrInvalidTask
	}
	select {
	case <-c.chClose:
		return ErrClosed
	case <-chCancel:
		return ErrTaskCancelled
	case c.chJob <- job{chCancel: chCancel, t: t, task: task}:
		return nil
	}
}

type scheduleState struct {
	heap     *minHeap
	job      job
	chWork   chan<- job
	timer    *time.Timer
	expired  bool // timer is expired or not
	lastTime time.Time

	chWorkPrivate chan<- job // cache
}

func newScheduleState(heapSize int, chWork chan<- job) *scheduleState {
	timer := time.NewTimer(time.Second)
	if !timer.Stop() {
		<-timer.C
	}
	return &scheduleState{
		heap:          newMinHeap(heapSize),
		timer:         timer,
		expired:       true,
		chWorkPrivate: chWork,
	}
}

func (s *scheduleState) add(newJob job) error {
	if err := s.heap.add(newJob); err != nil {
		return err
	}
	if !s.expired && !s.timer.Stop() {
		<-s.timer.C
	}
	// TODO: if job is expired not reset for performance
	s.job = s.heap.peek()
	s.chWork = nil
	// s.job must not be empty
	s.timer.Reset(s.job.t.Sub(time.Now()))
	s.expired = false
	return nil
}

func (s *scheduleState) next() bool {
	if !s.expired && !s.timer.Stop() {
		<-s.timer.C
	}
	_ = s.heap.pop()
	s.job = s.heap.peek()
	if s.job.t.IsZero() {
		s.expired = true
		s.chWork = nil
		return false
	} else if s.job.t.Before(s.lastTime) {
		// skip to reset timer and execute next job directly
		s.expired = true
		s.chWork = s.chWorkPrivate
		s.job.t = s.lastTime
		return true
	} else {
		s.timer.Reset(s.job.t.Sub(time.Now()))
		s.expired = false
		s.chWork = nil
		return false
	}
}

func (s *scheduleState) time(t time.Time) {
	s.expired = true
	s.chWork = s.chWorkPrivate
	s.job.t = t
	s.lastTime = t
}

func (c *Scheduler) scheduler(wg *sync.WaitGroup, workers int) {
	defer wg.Done()
	// no limited min heap
	// TODO: use limited heap
	state := newScheduleState(0, c.chWork)
	for {
		select {
		case <-c.chClose:
			return
		case workers = <-c.chWorkers:
			if workers == 0 && state.chWork != nil {
				go state.job.task(state.job.t)
				for state.next() {
					go state.job.task(state.job.t)
				}
			}
		case newJob := <-c.chJob:
			if err := state.add(newJob); err != nil {
				// TODO: heap is unlimited then no error will occur
				panic(err)
			}
		case <-state.job.chCancel:
			for state.next() {
				if workers == 0 {
					go state.job.task(state.job.t)
				} else {
					// chWork works
					break
				}
			}
		case t := <-state.timer.C:
			state.time(t)
			if workers == 0 {
				go state.job.task(state.job.t)
				for state.next() {
					go state.job.task(state.job.t)
				}
			}
		case state.chWork <- state.job:
			_ = state.next()
		}
	}
}

func (c *Scheduler) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-c.chClose:
			return
		case <-c.chFin:
			return
		case j := <-c.chWork:
			j.task(j.t)
		}
	}
}

// ChangeWorkers will change workers size. workers must greater than 0.
// if new size is smaller, shut appropriate number of workers down.
// if new size is bigger, create appropriate number of workers.
func (c *Scheduler) ChangeWorkers(workers int) error {
	if workers < 0 {
		return ErrInvalidWorkers
	}
	for c.wNum != workers {
		if c.wNum > workers {
			select {
			case <-c.chClose:
				return ErrClosed
			case c.chFin <- struct{}{}:
				c.wNum--
			}
		} else {
			c.wg.Add(1)
			go c.worker(c.wg)
			c.wNum++
		}
	}
	select {
	case <-c.chClose:
	case c.chWorkers <- workers:
		// notify scheduler that workers size have changed
	}
	return nil
}

// Close shutdown scheduler and workers goroutine.
// if Scheduler is already closed then returns ErrClosed.
func (c *Scheduler) Close() error {
	for c.wNum > 0 {
		select {
		case <-c.chClose:
			return ErrClosed
		case c.chFin <- struct{}{}:
		}
		c.wNum--
	}
	select {
	case <-c.chClose:
		return ErrClosed
	default:
		close(c.chClose)
	}
	return nil
}
