package htask

import (
	"testing"
	"time"
)

func TestMinHeap(t *testing.T) {
	times := make([]time.Time, 10)
	times[0] = time.Now()
	for i := 1; i < 10; i++ {
		times[i] = times[i-1].Add(time.Second)
	}

	t.Run("no limit min heap", func(t *testing.T) {
		h := newMinHeap(0)
		h.add(job{t: times[5]})
		h.add(job{t: times[4]})
		h.add(job{t: times[6]})
		h.add(job{t: times[3]})
		h.add(job{t: times[4]})

		result := []int{3, 4, 4, 5, 6}
		for _, i := range result {
			if peek := h.peek(); !peek.t.Equal(times[i]) {
				t.Errorf("peek = %v expected %v", peek.t, times[i])
			}
			if pop := h.pop(); !pop.t.Equal(times[i]) {
				t.Errorf("pop = %v expected %v", pop.t, times[i])
			}
		}
		if h.size() != 0 {
			t.Errorf("expect empty but size = %v", h.size())
		}
		if peek := h.peek(); !peek.t.IsZero() {
			t.Errorf("empty peek = %v", peek.t)
		}
		if pop := h.peek(); !pop.t.IsZero() {
			t.Errorf("empty pop = %v", pop.t)
		}
	})

	t.Run("limited min heap", func(t *testing.T) {
		max := 5
		h := newMinHeap(max)
		h.add(job{t: times[5]})
		h.add(job{t: times[4]})
		h.add(job{t: times[6]})
		h.add(job{t: times[3]})
		h.add(job{t: times[4]})

		if err := h.add(job{t: times[1]}); err != ErrMax {
			t.Errorf("max add expected error: %v, but %v", ErrMax, err)
		}

		result := []int{3, 4, 4, 5, 6}
		for _, i := range result {
			if peek := h.peek(); !peek.t.Equal(times[i]) {
				t.Errorf("peek = %v expected %v", peek.t, times[i])
			}
			if pop := h.pop(); !pop.t.Equal(times[i]) {
				t.Errorf("pop = %v expected %v", pop.t, times[i])
			}
		}
		if h.size() != 0 {
			t.Errorf("expect empty but size = %v", h.size())
		}
		if peek := h.peek(); !peek.t.IsZero() {
			t.Errorf("empty peek = %v", peek.t)
		}
		if pop := h.peek(); !pop.t.IsZero() {
			t.Errorf("empty pop = %v", pop.t)
		}
	})
}
