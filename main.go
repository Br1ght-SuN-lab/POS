package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/dig"

	"pos/internal/application"
	"pos/internal/infrastructure"
)

func main() {
	dataPath := flag.String("data", "pos.json", "файл ведомости JSON")
	scriptPath := flag.String("script", "", "файл команд для сквозного прогона")
	flag.Parse()

	if err := run(*dataPath, *scriptPath); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run(dataPath, scriptPath string) error {
	if !strings.EqualFold(filepath.Ext(dataPath), ".json") {
		return fmt.Errorf("файл ведомости должен иметь расширение .json")
	}
	var input io.Reader = os.Stdin
	if scriptPath != "" {
		file, err := os.Open(scriptPath)
		if err != nil {
			return err
		}
		defer file.Close()
		input = file
	}

	// Собираем зависимости приложения на входе в программу.
	container := dig.New()
	providers := []any{
		func() application.Repository { return infrastructure.NewJSONStore(dataPath) },
		func() application.ReportExporter { return infrastructure.CSVExporter{} },
		application.NewService,
		func(service *application.Service) *infrastructure.Console {
			return infrastructure.NewConsole(input, os.Stdout, service)
		},
	}
	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			return err
		}
	}
	return container.Invoke(func(console *infrastructure.Console) error {
		return console.Run(scriptPath != "")
	})
}
