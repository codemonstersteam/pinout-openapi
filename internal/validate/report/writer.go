// Package report — слайс slice-01-validate: форма Report DTO (report.schema.json) и её
// персистентность. Этот файл — ReportWriter.Write (ticket 15, contracts.md
// §ReportWriter.Write): чистая I/O-труба записи JSON-отчёта на диск (ADR-0003 —
// filesystem loaders тегированы io: none, юнит-тестов нет).
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"pinout-openapi/internal/validate/domain"
)

// ReportWriter — порт персистентности отчёта. Интерфейс, чтобы ProcessValidate (head)
// и его юнит-тесты подменяли реализацию фейком (contracts.md Deps{..., ReportWriter, ...}).
type ReportWriter interface {
	// Write иff settings.SaveJSONReport атомарно пишет r как JSON в
	// settings.JSONReportFile; в любом случае — ROP pass-through того же Report
	// (никогда не мутирует частично уже существующий отчёт). Ошибка окружения —
	// не сентинел слайса (нет error.code в контракте), просто оборачивается и
	// поднимается наверх (main логирует в stderr и выходит с кодом 3).
	Write(s domain.Settings, r domain.Report) (domain.Report, error)
}

// fileReportWriter — файловая реализация ReportWriter (encapsulates the OS filesystem).
type fileReportWriter struct{}

// NewFileReportWriter — конструктор боевого порта.
func NewFileReportWriter() ReportWriter { return fileReportWriter{} }

// Write — см. ReportWriter.Write. Атомарность: пишем во временный файл в той же
// директории, что и целевой путь, затем os.Rename (POSIX rename — atomic replace),
// так что читатель никогда не увидит частично записанный отчёт.
func (fileReportWriter) Write(s domain.Settings, r domain.Report) (domain.Report, error) {
	if !s.SaveJSONReport {
		return r, nil
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return r, fmt.Errorf("report: marshal: %w", err)
	}

	dir := filepath.Dir(s.JSONReportFile)
	tmp, err := os.CreateTemp(dir, ".report-*.json.tmp")
	if err != nil {
		return r, fmt.Errorf("report: create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return r, fmt.Errorf("report: write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return r, fmt.Errorf("report: close temp file: %w", err)
	}

	if err := os.Rename(tmpName, s.JSONReportFile); err != nil {
		_ = os.Remove(tmpName)
		return r, fmt.Errorf("report: rename temp file: %w", err)
	}

	return r, nil
}
