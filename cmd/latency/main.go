package main

import (
	"flag"
	"fmt"
	"math"
	"sort"
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
	i int
}

func (j job) task(_ time.Time) {
	results[j.i].executedAt = time.Now()
}

type result struct {
	scheduledAt time.Time
	executedAt  time.Time
}

var results []result

func run(total int, workers int, interval time.Duration) {
	results = make([]result, total)
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
		s.Set(nil, next, job{i: i}.task)
		results[i].scheduledAt = next
		next = next.Add(interval * time.Nanosecond)
	}
	chDone := make(chan struct{})
	s.Set(nil, next.Add(interval*time.Nanosecond), func(_ time.Time) {
		time.Sleep(500 * time.Millisecond)
		close(chDone)
	})

	fmt.Printf("set %v tasks in %v. interval = %v, total=%v, workers=%v\n", total, time.Now().Sub(start), interval, interval*time.Duration(total), workers)
	var min, max, sum time.Duration
	min = time.Duration(math.MaxInt64)
	var imin, imax int
	var lastExecutedAt time.Time

	// wait all task executed
	<-chDone

	for i := 0; i < total; i++ {
		r := results[i]
		executed := r.executedAt.Sub(r.scheduledAt)
		if executed > max {
			max = executed
			imax = i
		}
		if executed < min {
			min = executed
			imin = i
		}
		sum += executed

		if r.executedAt.After(lastExecutedAt) {
			lastExecutedAt = r.executedAt
		}
	}
	mean := sum / time.Duration(total)

	sort.Slice(results, func(i, j int) bool {
		return results[i].executedAt.Sub(results[i].scheduledAt) < results[j].executedAt.Sub(results[j].scheduledAt)
	})

	median := results[total/2].executedAt.Sub(results[total/2].scheduledAt)

	fmt.Printf("all task have executed in %v.\n", lastExecutedAt.Sub(first))
	fmt.Printf("task executed latency : mean=%v, median=%v, min=%v, max=%v\n", mean, median, min, max)
	fmt.Printf("executed min index=%v, max index=%v\n", imin, imax)
}
