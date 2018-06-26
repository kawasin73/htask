package hcron

import (
	"container/heap"
	"errors"
)

type jobHeap []Job

func (h *jobHeap) Len() int {
	return len(*h)
}

func (h *jobHeap) Less(i, j int) bool {
	return !(*h)[i].t.After((*h)[j].t)
}

// Swap swaps the elements with indexes i and j.
func (h *jobHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *jobHeap) Push(x interface{}) {
	*h = append(*h, x.(Job))
}

func (h *jobHeap) Pop() (x interface{}) {
	x, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]
	return
}

var (
	ErrMax = errors.New("heap max size")
)

// MinHeap is min heap for Cron
type MinHeap struct {
	heap jobHeap
	max  int
}

func NewMinHeap(max int) *MinHeap {
	if max < 0 {
		max = 0
	}
	h := &MinHeap{heap: make(jobHeap, 0, max), max: max}
	heap.Init(&h.heap)
	return h
}

func (h *MinHeap) Add(job Job) error {
	if h.max > 0 && len(h.heap) >= h.max {
		return ErrMax
	}
	heap.Push(&h.heap, job)
	return nil
}

func (h *MinHeap) Pop() Job {
	if len(h.heap) == 0 {
		return Job{}
	}
	return heap.Pop(&h.heap).(Job)
}

func (h *MinHeap) Peek() Job {
	if len(h.heap) == 0 {
		return Job{}
	}
	return h.heap[0]
}

func (h *MinHeap) Size() int {
	return len(h.heap)
}
