package hcron

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidWorkers = errors.New("workers must be more than 0")
)

type Job struct {
	chDone <-chan struct{}
	t      time.Time
	task   func(time.Time)
}

type Cron struct {
	ctx    context.Context
	chJob  chan Job
	chWork chan Job
	chFin  chan struct{}
	wNum   int
}

func NewCron(ctx context.Context, wg *sync.WaitGroup, workers int) *Cron {
	if workers < 1 {
		workers = 1
	}
	c := &Cron{
		ctx:    ctx,
		chJob:  make(chan Job),
		chWork: make(chan Job),
		chFin:  make(chan struct{}),
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

func (c *Cron) Add(ctx context.Context, t time.Time, task func(time.Time)) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.chJob <- Job{chDone: ctx.Done(), t: t, task: task}:
		return nil
	}
}

func (c *Cron) scheduler(wg *sync.WaitGroup) {
	defer wg.Done()
	// no limited min heap
	// TODO: use limited heap
	h := NewMinHeap(0)
	timer := time.NewTimer(time.Second)
	if !timer.Stop() {
		<-timer.C
	}
	var job Job
	var chWork chan<- Job
	for {
		select {
		case <-c.ctx.Done():
			return
		case newJob := <-c.chJob:
			if err := h.Add(newJob); err != nil {
				// TODO: heap is unlimited then no error will occur
				panic(err)
			}
			if !job.t.IsZero() && !timer.Stop() {
				<-timer.C
			}
			job = h.Peek()
			timer.Reset(job.t.Sub(time.Now()))
		case <-job.chDone:
			if !job.t.IsZero() && !timer.Stop() {
				<-timer.C
			}
			_ = h.Pop()
			job = h.Peek()
			if !job.t.IsZero() {
				timer.Reset(job.t.Sub(time.Now()))
			}
		case t := <-timer.C:
			chWork = c.chWork
			job.t = t
		case chWork <- job:
			chWork = nil
			_ = h.Pop()
			job = h.Peek()
			if !job.t.IsZero() {
				timer.Reset(job.t.Sub(time.Now()))
			}
		}
	}
}

func (c *Cron) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.chFin:
			return
		case job := <-c.chWork:
			job.task(job.t)
		}
	}
}

// ChangeWorkers will change workers size.
// if new size is smaller, shut appropriate number of workers down.
// if new size is bigger, create appropriate number of workers.
func (c *Cron) ChangeWorkers(wg *sync.WaitGroup, workers int) error {
	if workers < 1 {
		return ErrInvalidWorkers
	}
	for c.wNum != workers {
		if c.wNum < workers {
			select {
			case <-c.ctx.Done():
				return c.ctx.Err()
			case c.chFin <- struct{}{}:
				c.wNum--
			}
		} else {
			wg.Add(1)
			go c.worker(wg)
			c.wNum++
		}
	}
	return nil
}
