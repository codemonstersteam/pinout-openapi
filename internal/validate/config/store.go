// Package config — слайс slice-01-validate: чтение и валидация config-файла
// (contracts.md §ConfigStore.Load, §NewConfig; module-tree.md "config").
package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"pinout-openapi/internal/validate/domain"
)

// ConfigStore — I/O-труба чтения config-файла (io: none — ADR-0003, локальная ФС не
// маршрутизируется на io-сабскилл). Секрет: формат файла (YAML) и то, что файловая
// система инкапсулирована в объекте — вызывающий видит только Load.
type ConfigStore struct{}

// NewConfigStore — конструктор трубы (без зависимостей — ФС инкапсулирована внутри).
func NewConfigStore() ConfigStore {
	return ConfigStore{}
}

// Load — contracts.md §ConfigStore.Load: Load(path) -> Result[RawConfig, Error].
// Антецедент: path непустой. Следствие: Ok — структурно раскодированный (ещё НЕ
// провалидированный) RawConfig; Fail — ErrConfig (не найден / нечитаем / битый YAML),
// error.code=CONFIG_ERROR, exit 2. Валидация полей — не здесь (это NewConfig).
func (ConfigStore) Load(path string) (domain.RawConfig, error) {
	if path == "" {
		return domain.RawConfig{}, domain.ErrConfig
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return domain.RawConfig{}, domain.ErrConfig
	}

	var raw domain.RawConfig
	if err := yaml.Unmarshal(bytes, &raw); err != nil {
		return domain.RawConfig{}, domain.ErrConfig
	}

	return raw, nil
}
