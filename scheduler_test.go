package htask

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockTask struct {
	ctx      context.Context
	i        int
	chResult chan int
}

func (m mockTask) Task(ts time.Time) {
	select {
	case <-m.ctx.Done():
	case m.chResult <- m.i:
	}
}

func TestScheduler(t *testing.T) {
	var wg sync.WaitGroup
	scheduler := NewScheduler(&wg, 1)
	defer func() {
		scheduler.Close()
		wg.Wait()
	}()

	chResult := make(chan int)
	ctx := context.Background()
	ctxCancel, cancelTask := context.WithCancel(context.Background())
	cancelTask()

	times := make([]time.Time, 10)
	times[0] = time.Now().Add(time.Millisecond * 100)
	for i := 1; i < 10; i++ {
		times[i] = times[i-1].Add(1)
	}
	scheduler.Set(ctx.Done(), times[0], mockTask{ctx: ctx, i: 0, chResult: chResult}.Task)
	scheduler.Set(ctx.Done(), times[3], mockTask{ctx: ctx, i: 1, chResult: chResult}.Task)
	scheduler.Set(ctx.Done(), times[2], mockTask{ctx: ctx, i: 2, chResult: chResult}.Task)
	scheduler.Set(ctx.Done(), times[5], mockTask{ctx: ctx, i: 3, chResult: chResult}.Task)
	scheduler.Set(ctx.Done(), times[1], mockTask{ctx: ctx, i: 4, chResult: chResult}.Task)
	scheduler.Set(ctx.Done(), times[1], mockTask{ctx: ctx, i: 4, chResult: chResult}.Task)
	scheduler.Set(ctxCancel.Done(), times[4], mockTask{ctx: ctx, i: 5, chResult: chResult}.Task)
	scheduler.Set(ctx.Done(), times[2], func(ts time.Time) {
		scheduler.Set(ctx.Done(), times[7], mockTask{ctx: ctx, i: 6, chResult: chResult}.Task)
		scheduler.Set(ctx.Done(), times[6], mockTask{ctx: ctx, i: 7, chResult: chResult}.Task)
	})

	select {
	case i := <-chResult:
		t.Fatal("unexpected received result befor timer expired : ", i)
	case <-time.After(time.Millisecond * 10):
	}

	time.Sleep(time.Millisecond * 100)

	result := []int{0, 4, 4, 2, 1, 3, 7, 6}
	for _, i := range result {
		select {
		case r := <-chResult:
			if r != i {
				t.Errorf("result received but not euqal %v != %v", r, i)
			}
		}
	}
}

func TestScheduler_Set(t *testing.T) {
	var wg sync.WaitGroup
	scheduler := NewScheduler(&wg, 1)
	defer func() {
		scheduler.Close()
		wg.Wait()
	}()

	task := func(_ time.Time) { fmt.Println("invalid time") }

	if err := scheduler.Set(nil, time.Time{}, task); err != ErrInvalidTime {
		t.Errorf("zero time error : %v expected %v", err, ErrInvalidTime)
	}
	if err := scheduler.Set(nil, time.Now(), nil); err != ErrInvalidTask {
		t.Errorf("nil task error : %v expected %v", err, ErrInvalidTask)
	}
}
