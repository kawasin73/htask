# htask (min Heap TASK scheduler)

[![CircleCI](https://circleci.com/gh/kawasin73/htask/tree/master.svg?style=svg)](https://circleci.com/gh/kawasin73/htask/tree/master)

![htask.png](./doc/htask.png)

High Scalable In-memory task scheduler using Min Heap implemented in Golang.

htask creates only `1 (scheduler) + n (worker)` goroutines, NOT creating goroutines for each task.

if workers size == 0 then scheduler create goroutine for each task when timer have expired.

`github.com/kawasin73/htask/cron` is wrapper of htask.Scheduler, cron implementation with human friendly interface.

Japanese blog -> [Goでスケーラブルなスケジューラを書いた](https://qiita.com/kawasin73/items/7af6766c7898a656b1ee)

## Install

```bash
go get github.com/kawasin73/htask
```

## Cron Usage

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/kawasin73/htask/cron"
)

func main() {
	var wg sync.WaitGroup
	workers := 1
	c := cron.NewCron(&wg, cron.Option{
		Workers: workers,
	})

	task := func() {
		fmt.Println("hello world")
	}

	// executed every 10:11 AM.
	c.Every(1).Day().At(10, 11).Run(task)

	// task will be executed in every 1 minute from now.
	c.Every(1).Minute().Run(task)

	tenSecondsLater := time.Now().Add(10 * time.Second)
	// executed in every 2 seconds started from 10 seconds later.
	cancel, err := c.Every(2).Second().From(tenSecondsLater).Run(task)
	if err != nil {
		// handle error
	}

	// cron can schedule one time task.
	c.Once(tenSecondsLater.Add(time.Minute)).Run(func() {
		// task can be cancelled.
		cancel()
	})

	c.ChangeWorkers(0)

	time.Sleep(3 * time.Second)

	// on shutdown all queued task will be discarded.
	c.Close()
	wg.Wait()
}

```

## Scheduler Usage

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kawasin73/htask"
)

func main() {
	var wg sync.WaitGroup
	workers := 1
	scheduler := htask.NewScheduler(&wg, workers)

	ctx, _ := context.WithCancel(context.Background())
	scheduler.Set(ctx.Done(), time.Now().Add(time.Second*2), func(t time.Time) {
		fmt.Println("later executed at :", t)
	})
	scheduler.Set(ctx.Done(), time.Now().Add(time.Second), func(t time.Time) {
		fmt.Println("first executed at :", t)
		// it can set to scheduler while executing task.
		scheduler.Set(ctx.Done(), time.Now().Add(time.Millisecond*500), func(t time.Time) {
			fmt.Println("second executed at :", t)
		})
	})

	scheduler.ChangeWorkers(10)

	time.Sleep(3 * time.Second)

	// on shutdown
	scheduler.Close()
	wg.Wait()
}

```

## Interface

Scheduler Interface

- `func NewScheduler(wg *sync.WaitGroup, workers int) *Scheduler`
- `func (s *Scheduler) Set(chCancel <-chan struct{}, t time.Time, task func(time.Time)) error`
- `func (s *Scheduler) ChangeWorkers(workers int) error`
- `func (s *Scheduler) Close() error`

## Notes

- min heap have no limit size.
- when scheduler is closed, all pending tasks will be discarded.

## Benchmarking

```
# using benchstat
go get -u golang.org/x/perf/cmd/benchstat

# benchmark Scheduler.Set()
go test -bench=. -count=10 > bench.txt && benchstat bench.txt

# benchmark latency
go run cmd/latency/main.go -interval=1000000 -n 10000 -worker=0
```

## LICENSE

MIT
