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
	var wg sync.WaitGroup
	workers := 1
	cron := hcron.NewCron(&wg, workers)

	ctx, _ := context.WithCancel(context.Background())
	cron.Set(ctx.Done(), time.Now().Add(time.Second*2), func(t time.Time) {
		fmt.Println("later executed at :", t)
	})
	cron.Set(ctx.Done(), time.Now().Add(time.Second), func(t time.Time) {
		fmt.Println("first executed at :", t)
		cron.Set(ctx.Done(), time.Now().Add(time.Millisecond*500), func(t time.Time) {
			fmt.Println("second executed at :", t)
		})
	})

	cron.ChangeWorkers(10)

	time.Sleep(3 * time.Second)

	// on shutdown
	cron.Close()
	wg.Wait()
}

```

## Interface

- `func NewCron(wg *sync.WaitGroup, workers int) *Cron`
- `func (c *Cron) Set(chCancel <-chan struct{}, t time.Time, task func(time.Time)) error`
- `func (c *Cron) ChangeWorkers(workers int) error`
- `func (c *Cron) Close() error`

## Notes

- min heap have no limit size.
- when main context is canceled, all pending tasks will be discarded.

## LICENSE

MIT
