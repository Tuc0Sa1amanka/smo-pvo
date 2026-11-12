package main

// SectorBatch — пакет целей одного сектора (Д2Б5). Пакет не расширяется.
type SectorBatch struct {
	sectorID int
	targets  []*Target
}

func NewSectorBatch(sectorID int) *SectorBatch {
	return &SectorBatch{sectorID: sectorID, targets: make([]*Target, 0)}
}

func (b *SectorBatch) add(t *Target) {
	b.targets = append(b.targets, t)
}

func (b *SectorBatch) Pop() (*Target, bool) {
	if len(b.targets) == 0 {
		return nil, false
	}
	t := b.targets[0]
	b.targets = b.targets[1:]
	return t, true
}

func (b *SectorBatch) IsEmpty() bool { return len(b.targets) == 0 }
func (b *SectorBatch) Size() int     { return len(b.targets) }
func (b *SectorBatch) SectorID() int { return b.sectorID }

func (b *SectorBatch) Snapshot() []*Target {
	out := make([]*Target, len(b.targets))
	copy(out, b.targets)
	return out
}

// TrackingZone — общая FIFO-зона сопровождения (Д10З2).
type TrackingZone struct {
	queue    []*Target
	capacity int
}

func NewTrackingZone(capacity int) *TrackingZone {
	return &TrackingZone{
		queue:    make([]*Target, 0, capacity),
		capacity: capacity,
	}
}

func (z *TrackingZone) Put(t *Target) bool {
	if z.IsFull() {
		return false
	}
	z.queue = append(z.queue, t)
	return true
}

// TakeSector физически извлекает из буфера все цели сектора (порядок FIFO внутри пакета).
func (z *TrackingZone) TakeSector(sectorID int) *SectorBatch {
	batch := NewSectorBatch(sectorID)
	rest := make([]*Target, 0, len(z.queue))
	for _, t := range z.queue {
		if t.SectorID() == sectorID {
			batch.add(t)
		} else {
			rest = append(rest, t)
		}
	}
	z.queue = rest
	return batch
}

// ExtractBatch — алиас к TakeSector (имя из ТЗ пользователя).
func (z *TrackingZone) ExtractBatch(sectorID int) *SectorBatch {
	return z.TakeSector(sectorID)
}

func (z *TrackingZone) Sectors() []int {
	seen := make(map[int]struct{})
	var ids []int
	for _, t := range z.queue {
		id := t.SectorID()
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return ids
}

func (z *TrackingZone) Size() int    { return len(z.queue) }
func (z *TrackingZone) IsFull() bool { return len(z.queue) >= z.capacity }
func (z *TrackingZone) IsEmpty() bool {
	return len(z.queue) == 0
}
func (z *TrackingZone) Capacity() int { return z.capacity }

func (z *TrackingZone) Snapshot() []*Target {
	out := make([]*Target, len(z.queue))
	copy(out, z.queue)
	return out
}
