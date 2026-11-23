package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
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

func chooseMode() RunMode {
	if len(os.Args) > 1 {
		switch strings.ToLower(strings.TrimSpace(os.Args[1])) {
		case "auto", "1", "a":
			return ModeAuto
		case "step", "2", "s":
			return ModeStep
		default:
			fmt.Printf("Неизвестный режим %q. Доступны: auto, step.\n", os.Args[1])
			os.Exit(1)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Режим:")
		fmt.Println("  1 — автоматический (ОР1), сводная таблица")
		fmt.Println("  2 — пошаговый (ОД3)")
		fmt.Print("Выбор: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			fmt.Println("Ввод закрыт, запускаю пошаговый режим.")
			return ModeStep
		}
		switch strings.TrimSpace(line) {
		case "1":
			return ModeAuto
		case "2":
			return ModeStep
		default:
			fmt.Println("Введите 1 или 2.")
		}
	}
}

func printAutoReport(sim *Simulation) {
	stats := sim.stats
	cfg := sim.cfg

	fmt.Println("Сводная таблица (ОР1)")
	fmt.Printf("Источников: %d, λ: %s\n", cfg.SourceCount, formatList(cfg.Lambdas))
	fmt.Printf("Приборов: %d, обслуживание: [%.2f; %.2f], буфер: %d\n",
		cfg.DeviceCount, cfg.MinService, cfg.MaxService, cfg.BufferCapacity)
	fmt.Printf("Заявок: %d, модельное время: %.2f, seed: %d\n",
		stats.Generated, stats.TotalTime, cfg.Seed)
	fmt.Println("Генерация остановлена после лимита, принятые цели дообслужены.")
	fmt.Println()
	fmt.Println("Pотк — вероятность отказа, Tож — среднее ожидание, Tобсл — среднее обслуживание,")
	fmt.Println("Tпреб — среднее пребывание, Dож — дисперсия ожидания, Kисп — загрузка прибора.")
	fmt.Println()

	srcWidths := []int{8, 8, 8, 8, 8, 10, 10, 10, 10}
	fmt.Println("Источники")
	fmt.Println(tableRow([]string{
		"Сектор", "Сгенер.", "Отказов", "Pотк", "Обслуж.", "Tож", "Tобсл", "Tпреб", "Dож",
	}, srcWidths))
	for _, src := range stats.Sources {
		fmt.Println(tableRow([]string{
			fmt.Sprintf("С%d", src.ID),
			fmt.Sprintf("%d", src.Generated),
			fmt.Sprintf("%d", src.Refused),
			pct(src.RefuseProb()),
			fmt.Sprintf("%d", src.Served),
			num(src.AvgWait()),
			num(src.AvgService()),
			num(src.AvgResidence()),
			num(src.WaitVariance()),
		}, srcWidths))
	}
	fmt.Println(tableRow([]string{
		"Итого",
		fmt.Sprintf("%d", stats.Generated),
		fmt.Sprintf("%d", stats.Refused),
		pct(stats.RefuseProb()),
		fmt.Sprintf("%d", stats.Served),
		num(stats.AvgWait()),
		num(stats.AvgService()),
		num(stats.AvgResidence()),
		num(stats.WaitVariance()),
	}, srcWidths))

	fmt.Println()
	devWidths := []int{8, 10, 18, 8}
	fmt.Println("Приборы")
	fmt.Println(tableRow([]string{"Прибор", "Обслужено", "Время занятости", "Kисп"}, devWidths))
	for _, dev := range stats.Devices {
		fmt.Println(tableRow([]string{
			fmt.Sprintf("П%d", dev.ID),
			fmt.Sprintf("%d", dev.Served),
			num(dev.BusyTime),
			pct(dev.Utilization(stats.TotalTime)),
		}, devWidths))
	}
	fmt.Println(tableRow([]string{
		"Среднее", "", "", pct(stats.Utilization(len(stats.Devices))),
	}, devWidths))

	fmt.Println()
	fmt.Println("Система")
	printKV("Вероятность отказа", pct(stats.RefuseProb()))
	printKV("Среднее время ожидания", num(stats.AvgWait()))
	printKV("Среднее время обслуживания", num(stats.AvgService()))
	printKV("Среднее время пребывания", num(stats.AvgResidence()))
	printKV("Дисперсия времени ожидания", num(stats.WaitVariance()))
	printKV("Средняя загрузка приборов", pct(stats.Utilization(len(stats.Devices))))
	printKV("Осталось в системе", fmt.Sprintf("%d", stats.Generated-stats.Refused-stats.Served))
}

func printKV(name, value string) {
	fmt.Printf("  %s %s\n", padRight(name, 32), value)
}

func tableRow(cols []string, widths []int) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = padLeft(col, widths[i])
	}
	return strings.Join(parts, "  ")
}

func padLeft(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return strings.Repeat(" ", width-n) + s
}

func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func pct(v float64) string { return fmt.Sprintf("%.2f%%", v*100) }
func num(v float64) string { return fmt.Sprintf("%.2f", v) }

func formatList(values []float64) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("%.2f", v)
	}
	return strings.Join(parts, ", ")
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
