# htask (min Heap TASK scheduler)

[![CircleCI](https://circleci.com/gh/kawasin73/htask/tree/master.svg?style=svg)](https://circleci.com/gh/kawasin73/htask/tree/master)

In-memory task scheduler using Min Heap implemented in Golang.

htask creates only `1 (scheduler) + n (worker)` goroutines, NOT creating goroutines for each task.

## Install

```bash
go get github.com/kawasin73/htask
```

## Usage

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

- `func NewScheduler(wg *sync.WaitGroup, workers int) *Scheduler`
- `func (s *Scheduler) Set(chCancel <-chan struct{}, t time.Time, task func(time.Time)) error`
- `func (s *Scheduler) ChangeWorkers(workers int) error`
- `func (s *Scheduler) Close() error`

## Notes

- min heap have no limit size.
- when main context is canceled, all pending tasks will be discarded.

## LICENSE

MIT
