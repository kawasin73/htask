# hcron (min Heap CRON)

[![CircleCI](https://circleci.com/gh/kawasin73/hcron/tree/master.svg?style=svg)](https://circleci.com/gh/kawasin73/hcron/tree/master)

In-memory task scheduler using Min Heap implemented in Golang.

hcron creates only `1 (scheduler) + n (worker)` goroutines, NOT creating goroutines for each task.

## Install

```bash
go get github.com/kawasin73/hcron
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kawasin73/hcron"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	workerNumber := 1
	cron := hcron.NewCron(ctx, &wg, workerNumber)

	ctxTask, _ := context.WithCancel(context.Background())
	cron.Add(ctxTask, time.Now().Add(time.Minute), func(t time.Time) {
		fmt.Println("later executed at :", t)
	})
	cron.Add(ctxTask, time.Now().Add(30*time.Second), func(t time.Time) {
		fmt.Println("first executed at :", t)
	})

	cron.ChangeWorkers(&wg, 10)

	// on shutdown
	cancel()
	wg.Wait()
}

```

## Interface

- `func NewCron(ctx context.Context, wg *sync.WaitGroup, workers int) *Cron`
- `func (c *Cron) Add(ctx context.Context, t time.Time, task func(time.Time)) error`
- `func (c *Cron) ChangeWorkers(wg *sync.WaitGroup, workers int) error`

## Notes

- min heap have no limit size.
- when main context is canceled, all pending tasks will be discarded.

## LICENSE

MIT
