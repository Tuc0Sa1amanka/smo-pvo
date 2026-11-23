package main

// scheduleFunc планирует событие «прибор освободился» в календаре симуляции.
type scheduleFunc func(deviceID int, at float64)

// IntakeDispatcher — постановка: первый свободный прибор (Д2П1), иначе FIFO/отказ (Д10З2/Д10О5).
type IntakeDispatcher struct {
	zone     *TrackingZone
	devices  []*ComputingModule
	schedule scheduleFunc
	stats    *Statistics
}

func NewIntakeDispatcher(zone *TrackingZone, devices []*ComputingModule, schedule scheduleFunc, stats *Statistics) *IntakeDispatcher {
	return &IntakeDispatcher{
		zone:     zone,
		devices:  devices,
		schedule: schedule,
		stats:    stats,
	}
}

// OnTargetArrived обрабатывает прибытие цели. true — принята (прибор или буфер), false — отказ.
func (d *IntakeDispatcher) OnTargetArrived(t *Target, now float64) bool {
	if device := d.findFreeDevice(); device != nil {
		end := device.Start(t, now)
		d.schedule(device.ID(), end)
		return true
	}
	if d.zone.Put(t) {
		return true
	}
	// Д10О5: отказ вновь пришедшей заявке.
	d.stats.AddRefused(t.SectorID())
	return false
}

func (d *IntakeDispatcher) findFreeDevice() *ComputingModule {
	for _, device := range d.devices {
		if !device.IsBusy() {
			return device
		}
	}
	return nil
}

// SelectionDispatcher — выбор из пакета / формирование пакета по приоритету сектора (Д2Б5).
type SelectionDispatcher struct {
	zone     *TrackingZone
	sources  []*TargetGenerator
	batch    *SectorBatch
	schedule scheduleFunc
}

func NewSelectionDispatcher(zone *TrackingZone, sources []*TargetGenerator, schedule scheduleFunc) *SelectionDispatcher {
	return &SelectionDispatcher{
		zone:     zone,
		sources:  sources,
		schedule: schedule,
	}
}

func (d *SelectionDispatcher) Batch() *SectorBatch { return d.batch }

// OnDeviceFreed ставит на освободившийся прибор следующую цель из пакета/буфера.
// true — прибор снова занят, false — простаивает.
func (d *SelectionDispatcher) OnDeviceFreed(device *ComputingModule, now float64) bool {
	target, ok := d.takeNextTarget()
	if !ok {
		return false
	}
	end := device.Start(target, now)
	d.schedule(device.ID(), end)
	return true
}

func (d *SelectionDispatcher) takeNextTarget() (*Target, bool) {
	if d.batch != nil && !d.batch.IsEmpty() {
		return d.batch.Pop()
	}
	d.batch = nil

	if d.zone.IsEmpty() {
		return nil, false
	}

	sectorID, ok := d.selectSector()
	if !ok {
		return nil, false
	}

	// Пакет фиксируется в момент выбора: все цели сектора уходят из буфера.
	d.batch = d.zone.TakeSector(sectorID)
	return d.batch.Pop()
}

// selectSector выбирает сектор с минимальным ID среди присутствующих в буфере (приоритет источника).
func (d *SelectionDispatcher) selectSector() (int, bool) {
	sectors := d.zone.Sectors()
	if len(sectors) == 0 {
		return 0, false
	}
	best := sectors[0]
	for _, id := range sectors[1:] {
		if id < best {
			best = id
		}
	}
	return best, true
}
