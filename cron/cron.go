package cron

import (
	"errors"
	"sync"
	"time"

	"github.com/kawasin73/htask"
)

// Error on job builder
var (
	ErrInvalidAt = errors.New("At(hour, minute, sec, nsec) bigger args")
)

// Cron is wrapper of htask.Scheduler with human friendly interface.
type Cron struct {
	*htask.Scheduler
	loc *time.Location
}

// Option can accept zero value.
// Workers is number of worker goroutine.
// Location is used Every(x).Day().At(hour, minute, sec, nsec) to identify `At` time.
type Option struct {
	Workers  int
	Location *time.Location
}

// NewCron creates Cron.
func NewCron(wg *sync.WaitGroup, option Option) *Cron {
	if option.Location == nil {
		option.Location = time.Local
	}
	return &Cron{
		Scheduler: htask.NewScheduler(wg, option.Workers),
		loc:       option.Location,
	}
}

// Every build intervalJob builder
func (c *Cron) Every(interval int) JobBuilder {
	return JobBuilder{cron: c, num: time.Duration(interval), interval: time.Duration(interval) * time.Second}
}

// Once build OneTimeJob
func (c *Cron) Once(at time.Time) OneTimeJobBuilder {
	return OneTimeJobBuilder{cron: c, at: at}
}

// JobBuilder builds intervalJob with Run method.
type JobBuilder struct {
	cron     *Cron
	num      time.Duration
	interval time.Duration
	from     time.Time
	err      error
}

// Day define Daily or more interval.
func (j JobBuilder) Day() JobBuilder {
	j.interval = j.num * time.Hour * 24
	return j
}

// Hour define Hourly or more interval.
func (j JobBuilder) Hour() JobBuilder {
	j.interval = j.num * time.Hour
	return j
}

// Minute define Minutely or more interval.
func (j JobBuilder) Minute() JobBuilder {
	j.interval = j.num * time.Minute
	return j
}

// Second define Secondly or more interval.
func (j JobBuilder) Second() JobBuilder {
	j.interval = j.num * time.Second
	return j
}

// Millisecond define Millisecondly or more interval.
func (j JobBuilder) Millisecond() JobBuilder {
	j.interval = j.num * time.Millisecond
	return j
}

// At builds the time cron start from.
// At is available only for Day Job.
// accepts 1 ~ 4 arguments. `At(hour, minute, sec, nsec)`
func (j JobBuilder) At(times ...int) JobBuilder {
	var hour, minute, sec, nsec int
	for i, v := range times {
		switch i {
		case 0:
			hour = v
		case 1:
			minute = v
		case 2:
			sec = v
		case 3:
			nsec = v
		default:
			j.err = ErrInvalidAt
			return j
		}
	}
	now := time.Now()
	j.from = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, sec, nsec, j.cron.loc)
	if j.from.Before(now) {
		j.from = j.from.Add(24 * time.Hour)
	}
	return j
}

// From is the time cron start from.
func (j JobBuilder) From(from time.Time) JobBuilder {
	j.from = from
	return j
}

// Run starts Job and returns cancel func.
func (j JobBuilder) Run(task func()) (cancel func(), err error) {
	if j.err != nil {
		return nil, j.err
	}
	chCancel := make(chan struct{})
	if j.from.IsZero() {
		j.from = time.Now()
	}
	job := intervalJob{
		cron:     j.cron,
		chCancel: chCancel,
		next:     j.from,
		interval: j.interval,
		task:     task,
	}
	if err = job.cron.Set(chCancel, job.next, job.callback); err != nil {
		return nil, err
	}
	return func() {
		select {
		case <-chCancel:
		default:
			close(chCancel)
		}
	}, nil
}

type intervalJob struct {
	cron     *Cron
	chCancel chan struct{}
	next     time.Time
	interval time.Duration
	task     func()
}

func (j *intervalJob) nextTime() {
	j.next = j.next.Add(j.interval)
}

func (j *intervalJob) callback(_ time.Time) {
	j.nextTime()
	j.cron.Set(j.chCancel, j.next, j.callback)
	j.task()
}

// OneTimeJobBuilder builds a job executed once.
type OneTimeJobBuilder struct {
	cron *Cron
	at   time.Time
}

// Run starts one time job and returns cancel func.
func (j OneTimeJobBuilder) Run(task func()) (cancel func(), err error) {
	chCancel := make(chan struct{})
	if err := j.cron.Set(chCancel, j.at, func(_ time.Time) { task() }); err != nil {
		return nil, err
	}
	return func() {
		select {
		case <-chCancel:
		default:
			close(chCancel)
		}
	}, nil
}
