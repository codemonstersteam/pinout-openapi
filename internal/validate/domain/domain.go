// Package domain — слайс slice-01-validate: доменные value types + сентинел-ошибки,
// общие для всех подпакетов (config/, contract/, provider/, compare/, report/, cli/) И
// головной трубы head.go (package validate). Листовой пакет: НЕ импортирует
// internal/validate/<subpkg> и НЕ импортирует internal/validate — все импортируют его,
// не наоборот (ADR-0004: домен вынесен в лист, чтобы head в package validate мог звать
// адаптеры без import cycle).
//
// Формы полей взяты дословно из api-specification/config.schema.json (вход) и
// api-specification/report.schema.json (выход) — см. docs/design/slice-01-validate/
// {module-tree,contracts}.md. Ничего не придумано сверх зафиксированного контракта.
package domain

// Invocation — плоский DTO ingress-а (cli.Parse, ticket 04): единственное, что нужно
// головной функции из argv. Путь к конфигу — только путь; сам файл читает ConfigStore.
type Invocation struct {
	ConfigPath string // позиционный аргумент <config.yaml>
	JSONReport bool   // флаг --json
	Verbose    bool   // флаг --verbose
}

// OperationRef — селектор операции провайдера (path+method), как в
// config.schema.json#/properties/consumer/properties/operations/items. Используется и
// как элемент scope-а (Config.Operations), и как ключ per-операции consumed-contract
// (ConsumedOperation.Ref), и как аргумент DeriveProviderOperation.
type OperationRef struct {
	Path   string `json:"path" yaml:"path"`
	Method string `json:"method" yaml:"method"`
}

// ProviderConfig — источник provider-спеки: name + РОВНО ОДИН из SpecPath/SpecURL
// (config.schema.json#/properties/provider, oneOf). Форма одинакова что до, что после
// валидации (NewConfig лишь проверяет exactly-one/non-empty, не переформирует поля) —
// используется и в RawConfig (сырое YAML-значение), и в валидном Config.
type ProviderConfig struct {
	Name     string `json:"name" yaml:"name"`
	SpecPath string `json:"spec_path,omitempty" yaml:"spec_path,omitempty"`
	SpecURL  string `json:"spec_url,omitempty" yaml:"spec_url,omitempty"`
}

// Settings — прогонные настройки после дефолтов и проверки диапазона NewConfig
// (config.schema.json#/properties/settings). Timeout в секундах, > 0.
type Settings struct {
	LogLevel       string
	SaveJSONReport bool
	JSONReportFile string
	Timeout        int
	IgnoreWarnings bool
}

// RawConfig — конфиг-файл, структурно раскодированный (YAML → Go), ЕЩЁ НЕ
// провалидированный. Форма 1:1 повторяет config.schema.json; NewConfig (ticket 07)
// превращает это в валидный Config. Поля Settings — указатели: так NewConfig отличает
// «поле отсутствует → применить дефолт схемы» от явного нулевого значения (нужно для
// bool-полей, где zero-value false неотличим от «не задано»).
type RawConfig struct {
	Consumer struct {
		Name                 string         `yaml:"name"`
		ConsumedContractPath string         `yaml:"consumed_contract_path"`
		Operations           []OperationRef `yaml:"operations"`
	} `yaml:"consumer"`
	Provider ProviderConfig `yaml:"provider"`
	Settings struct {
		LogLevel       *string `yaml:"log_level"`
		SaveJSONReport *bool   `yaml:"save_json_report"`
		JSONReportFile *string `yaml:"json_report_file"`
		Timeout        *int    `yaml:"timeout"`
		IgnoreWarnings *bool   `yaml:"ignore_warnings"`
	} `yaml:"settings"`
}

// Config — валидный, схема-совместимый конфиг (следствие NewConfig, ticket 07):
// непустые имена; Operations ≥ 1 (каждый Path соответствует `^/`, Method — из enum);
// Provider — ровно один источник задан; Settings — продефолчены и в диапазоне.
type Config struct {
	ConsumerName         string
	ConsumedContractPath string
	Operations           []OperationRef
	Provider             ProviderConfig
	Settings             Settings
}

// Provenance — штамп происхождения спеки провайдера (FRD Data dictionary B/C,
// config.schema.json не участвует — это данные consumed-contract, эхом попадающие в
// report.schema.json#/properties/provenance). Используется и во входном
// ConsumedContract, и в выходном Report (сквозной прогон через Comparison/Outcome).
type Provenance struct {
	Provider        string `json:"provider" yaml:"provider"`
	ProviderVersion string `json:"provider_version" yaml:"provider_version"`
	CapturedHash    string `json:"captured_hash" yaml:"captured_hash"`
}

// ConsumedOperation — одна операция consumed-contract: уже типизированные (не
// инферированные) request/response поля. Sends/Reads — map "имя поля" → "имя типа"
// (∈ {string,integer,number,boolean,array,object}, FRD Data dictionary B).
type ConsumedOperation struct {
	Ref   OperationRef      `json:"ref"`
	Sends map[string]string `json:"sends"`
	Reads map[string]string `json:"reads"`
}

// ConsumedContract — артефакт E-harness (consumer.consumed_contract_path): все
// операции консьюмера + провенанс. Типы приходят готовыми — инструмент их не
// инферирует (FRD UC-1, шаг 2).
type ConsumedContract struct {
	Operations []ConsumedOperation `json:"operations"`
	Provenance Provenance          `json:"provenance"`
}

// ProviderSpec — приобретённая и разобранная спека провайдера ($ref разрешены).
// Doc хранит распарсенный документ (наполняет provider.SpecLoader — kin-openapi
// *openapi3.T, ADR-0001); здесь — как `any`, чтобы этот листовой файл не тянул
// внешнюю OpenAPI-зависимость (её добавляет ticket 09 по скилу http-io).
type ProviderSpec struct {
	Doc any
}

// ProviderOperation — {requires, provides} одной операции провайдера, полученные
// навигацией по ProviderSpec (DeriveProviderOperation, ticket 10) по телу и
// параметрам (path/query/header): Requires — обязательные для запроса (контравариант),
// Provides — поля тела ответа (ковариант). Map "имя поля" → "имя типа", как в
// ConsumedOperation.Sends/Reads — чтобы R2..R4 сравнивали по одному ключу.
type ProviderOperation struct {
	Requires map[string]string
	Provides map[string]string
}

// Comparison — объединение трёх уже валидных входов в одну доменную сущность
// (NewComparison, ticket 12; конструктор-объединитель — правило «2+ сущности ⇒
// uniting constructor», без падающего антецедента). ScopedOps — cfg.Operations
// (сфера сравнения); Consumed/Spec/Provenance — данные для CompareContracts.
type Comparison struct {
	ScopedOps  []OperationRef
	Consumed   ConsumedContract
	Spec       ProviderSpec
	Provenance Provenance
}

// Violation — одно нарушение форвард-совместимости (или io/parse-нарушение,
// попавшее в отчёт) — форма 1:1 с report.schema.json#/properties/errors/items.
// Code — один из четырёх verdict-кодов (errors.go) при exit 1, либо io/parse-код
// при exit 3 (FoldReport/ReportWriter).
type Violation struct {
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Location string         `json:"location,omitempty"`
	Details  string         `json:"details,omitempty"`
	Context  map[string]any `json:"context,omitempty"`
}

// ComparisonOutcome — итог чистого свёртывания CompareContracts (ticket 13) по всем
// ScopedOps: Violations пуст ⇔ compatible; UncoveredOps — операции провайдера вне
// scope (информационно, не влияют на вердикт/exit — report.schema.json
// #/properties/uncovered_operations).
type ComparisonOutcome struct {
	Violations   []Violation
	UncoveredOps []string
	Provenance   Provenance
}

// Report — выходной DTO, печатаемый на stdout (и, при Settings.SaveJSONReport, в
// файл) — форма 1:1 с report.schema.json. Инвариант: Compatible ⇔ len(Errors) == 0.
type Report struct {
	SchemaVersion       string      `json:"schema_version"`
	Compatible          bool        `json:"compatible"`
	Provenance          Provenance  `json:"provenance"`
	Errors              []Violation `json:"errors"`
	UncoveredOperations []string    `json:"uncovered_operations,omitempty"`
}
