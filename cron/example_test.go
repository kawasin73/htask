package cron_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kawasin73/htask/cron"
)

func ExampleCron_Every() {
	var wg sync.WaitGroup
	c := cron.NewCron(&wg, cron.Option{
		Workers: 1,
	})
	defer func() {
		c.Close()
		wg.Wait()
	}()

	var count uint64
	task1 := func() {
		fmt.Println("task1 :", atomic.AddUint64(&count, 1))
	}
	task2 := func() {
		fmt.Println("task2 :", atomic.AddUint64(&count, 1))
	}
	noneTask := func() {}

	cancel1, _ := c.Every(100).Millisecond().Run(task1)
	cancel2, _ := c.Every(1).Second().Run(task2)
	cancel2()

	time.Sleep(1010 * time.Millisecond)
	cancel1()

	// executed everyday at 01:02 AM and 3 second.
	c.Every(1).Day().At(1, 2, 3).Run(noneTask)
	// executed everyday and it starts from an hour later.
	c.Every(1).Day().From(time.Now().Add(time.Hour)).Run(noneTask)

	// Output:
	// task1 : 1
	// task1 : 2
	// task1 : 3
	// task1 : 4
	// task1 : 5
	// task1 : 6
	// task1 : 7
	// task1 : 8
	// task1 : 9
	// task1 : 10
	// task1 : 11
}

func ExampleCron_Once() {
	var wg sync.WaitGroup
	c := cron.NewCron(&wg, cron.Option{
		Workers: 1,
	})
	defer func() {
		c.Close()
		wg.Wait()
	}()

	task1 := func() {
		fmt.Println("task1")
	}
	task2 := func() {
		fmt.Println("task2")
	}

	cancel1, _ := c.Once(time.Now().Add(100 * time.Millisecond)).Run(task1)
	cancel2, _ := c.Once(time.Now().Add(200 * time.Millisecond)).Run(task2)
	cancel2()

	time.Sleep(500 * time.Millisecond)
	cancel1()

	// Output:
	// task1
}
