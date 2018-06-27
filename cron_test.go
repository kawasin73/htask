package hcron

import (
	"context"
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

func TestCron(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	cron := NewCron(ctx, &wg, 1)

	chResult := make(chan int)
	ctxTask := context.Background()
	ctxCancel, cancelTask := context.WithCancel(context.Background())
	cancelTask()

	times := make([]time.Time, 10)
	times[0] = time.Now().Add(time.Millisecond * 100)
	for i := 1; i < 10; i++ {
		times[i] = times[i-1].Add(1)
	}
	cron.Add(ctxTask, times[0], mockTask{ctx: ctxTask, i: 0, chResult: chResult}.Task)
	cron.Add(ctxTask, times[3], mockTask{ctx: ctxTask, i: 1, chResult: chResult}.Task)
	cron.Add(ctxTask, times[2], mockTask{ctx: ctxTask, i: 2, chResult: chResult}.Task)
	cron.Add(ctxTask, times[5], mockTask{ctx: ctxTask, i: 3, chResult: chResult}.Task)
	cron.Add(ctxTask, times[1], mockTask{ctx: ctxTask, i: 4, chResult: chResult}.Task)
	cron.Add(ctxTask, times[1], mockTask{ctx: ctxTask, i: 4, chResult: chResult}.Task)
	cron.Add(ctxCancel, times[4], mockTask{ctx: ctxTask, i: 5, chResult: chResult}.Task)
	cron.Add(ctxTask, times[2], func(ts time.Time) {
		cron.Add(ctxTask, times[7], mockTask{ctx: ctxTask, i: 6, chResult: chResult}.Task)
		cron.Add(ctxTask, times[6], mockTask{ctx: ctxTask, i: 7, chResult: chResult}.Task)
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
		case <-time.After(time.Millisecond * 50):
			t.Fatal("result waited timeout :", i)
		}
	}
	cancel()
	wg.Wait()
}
