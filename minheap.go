package htask

import (
	"container/heap"
	"errors"
)

type jobHeap []job

// Len is length of jobHeap
func (h *jobHeap) Len() int {
	return len(*h)
}

// Less means job j is newer than i
func (h *jobHeap) Less(i, j int) bool {
	return !(*h)[i].t.After((*h)[j].t)
}

// Swap swaps the elements with indexes i and j.
func (h *jobHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

// Push adds x to tail
func (h *jobHeap) Push(x interface{}) {
	*h = append(*h, x.(job))
}

// Pop removes x from head
func (h *jobHeap) Pop() (x interface{}) {
	x, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]
	return
}

// errors
var (
	ErrMax = errors.New("heap max size")
)

type minHeap struct {
	heap jobHeap
	max  int
}

func newMinHeap(max int) *minHeap {
	if max < 0 {
		max = 0
	}
	h := &minHeap{heap: make(jobHeap, 0, max), max: max}
	heap.Init(&h.heap)
	return h
}

func (h *minHeap) add(j job) error {
	if h.max > 0 && len(h.heap) >= h.max {
		return ErrMax
	}
	heap.Push(&h.heap, j)
	return nil
}

func (h *minHeap) pop() job {
	if len(h.heap) == 0 {
		return job{}
	}
	return heap.Pop(&h.heap).(job)
}

func (h *minHeap) peek() job {
	if len(h.heap) == 0 {
		return job{}
	}
	return h.heap[0]
}

func (h *minHeap) size() int {
	return len(h.heap)
}
