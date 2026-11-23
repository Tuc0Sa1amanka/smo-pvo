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

// SourceStat — показатели одного сектора.
type SourceStat struct {
	ID           int
	Generated    int
	Refused      int
	Served       int
	SumWait      float64
	SumWaitSq    float64
	SumService   float64
	SumResidence float64
}

// DeviceStat — показатели одного прибора.
type DeviceStat struct {
	ID       int
	Served   int
	BusyTime float64
}

// Statistics собирает показатели пошагового отчёта и сводной таблицы.
type Statistics struct {
	Generated      int
	Served         int
	Refused        int
	TotalWait      float64
	TotalWaitSq    float64
	TotalService   float64
	TotalResidence float64
	TotalBusy      float64
	TotalTime      float64
	Sources        []SourceStat
	Devices        []DeviceStat
}

func NewStatistics(sourceCount, deviceCount int) *Statistics {
	s := &Statistics{
		Sources: make([]SourceStat, sourceCount),
		Devices: make([]DeviceStat, deviceCount),
	}
	for i := range s.Sources {
		s.Sources[i].ID = i + 1
	}
	for i := range s.Devices {
		s.Devices[i].ID = i + 1
	}
	return s
}

func (s *Statistics) AddGenerated(sectorID int) {
	s.Generated++
	s.Sources[sectorID-1].Generated++
}

func (s *Statistics) AddRefused(sectorID int) {
	s.Refused++
	s.Sources[sectorID-1].Refused++
}

func (s *Statistics) AddServed(t *Target, deviceID int) {
	wait := t.WaitTime()
	service := t.ServiceTime()
	residence := t.ResidenceTime()

	s.Served++
	s.TotalWait += wait
	s.TotalWaitSq += wait * wait
	s.TotalService += service
	s.TotalResidence += residence

	src := &s.Sources[t.SectorID()-1]
	src.Served++
	src.SumWait += wait
	src.SumWaitSq += wait * wait
	src.SumService += service
	src.SumResidence += residence

	s.Devices[deviceID-1].Served++
}

func (s *Statistics) AvgWait() float64 {
	return avg(s.TotalWait, s.Served)
}

func (s *Statistics) AvgService() float64 {
	return avg(s.TotalService, s.Served)
}

func (s *Statistics) AvgResidence() float64 {
	return avg(s.TotalResidence, s.Served)
}

func (s *Statistics) WaitVariance() float64 {
	return variance(s.TotalWait, s.TotalWaitSq, s.Served)
}

func (s *Statistics) RefuseProb() float64 {
	if s.Generated == 0 {
		return 0
	}
	return float64(s.Refused) / float64(s.Generated)
}

// Utilization — средняя загрузка приборов: суммарная занятость / (N * T).
func (s *Statistics) Utilization(deviceCount int) float64 {
	if s.TotalTime <= 0 || deviceCount <= 0 {
		return 0
	}
	return s.TotalBusy / (float64(deviceCount) * s.TotalTime)
}

func (s SourceStat) RefuseProb() float64 {
	if s.Generated == 0 {
		return 0
	}
	return float64(s.Refused) / float64(s.Generated)
}

func (s SourceStat) AvgWait() float64      { return avg(s.SumWait, s.Served) }
func (s SourceStat) AvgService() float64   { return avg(s.SumService, s.Served) }
func (s SourceStat) AvgResidence() float64 { return avg(s.SumResidence, s.Served) }
func (s SourceStat) WaitVariance() float64 {
	return variance(s.SumWait, s.SumWaitSq, s.Served)
}

func (d DeviceStat) Utilization(totalTime float64) float64 {
	if totalTime <= 0 {
		return 0
	}
	return d.BusyTime / totalTime
}

func avg(sum float64, n int) float64 {
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func variance(sum, sumSq float64, n int) float64 {
	if n == 0 {
		return 0
	}
	mean := sum / float64(n)
	v := sumSq/float64(n) - mean*mean
	if v < 0 {
		return 0
	}
	return v
}
