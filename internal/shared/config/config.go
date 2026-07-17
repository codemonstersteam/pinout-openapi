// Package config загружает CLI-конфиг инструмента из файла (config.yaml) с
// override из env. В коде НЕТ хардкодов — всё приходит отсюда. Fail-fast на
// невалидном конфиге. Это конфиг САМОГО инструмента (не путать с доменным
// config-аргументом команды run). Добавляй поля по мере роста CLI.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config — конфигурация инструмента.
type Config struct {
	// OutputFormat — формат отчёта в stdout. Пока поддержан только "json".
	OutputFormat string `yaml:"output_format"`
}

// Load читает config.yaml (путь из CONFIG_FILE, дефолт "config.yaml") и применяет
// env-override (OUTPUT_FORMAT). Отсутствие файла — не ошибка (дефолты + env).
// Невалидный YAML или пустой формат → ошибка (fail-fast).
func Load() (Config, error) {
	cfg := Config{OutputFormat: "json"} // дефолт

	path := getenv("CONFIG_FILE", "config.yaml")
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	if f := os.Getenv("OUTPUT_FORMAT"); f != "" {
		cfg.OutputFormat = f
	}
	if cfg.OutputFormat == "" {
		return Config{}, fmt.Errorf("output_format пуст")
	}
	return cfg, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
