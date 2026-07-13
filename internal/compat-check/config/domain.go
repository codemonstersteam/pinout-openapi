package config

// RawConfig is the unparsed contract-tests.yaml payload (bytes), produced by ConfigReader.Read
// and consumed by NewConfig. No transformation happens on the I/O pipe — bytes in, bytes out.
type RawConfig []byte

// ProviderSourceKind distinguishes how the provider master spec is resolved.
type ProviderSourceKind string

const (
	ProviderSourceHTTP ProviderSourceKind = "http"
	ProviderSourceFile ProviderSourceKind = "file"
)

// Method is a normalized (lower-case) OpenAPI HTTP verb. Unexported construction — only
// NewMethod can produce a valid value.
type Method struct {
	verb string
}

// String returns the normalized (lower-case) verb.
func (m Method) String() string { return m.verb }

// OperationPath is an OpenAPI path key with a validated leading '/'.
type OperationPath struct {
	path string
}

// String returns the path.
func (p OperationPath) String() string { return p.path }

// Operation composes a validated OperationPath and Method.
type Operation struct {
	path   OperationPath
	method Method
}

// Path returns the operation's path.
func (o Operation) Path() OperationPath { return o.path }

// Method returns the operation's method.
func (o Operation) Method() Method { return o.method }

// ProviderSource resolves to exactly one origin for the provider master spec.
type ProviderSource struct {
	kind ProviderSourceKind
	url  string
	path string
}

// Kind reports whether the source is fetched over HTTP or read from a local file.
func (s ProviderSource) Kind() ProviderSourceKind { return s.kind }

// URL returns the HTTP(S) URL (only meaningful when Kind() == ProviderSourceHTTP).
func (s ProviderSource) URL() string { return s.url }

// Path returns the local filesystem path (only meaningful when Kind() == ProviderSourceFile).
func (s ProviderSource) Path() string { return s.path }

// Consumer is the expectations side: the OpenAPI spec whose consumer expectations must be
// satisfiable by the provider.
type Consumer struct {
	name       string
	specPath   string
	operations []Operation
}

// Name returns the consumer's free-form label.
func (c Consumer) Name() string { return c.name }

// SpecPath returns the filesystem path to the consumer OpenAPI document.
func (c Consumer) SpecPath() string { return c.specPath }

// Operations returns the configured operations the consumer expects to be compatible.
func (c Consumer) Operations() []Operation { return c.operations }

// Provider is the source-of-truth side: the provider master (= prod) OpenAPI spec.
type Provider struct {
	name   string
	source ProviderSource
}

// Name returns the provider's free-form label.
func (p Provider) Name() string { return p.name }

// Source returns the resolved provider spec origin.
func (p Provider) Source() ProviderSource { return p.source }

// Settings are optional run settings; every field has a committed default (config.schema.json).
type Settings struct {
	logLevel       string
	saveJSONReport bool
	jsonReportFile string
	timeout        int
}

// LogLevel returns the diagnostics verbosity (one of debug|info|warn|error).
func (s Settings) LogLevel() string { return s.logLevel }

// SaveJSONReport reports whether the JSON report is also written to JSONReportFile().
func (s Settings) SaveJSONReport() bool { return s.saveJSONReport }

// JSONReportFile returns the writable path for the JSON report.
func (s Settings) JSONReportFile() string { return s.jsonReportFile }

// Timeout returns the provider HTTP(S) fetch timeout in seconds.
func (s Settings) Timeout() int { return s.timeout }

// Config is the fully-validated contract-tests.yaml — the sole boundary through which an
// illegal config state can enter the program. Unexported fields; NewConfig is the only factory.
type Config struct {
	consumer Consumer
	provider Provider
	settings Settings
}

// Consumer returns the validated consumer block.
func (c Config) Consumer() Consumer { return c.consumer }

// Provider returns the validated provider block.
func (c Config) Provider() Provider { return c.provider }

// Settings returns the validated settings block.
func (c Config) Settings() Settings { return c.settings }
