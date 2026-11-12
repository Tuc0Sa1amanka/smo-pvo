package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func waitNextStep() {
	fmt.Print("Enter — следующий шаг... ")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	fmt.Println()
}

func printStep(step int, now float64, description string, sim *Simulation) {
	fmt.Printf("[Шаг %d] Время: %.2f\n", step, now)
	fmt.Printf("Событие: %s\n", description)

	parts := make([]string, 0, len(sim.sources))
	for _, src := range sim.sources {
		parts = append(parts, fmt.Sprintf("И%d: %d", src.SectorID(), src.Generated()))
	}
	fmt.Printf("Источники: %s\n", strings.Join(parts, " | "))

	buf := sim.zone.Snapshot()
	if len(buf) == 0 {
		fmt.Printf("Буфер: [] (занято 0 из %d)\n", sim.zone.Capacity())
	} else {
		items := make([]string, 0, len(buf))
		for _, t := range buf {
			items = append(items, fmt.Sprintf("Цель %d (С%d)", t.ID(), t.SectorID()))
		}
		fmt.Printf("Буфер: [%s] (занято %d из %d)\n",
			strings.Join(items, ", "), len(buf), sim.zone.Capacity())
	}

	for _, d := range sim.devices {
		if !d.IsBusy() || d.Current() == nil {
			fmt.Printf("П%d: Свободен\n", d.ID())
			continue
		}
		left := d.ServiceEnd() - now
		if left < 0 {
			left = 0
		}
		cur := d.Current()
		fmt.Printf("П%d: Занят (Цель %d, осталось %.1f сек)\n", d.ID(), cur.ID(), left)
	}

	batch := sim.selection.Batch()
	if batch == nil || batch.IsEmpty() {
		fmt.Println("Активный пакет: Нет")
	} else {
		items := make([]string, 0, batch.Size())
		for _, t := range batch.Snapshot() {
			items = append(items, fmt.Sprintf("Цель %d", t.ID()))
		}
		fmt.Printf("Активный пакет: С%d [%s] (осталось %d)\n",
			batch.SectorID(), strings.Join(items, ", "), batch.Size())
	}
	fmt.Println()
}

func printSummary(stats *Statistics, deviceCount int) {
	fmt.Println("(достигнут лимит)")
	fmt.Println("СИМУЛЯЦИЯ ОСТАНОВЛЕНА.")
	fmt.Printf(
		"Сводка: Сгенерировано: %d | Отказов: %d | Ср. время ожидания: %.1f | Загрузка приборов: %.0f%%\n",
		stats.Generated,
		stats.Refused,
		stats.AvgWait(),
		stats.Utilization(deviceCount)*100,
	)
}
