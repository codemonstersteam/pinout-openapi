package steps

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cucumber/godog"
)

// registerCompatCheckSteps — доменные степы для слайса compat-check (CS1-CS9,
// ticket-02). Общие CLI-степы (запуск бинаря, код выхода, подстрока в выводе) уже
// есть в cli_steps.go — здесь только то, чего не хватает: выбор фикстуры,
// проверка JSON-полей отчёта на stdout, детерминизм файла отчёта (Gate #1 A4),
// неизменность provider-спеки (CS9). Все запускаются против /fixtures (том,
// смонтированный в docker-compose.test.yml из component-tests/fixtures).
func (w *World) registerCompatCheckSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^используем фикстуру "([^"]*)"$`, w.useFixture)
	ctx.Step(`^запускаем инструмент с этим конфигом$`, w.runWithFixtureConfig)
	ctx.Step(`^JSON-поле "([^"]*)" в stdout равно "([^"]*)"$`, w.jsonFieldEquals)
	ctx.Step(`^все операции в отчёте совместимы \(status="compatible", findings пуст\)$`, w.allOperationsCompatible)
	ctx.Step(`^хотя бы одна операция в отчёте несовместима с findings, содержащими правило "([^"]*)"$`, w.anyOperationHasFindingRule)
	ctx.Step(`^файл отчёта записан по пути "([^"]*)"$`, w.reportFileWritten)
	ctx.Step(`^повторный запуск на той же фикстуре и файл отчёта "([^"]*)" побайтово идентичен первому без учёта поля "generated_at"$`, w.determinismCheck)
	ctx.Step(`^провайдерская спецификация на диске не изменилась$`, w.providerSpecUnchanged)
}

// useFixture выбирает /fixtures/<name>/contract-tests.yaml как конфиг для
// следующего запуска. Каждая фикстура (CS1..CS9) изолирована по своей директории
// (reference.md: fixtures isolated per slice) — ничего общего между сценариями
// не переиспользуется, кроме самого шага выбора.
func (w *World) useFixture(name string) error {
	dir := filepath.Join("/fixtures", name)
	cfg := filepath.Join(dir, "contract-tests.yaml")
	if _, err := os.Stat(cfg); err != nil {
		return fmt.Errorf("фикстура %q: конфиг не найден по %s: %w", name, cfg, err)
	}
	w.fixtureDir = dir
	w.configPath = cfg
	// Базовый хэш локальной provider-спеки, если она у фикстуры есть (CS9: "провайдерская
	// спецификация на диске не изменилась" сверяется с этим базовым значением).
	providerPath := filepath.Join(dir, "provider.yaml")
	if b, err := os.ReadFile(providerPath); err == nil {
		w.providerSpecPath = providerPath
		w.providerSpecHashBefore = sha256Hex(b)
	} else {
		w.providerSpecPath = ""
		w.providerSpecHashBefore = ""
	}
	return nil
}

// runWithFixtureConfig — «запускаем инструмент с этим конфигом»: `run <config>`
// на бинаре, выбранном useFixture. Обёртка над общим w.exec (cli_steps.go).
func (w *World) runWithFixtureConfig() error {
	if w.configPath == "" {
		return fmt.Errorf("фикстура не выбрана — нужен шаг \"используем фикстуру\" раньше")
	}
	return w.exec([]string{"run", w.configPath})
}

// decodeStdoutJSON разбирает ПЕРВЫЙ JSON-объект из вывода инструмента. Вывод —
// combined stdout+stderr (w.exec использует CombinedOutput), поэтому используем
// json.Decoder.Decode (останавливается после первого валидного значения), а не
// json.Unmarshal на всю строку — иначе хвост stderr после отчёта ломает разбор.
func (w *World) decodeStdoutJSON() (map[string]any, error) {
	if !w.ran {
		return nil, fmt.Errorf("бинарь ещё не запускали")
	}
	dec := json.NewDecoder(strings.NewReader(w.lastOut))
	var v map[string]any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON-отчёт из вывода инструмента: %w (вывод: %s)", err, w.lastOut)
	}
	return v, nil
}

func (w *World) jsonFieldEquals(field, expected string) error {
	v, err := w.decodeStdoutJSON()
	if err != nil {
		return err
	}
	got, ok := v[field]
	if !ok {
		return fmt.Errorf("поле %q отсутствует в JSON stdout: %v", field, v)
	}
	if gotStr := fmt.Sprintf("%v", got); gotStr != expected {
		return fmt.Errorf("ожидали %s=%q, получили %s=%q (весь вывод: %s)", field, expected, field, gotStr, w.lastOut)
	}
	return nil
}

// allOperationsCompatible — CS1: report.schema.json инвариант status=compatible ⇒
// findings=[] для каждой operations[i].
func (w *World) allOperationsCompatible() error {
	v, err := w.decodeStdoutJSON()
	if err != nil {
		return err
	}
	ops, ok := v["operations"].([]any)
	if !ok {
		return fmt.Errorf("поле operations отсутствует или не массив: %v", v)
	}
	for i, raw := range ops {
		op, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("operations[%d] не объект: %v", i, raw)
		}
		if status, _ := op["status"].(string); status != "compatible" {
			return fmt.Errorf("operations[%d].status = %v, ожидали \"compatible\"", i, op["status"])
		}
		if findings, _ := op["findings"].([]any); len(findings) != 0 {
			return fmt.Errorf("operations[%d].findings не пуст: %v", i, op["findings"])
		}
	}
	return nil
}

// anyOperationHasFindingRule — CS2: хотя бы одна операция несовместима, и среди
// её findings есть указанное правило (report.schema.json: findings[].rule).
func (w *World) anyOperationHasFindingRule(rule string) error {
	v, err := w.decodeStdoutJSON()
	if err != nil {
		return err
	}
	ops, _ := v["operations"].([]any)
	for _, raw := range ops {
		op, _ := raw.(map[string]any)
		findings, _ := op["findings"].([]any)
		for _, rawF := range findings {
			f, _ := rawF.(map[string]any)
			if r, _ := f["rule"].(string); r == rule {
				return nil
			}
		}
	}
	return fmt.Errorf("ни одна операция не содержит finding с правилом %q (вывод: %s)", rule, w.lastOut)
}

func (w *World) reportFileWritten(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("файл отчёта %s не найден: %w", path, err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("файл отчёта %s пуст", path)
	}
	return nil
}

// determinismCheck — CS1 / Gate #1 A4: файл отчёта первого запуска (уже
// выполненного предыдущим Когда-шагом) сравнивается с файлом отчёта ВТОРОГО
// запуска на той же байт-в-байт идентичной фикстуре, без поля generated_at.
// Сравнение через map[string]any + reflect.DeepEqual (эквивалент
// `jq 'del(.generated_at)'` + diff, без внешней зависимости от jq в тестовом
// образе).
func (w *World) determinismCheck(path string) error {
	first, err := w.readReportSansGeneratedAt(path)
	if err != nil {
		return fmt.Errorf("первый запуск: %w", err)
	}
	firstExit := w.lastExit

	if err := w.runWithFixtureConfig(); err != nil {
		return err
	}
	secondExit := w.lastExit

	if firstExit != 0 || secondExit != 0 {
		return fmt.Errorf("оба запуска должны завершаться кодом 0, получили %d (первый) и %d (второй)", firstExit, secondExit)
	}

	second, err := w.readReportSansGeneratedAt(path)
	if err != nil {
		return fmt.Errorf("второй запуск: %w", err)
	}
	if !reflect.DeepEqual(first, second) {
		return fmt.Errorf("файлы отчёта различаются между запусками (без учёта generated_at):\n1) %v\n2) %v", first, second)
	}
	return nil
}

func (w *World) readReportSansGeneratedAt(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать %s: %w", path, err)
	}
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON из %s: %w", path, err)
	}
	delete(v, "generated_at")
	return v, nil
}

// providerSpecUnchanged — CS9: провайдерский master-спек — read-only ресурс;
// ReportWriter не должен его трогать даже когда сам не может записаться.
func (w *World) providerSpecUnchanged() error {
	if w.providerSpecPath == "" {
		return fmt.Errorf("нет базового хэша provider-спеки — фикстура без локального provider.yaml?")
	}
	b, err := os.ReadFile(w.providerSpecPath)
	if err != nil {
		return fmt.Errorf("не удалось перечитать provider-спеку %s: %w", w.providerSpecPath, err)
	}
	if after := sha256Hex(b); after != w.providerSpecHashBefore {
		return fmt.Errorf("provider-спека %s изменилась: было %s, стало %s", w.providerSpecPath, w.providerSpecHashBefore, after)
	}
	return nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
