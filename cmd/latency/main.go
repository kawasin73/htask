package main

import (
	"flag"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/kawasin73/htask"
)

var (
	n        = flag.Int("n", 1000000, "total task number")
	workers  = flag.Int("worker", 0, "number of workers goroutine")
	interval = flag.Int64("interval", 1000, "task schedule interval (ns)")
)

func main() {
	flag.Parse()
	run(*n, *workers, time.Duration(*interval))
}

type job struct {
	t        time.Time
	chResult chan<- result
}

func (j job) task(expired time.Time) {
	j.chResult <- result{scheduledAt: j.t, expiredAt: expired, executedAt: time.Now()}
}

type result struct {
	scheduledAt time.Time
	expiredAt   time.Time
	executedAt  time.Time
}

func run(total int, workers int, interval time.Duration) {
	chResult := make(chan result, total)
	start := time.Now()
	var wg sync.WaitGroup
	s := htask.NewScheduler(&wg, workers)
	defer func() {
		s.Close()
		wg.Wait()
	}()
	first := start.Add(time.Second + time.Duration(total*2000))
	next := first
	for i := 0; i < total; i++ {
		s.Set(nil, next, job{t: next, chResult: chResult}.task)
		next = next.Add(interval * time.Nanosecond)
	}

	fmt.Printf("set %v tasks in %v. interval = %v, total=%v\n", total, time.Now().Sub(start), interval, interval*time.Duration(total))
	var minExpired, maxExpired, sumExpired, minExecuted, maxExecuted, sumExecuted time.Duration
	minExpired, minExecuted = time.Duration(math.MaxInt64), time.Duration(math.MaxInt64)
	var iExpiredMin, iExpiredMax, iExecutedMin, iExecutedMax int
	var lastExecutedAt time.Time
	for i := 0; i < total; i++ {
		r := <-chResult
		expired := r.expiredAt.Sub(r.scheduledAt)
		if expired > maxExpired {
			maxExpired = expired
			iExpiredMax = i
		}
		if expired < minExpired {
			minExpired = expired
			iExpiredMin = i
		}
		sumExpired += expired

		executed := r.executedAt.Sub(r.expiredAt)
		if executed > maxExecuted {
			maxExecuted = executed
			iExecutedMax = i
		}
		if executed < minExecuted {
			minExecuted = executed
			iExecutedMin = i
		}
		sumExecuted += executed

		if r.executedAt.After(lastExecutedAt) {
			lastExecutedAt = r.executedAt
		}
	}
	meanExpired := sumExpired / time.Duration(total)
	meanExecuted := sumExecuted / time.Duration(total)

	fmt.Printf("all task have executed in %v.\n", lastExecutedAt.Sub(first))
	fmt.Printf("task expired  latency : mean=%v, min=%v, max=%v\n", meanExpired, minExpired, maxExpired)
	fmt.Printf("task executed latency : mean=%v, min=%v, max=%v\n", meanExecuted, minExecuted, maxExecuted)
	fmt.Printf("expired  min index=%v, max index=%v\n", iExpiredMin, iExpiredMax)
	fmt.Printf("executed min index=%v, max index=%v\n", iExecutedMin, iExecutedMax)
}
