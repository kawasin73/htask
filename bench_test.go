package htask_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/kawasin73/htask"
)

func BenchmarkScheduler_Set(b *testing.B) {
	times := make([]time.Time, b.N)
	times[0] = time.Now().Add(time.Hour)
	for i := 1; i < b.N; i++ {
		times[i] = times[i-1].Add(time.Second)
	}
	rand.Shuffle(b.N, func(i, j int) {
		times[i], times[j] = times[j], times[i]
	})

	task := func(_ time.Time) {}

	var wg sync.WaitGroup

	s := htask.NewScheduler(&wg, 0)

	defer func() {
		b.StopTimer()
		s.Close()
		wg.Wait()
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := s.Set(nil, times[i], task); err != nil {
			b.Errorf("error occurred : %v", err)
		}
	}
}
