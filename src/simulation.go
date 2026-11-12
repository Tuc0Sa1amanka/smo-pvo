package main

import (
	"fmt"
	"math/rand"
)

// Simulation — ядро метода особых состояний.
type Simulation struct {
	cfg        Config
	now        float64
	calendar   *Calendar
	sources    []*TargetGenerator
	devices    []*ComputingModule
	zone       *TrackingZone
	intake     *IntakeDispatcher
	selection  *SelectionDispatcher
	stats      *Statistics
	nextID     int
	stopReason string
}

func NewSimulation(cfg Config) *Simulation {
	stats := &Statistics{}
	zone := NewTrackingZone(cfg.BufferCapacity)

	s := &Simulation{
		cfg:      cfg,
		calendar: NewCalendar(),
		zone:     zone,
		stats:    stats,
		nextID:   1,
	}

	nextTarget := func() int {
		id := s.nextID
		s.nextID++
		return id
	}

	s.sources = make([]*TargetGenerator, cfg.SourceCount)
	for i := 0; i < cfg.SourceCount; i++ {
		lambda := cfg.Lambdas[i]
		// Отдельный детерминированный поток на источник — без корреляции интервалов.
		srcRNG := rand.New(rand.NewSource(cfg.Seed + int64(i)*1009 + 1))
		s.sources[i] = NewTargetGenerator(i+1, lambda, srcRNG, nextTarget)
	}

	s.devices = make([]*ComputingModule, cfg.DeviceCount)
	for i := 0; i < cfg.DeviceCount; i++ {
		devRNG := rand.New(rand.NewSource(cfg.Seed + int64(i)*1009 + 5000))
		s.devices[i] = NewComputingModule(i+1, cfg.MinService, cfg.MaxService, devRNG)
	}

	schedule := func(deviceID int, at float64) {
		s.calendar.Push(&Event{
			Time:     at,
			Type:     EventDeviceFreed,
			DeviceID: deviceID,
		})
	}

	s.intake = NewIntakeDispatcher(zone, s.devices, schedule, stats)
	s.selection = NewSelectionDispatcher(zone, s.sources, schedule)

	for _, src := range s.sources {
		s.calendar.Push(&Event{
			Time:     src.NextArrivalTime(),
			Type:     EventTargetArrived,
			SourceID: src.SectorID(),
		})
	}
	return s
}

func (s *Simulation) Run() {
	step := 0
	for s.calendar.Len() > 0 {
		if s.stats.Generated >= s.cfg.StepLimit {
			s.stopReason = "limit"
			break
		}

		ev := s.calendar.Pop()
		if ev == nil {
			break
		}
		s.now = ev.Time
		step++

		var description string
		switch ev.Type {
		case EventTargetArrived:
			description = s.handleArrival(ev)
		case EventDeviceFreed:
			description = s.handleDeviceFreed(ev)
		default:
			description = "неизвестное событие"
		}

		printStep(step, s.now, description, s)

		if s.stats.Generated >= s.cfg.StepLimit {
			s.stopReason = "limit"
			break
		}
		waitNextStep()
	}

	s.finalizeStats()
	printSummary(s.stats, len(s.devices))
}

func (s *Simulation) handleArrival(ev *Event) string {
	src := s.sourceByID(ev.SourceID)
	if src == nil {
		return "ошибка: неизвестный источник"
	}

	target := src.Generate(s.now)
	s.stats.Generated++

	accepted := s.intake.OnTargetArrived(target, s.now)

	// Планируем следующее прибытие от этого сектора.
	s.calendar.Push(&Event{
		Time:     src.NextArrivalTime(),
		Type:     EventTargetArrived,
		SourceID: src.SectorID(),
	})

	if accepted {
		return fmt.Sprintf("Цель %d от Сектора %d прилетела.", target.ID(), target.SectorID())
	}
	return fmt.Sprintf("Цель %d от Сектора %d прилетела и получила отказ (буфер полон).", target.ID(), target.SectorID())
}

func (s *Simulation) handleDeviceFreed(ev *Event) string {
	device := s.deviceByID(ev.DeviceID)
	if device == nil {
		return "ошибка: неизвестный прибор"
	}

	finished := device.Finish(s.now)
	desc := fmt.Sprintf("Прибор П%d освободился.", device.ID())
	if finished != nil {
		s.stats.Served++
		s.stats.TotalWait += finished.WaitTime()
		desc = fmt.Sprintf("Прибор П%d освободился (обслужена Цель %d).", device.ID(), finished.ID())
	}

	if s.selection.OnDeviceFreed(device, s.now) {
		cur := device.Current()
		if cur != nil {
			desc += fmt.Sprintf(" На него поставлена Цель %d.", cur.ID())
		}
	} else {
		desc += " Прибор простаивает."
	}
	return desc
}

func (s *Simulation) finalizeStats() {
	s.stats.TotalTime = s.now
	var busy float64
	for _, d := range s.devices {
		busy += d.BusyTime()
		// Учитываем текущую незавершённую занятость на момент остановки.
		if d.IsBusy() && d.Current() != nil {
			busy += s.now - d.Current().serviceStart
		}
	}
	s.stats.TotalBusy = busy
}

func (s *Simulation) sourceByID(id int) *TargetGenerator {
	for _, src := range s.sources {
		if src.SectorID() == id {
			return src
		}
	}
	return nil
}

func (s *Simulation) deviceByID(id int) *ComputingModule {
	for _, d := range s.devices {
		if d.ID() == id {
			return d
		}
	}
	return nil
}
