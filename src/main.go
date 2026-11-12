package main

import "fmt"

// Config — входные параметры пошаговой имитации.
type Config struct {
	SourceCount    int
	Lambdas        []float64
	DeviceCount    int
	MinService     float64
	MaxService     float64
	BufferCapacity int
	StepLimit      int
	Seed           int64
}

func main() {
	cfg := Config{
		SourceCount:    3,
		Lambdas:        []float64{0.5, 0.3, 0.1},
		DeviceCount:    2,
		MinService:     3.0,
		MaxService:     10.0,
		BufferCapacity: 4,
		StepLimit:      20,
		Seed:           42,
	}

	if len(cfg.Lambdas) != cfg.SourceCount {
		panic("длина Lambdas должна совпадать с SourceCount")
	}

	fmt.Println("Имитационная модель: Вычислительный центр ПВО")
	fmt.Println("Вариант: ИБ ИЗ1 ПЗ2 Д10З2 Д10О5 Д2П1 Д2Б5 ОР1 ОД3")
	fmt.Println("Режим: пошаговый (ОД3), метод особых состояний")
	fmt.Println("После каждого шага нажмите Enter, чтобы продолжить.")
	fmt.Println()

	sim := NewSimulation(cfg)
	sim.Run()
}
