package hcron

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
		h := NewMinHeap(0)
		h.Add(Job{t: times[5]})
		h.Add(Job{t: times[4]})
		h.Add(Job{t: times[6]})
		h.Add(Job{t: times[3]})
		h.Add(Job{t: times[4]})

		result := []int{3, 4, 4, 5, 6}
		for _, i := range result {
			if peek := h.Peek(); !peek.t.Equal(times[i]) {
				t.Errorf("peek = %v expected %v", peek.t, times[i])
			}
			if pop := h.Pop(); !pop.t.Equal(times[i]) {
				t.Errorf("pop = %v expected %v", pop.t, times[i])
			}
		}
		if h.Size() != 0 {
			t.Errorf("expect empty but size = %v", h.Size())
		}
		if peek := h.Peek(); !peek.t.IsZero() {
			t.Errorf("empty peek = %v", peek.t)
		}
		if pop := h.Peek(); !pop.t.IsZero() {
			t.Errorf("empty pop = %v", pop.t)
		}
	})

	t.Run("limited min heap", func(t *testing.T) {
		max := 5
		h := NewMinHeap(max)
		h.Add(Job{t: times[5]})
		h.Add(Job{t: times[4]})
		h.Add(Job{t: times[6]})
		h.Add(Job{t: times[3]})
		h.Add(Job{t: times[4]})

		if err := h.Add(Job{t: times[1]}); err != ErrMax {
			t.Errorf("max add expected error: %v, but %v", ErrMax, err)
		}

		result := []int{3, 4, 4, 5, 6}
		for _, i := range result {
			if peek := h.Peek(); !peek.t.Equal(times[i]) {
				t.Errorf("peek = %v expected %v", peek.t, times[i])
			}
			if pop := h.Pop(); !pop.t.Equal(times[i]) {
				t.Errorf("pop = %v expected %v", pop.t, times[i])
			}
		}
		if h.Size() != 0 {
			t.Errorf("expect empty but size = %v", h.Size())
		}
		if peek := h.Peek(); !peek.t.IsZero() {
			t.Errorf("empty peek = %v", peek.t)
		}
		if pop := h.Peek(); !pop.t.IsZero() {
			t.Errorf("empty pop = %v", pop.t)
		}
	})
}
