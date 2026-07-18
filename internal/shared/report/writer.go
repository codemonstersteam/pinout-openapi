// Package report — writer-скелет отчёта инструмента. Пишет произвольное значение
// как отступленный JSON в поток (обычно stdout). Generic: доменных типов не знает,
// принимает any — так его переиспользуют любые слайсы. Тикеты добавляют форматы/поля.
package report

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON сериализует v в JSON (2-пробельный отступ) и пишет в w с переводом строки.
func WriteJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}
