package main

// Target — обнаруженная цель (заявка СМО).
type Target struct {
	id           int
	sectorID     int
	arrivalTime  float64
	serviceStart float64
	serviceEnd   float64
}

func (t *Target) ID() int       { return t.id }
func (t *Target) SectorID() int { return t.sectorID }

func (t *Target) MarkServiceStart(now float64) { t.serviceStart = now }
func (t *Target) MarkServiceEnd(now float64)   { t.serviceEnd = now }

func (t *Target) WaitTime() float64 {
	if t.serviceStart < t.arrivalTime {
		return 0
	}
	return t.serviceStart - t.arrivalTime
}

func (t *Target) ServiceTime() float64 {
	if t.serviceEnd < t.serviceStart {
		return 0
	}
	return t.serviceEnd - t.serviceStart
}

func (t *Target) ResidenceTime() float64 {
	if t.serviceEnd < t.arrivalTime {
		return 0
	}
	return t.serviceEnd - t.arrivalTime
}

// Statistics собирает показатели пошагового/итогового отчёта.
type Statistics struct {
	Generated   int
	Served      int
	Refused     int
	TotalWait   float64
	TotalBusy   float64
	TotalTime   float64
}

func (s *Statistics) AvgWait() float64 {
	if s.Served == 0 {
		return 0
	}
	return s.TotalWait / float64(s.Served)
}

// Utilization — средняя загрузка приборов: суммарная занятость / (N * T).
func (s *Statistics) Utilization(deviceCount int) float64 {
	if s.TotalTime <= 0 || deviceCount <= 0 {
		return 0
	}
	return s.TotalBusy / (float64(deviceCount) * s.TotalTime)
}
