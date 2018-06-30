package cron

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCron_Every(t *testing.T) {
	var wg sync.WaitGroup
	cron := NewCron(&wg, Option{
		Workers: 1,
	})
	defer func() {
		cron.Close()
		wg.Wait()
	}()

	var count uint64
	chResult := make(chan uint64)

	task := func() {
		chResult <- atomic.AddUint64(&count, 1)
	}

	cancel, _ := cron.Every(10).Millisecond().Run(task)

	for i := 1; i < 10; i++ {
		select {
		case result := <-chResult:
			if i != int(result) {
				t.Errorf("invalid result %v, expected %v", result, i)
			}
		case <-time.After(20 * time.Millisecond):
			t.Fatal("Not come valid result")
		}
	}

	select {
	case <-chResult:
		t.Errorf("interval is smaller than 5 ms")
	case <-time.After(5 * time.Millisecond):
	}

	cancel()

	select {
	case <-chResult:
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case <-chResult:
		t.Errorf("timer is not cancelled")
	case <-time.After(20 * time.Millisecond):
	}

	count = 0

	cancel, _ = cron.Every(10).Millisecond().From(time.Now().Add(100 * time.Millisecond)).Run(task)

	select {
	case <-chResult:
		t.Errorf("Add() unexpectedly expired task")
	case <-time.After(50 * time.Millisecond):
	}
	time.Sleep(40 * time.Millisecond)

	for i := 1; i < 10; i++ {
		select {
		case result := <-chResult:
			if i != int(result) {
				t.Errorf("Add() invalid result %v, expected %v", result, i)
			}
		case <-time.After(20 * time.Millisecond):
			t.Fatal("Add() Not come valid result")
		}
	}
	cancel()
	// not panic when cancel twice.
	cancel()
}

func TestCron_Once(t *testing.T) {
	var wg sync.WaitGroup
	cron := NewCron(&wg, Option{
		Workers: 1,
	})
	defer func() {
		cron.Close()
		wg.Wait()
	}()

	chResult := make(chan struct{})

	task := func() {
		chResult <- struct{}{}
	}

	cancel, _ := cron.Once(time.Now().Add(100 * time.Millisecond)).Run(task)

	select {
	case <-chResult:
		t.Errorf("unexpectedly expired task")
	case <-time.After(50 * time.Millisecond):
	}
	time.Sleep(50 * time.Millisecond)

	select {
	case <-chResult:
	case <-time.After(5 * time.Millisecond):
		t.Errorf("task not executed")
	}

	cancel()
	// not panic when cancel twice.
	cancel()

	cancel, _ = cron.Once(time.Now().Add(100 * time.Millisecond)).Run(task)
	cancel()

	select {
	case <-chResult:
		t.Errorf("unexpectedly expired canceled task")
	case <-time.After(120 * time.Millisecond):
	}
}
