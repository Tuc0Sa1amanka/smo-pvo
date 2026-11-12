package main

import "container/heap"

// Типы особых событий метода особых состояний.
const (
	EventTargetArrived = iota
	EventDeviceFreed
)

// Event — запись календаря событий.
type Event struct {
	Time     float64
	Type     int
	SourceID int // для TargetArrived
	DeviceID int // для DeviceFreed
	seq      int // стабильный порядок при равном времени
}

type eventHeap []*Event

func (h eventHeap) Len() int { return len(h) }

func (h eventHeap) Less(i, j int) bool {
	if h[i].Time != h[j].Time {
		return h[i].Time < h[j].Time
	}
	// При совпадении времени сначала освобождаем прибор — меньше ложных отказов.
	if h[i].Type != h[j].Type {
		return h[i].Type > h[j].Type // DeviceFreed > TargetArrived
	}
	return h[i].seq < h[j].seq
}

func (h eventHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *eventHeap) Push(x any) {
	*h = append(*h, x.(*Event))
}

func (h *eventHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}

// Calendar — приоритетная очередь особых событий.
type Calendar struct {
	h   eventHeap
	seq int
}

func NewCalendar() *Calendar {
	c := &Calendar{h: make(eventHeap, 0)}
	heap.Init(&c.h)
	return c
}

func (c *Calendar) Push(e *Event) {
	c.seq++
	e.seq = c.seq
	heap.Push(&c.h, e)
}

func (c *Calendar) Pop() *Event {
	if c.h.Len() == 0 {
		return nil
	}
	return heap.Pop(&c.h).(*Event)
}

func (c *Calendar) Len() int { return c.h.Len() }
