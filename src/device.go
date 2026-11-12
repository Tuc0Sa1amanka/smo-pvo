package main

import "math/rand"

// ComputingModule — ЭВМ с равномерным временем обслуживания (ПЗ2).
type ComputingModule struct {
	id         int
	busy       bool
	current    *Target
	serviceEnd float64
	busyTime   float64
	minService float64
	maxService float64
	rng        *rand.Rand
}

func NewComputingModule(id int, minService, maxService float64, rng *rand.Rand) *ComputingModule {
	return &ComputingModule{
		id:         id,
		minService: minService,
		maxService: maxService,
		rng:        rng,
	}
}

func (m *ComputingModule) ID() int             { return m.id }
func (m *ComputingModule) IsBusy() bool        { return m.busy }
func (m *ComputingModule) ServiceEnd() float64 { return m.serviceEnd }
func (m *ComputingModule) BusyTime() float64   { return m.busyTime }
func (m *ComputingModule) Current() *Target    { return m.current }

// Start ставит цель на обслуживание и возвращает момент окончания (для календаря).
func (m *ComputingModule) Start(t *Target, now float64) float64 {
	duration := m.minService + m.rng.Float64()*(m.maxService-m.minService)
	m.busy = true
	m.current = t
	m.serviceEnd = now + duration
	t.MarkServiceStart(now)
	return m.serviceEnd
}

// Finish завершает обслуживание, учитывает занятость и возвращает обслуженную цель.
func (m *ComputingModule) Finish(now float64) *Target {
	if !m.busy || m.current == nil {
		return nil
	}
	t := m.current
	t.MarkServiceEnd(now)
	m.busyTime += now - t.serviceStart
	m.busy = false
	m.current = nil
	m.serviceEnd = 0
	return t
}
