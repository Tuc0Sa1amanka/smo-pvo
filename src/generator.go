package main

import "math/rand"

// TargetGenerator — сектор наблюдения с бесконечным пуассоновским потоком (ИБ, ИЗ1).
type TargetGenerator struct {
	id          int
	lambda      float64
	nextArrival float64
	generated   int
	rng         *rand.Rand
	nextTarget  func() int
}

func NewTargetGenerator(id int, lambda float64, rng *rand.Rand, nextTarget func() int) *TargetGenerator {
	g := &TargetGenerator{
		id:         id,
		lambda:     lambda,
		rng:        rng,
		nextTarget: nextTarget,
	}
	g.scheduleNext(0)
	return g
}

func (g *TargetGenerator) SectorID() int          { return g.id }
func (g *TargetGenerator) NextArrivalTime() float64 { return g.nextArrival }
func (g *TargetGenerator) Generated() int         { return g.generated }

// Generate создаёт цель в момент now и планирует следующее прибытие.
func (g *TargetGenerator) Generate(now float64) *Target {
	t := &Target{
		id:          g.nextTarget(),
		sectorID:    g.id,
		arrivalTime: now,
	}
	g.generated++
	g.scheduleNext(now)
	return t
}

func (g *TargetGenerator) scheduleNext(now float64) {
	// ИЗ1: экспоненциальный интервал, λ — интенсивность.
	interval := g.rng.ExpFloat64() / g.lambda
	g.nextArrival = now + interval
}
