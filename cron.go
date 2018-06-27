package hcron

import (
	"errors"
	"sync"
	"time"
)

// errors
var (
	ErrClosed         = errors.New("cron is already closed")
	ErrInvalidWorkers = errors.New("workers must be more than 0")
	ErrTaskCancelled  = errors.New("task cancelled")
)

type job struct {
	chCancel <-chan struct{}
	t        time.Time
	task     func(time.Time)
}

// Cron is used to schedule tasks.
type Cron struct {
	chClose chan struct{}
	wg      *sync.WaitGroup
	chJob   chan job
	chWork  chan job
	chFin   chan struct{}
	wNum    int
}

// NewCron creates Cron and start scheduler and workers.
// number of created goroutines is counted to sync.WaitGroup.
func NewCron(wg *sync.WaitGroup, workers int) *Cron {
	if workers < 1 {
		workers = 1
	}
	c := &Cron{
		chClose: make(chan struct{}),
		wg:      wg,
		chJob:   make(chan job),
		chWork:  make(chan job),
		chFin:   make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go c.worker(wg)
		c.wNum++
	}
	wg.Add(1)
	go c.scheduler(wg)
	return c
}

// Set enqueue new task to scheduler heap queue.
// task will be cancelled by closing chCancel.
func (c *Cron) Set(chCancel <-chan struct{}, t time.Time, task func(time.Time)) error {
	select {
	case <-c.chClose:
		return ErrClosed
	case <-chCancel:
		return ErrTaskCancelled
	case c.chJob <- job{chCancel: chCancel, t: t, task: task}:
		return nil
	}
}

func (c *Cron) scheduler(wg *sync.WaitGroup) {
	defer wg.Done()
	// no limited min heap
	// TODO: use limited heap
	h := newMinHeap(0)
	timer := time.NewTimer(time.Second)
	if !timer.Stop() {
		<-timer.C
	}
	var j job
	var chWork chan<- job
	for {
		select {
		case <-c.chClose:
			return
		case newJob := <-c.chJob:
			if err := h.add(newJob); err != nil {
				// TODO: heap is unlimited then no error will occur
				panic(err)
			}
			if chWork == nil && !j.t.IsZero() && !timer.Stop() {
				<-timer.C
			}
			j = h.peek()
			timer.Reset(j.t.Sub(time.Now()))
		case <-j.chCancel:
			if chWork == nil && !j.t.IsZero() && !timer.Stop() {
				<-timer.C
			}
			_ = h.pop()
			j = h.peek()
			if !j.t.IsZero() {
				timer.Reset(j.t.Sub(time.Now()))
			}
		case t := <-timer.C:
			chWork = c.chWork
			j.t = t
		case chWork <- j:
			chWork = nil
			_ = h.pop()
			j = h.peek()
			if !j.t.IsZero() {
				timer.Reset(j.t.Sub(time.Now()))
			}
		}
	}
}

func (c *Cron) worker(wg *sync.WaitGroup) {
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
func (c *Cron) ChangeWorkers(workers int) error {
	if workers < 1 {
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
	return nil
}

// Close shutdown scheduler and workers goroutine.
// if cron is already closed then returns ErrClosed.
func (c *Cron) Close() error {
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
