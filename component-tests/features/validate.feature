# Realized 1:1 from docs/design/slice-01-validate/contracts.md
# §"Component scenarios (DESIGN half)" — the Gherkin outline there is verbatim source.
# N = 2 (happy-class: compatible verdict + incompatible verdict) + 5 (distinguishable adapter
# branches) = 7. The incompatible-verdict scenario was added by change 001-report-schema-1.1
# (canon 1.1: errors[].subject is observable only on a verdict) — a recorded design act; do not
# add/remove scenarios here casually, that is a design act, not a realization act (STOP, back
# to wirth-moduledesigner / contracts.md).
#
# @wip while the slice is unimplemented (build → unit → component test sequence; a half-built
# slice cannot turn this green). Only the fixer (@fagan) removes @wip, on GREEN, at slice
# acceptance — never this ticket.
@component @slice-01-validate
Feature: Forward-compatibility validation of a consumer↔provider pair

  Scenario: forward-compatible pair yields a compatible report
    # scenario 1 (happy) — MSS step 7 (F0)
    Given a config whose consumed-contract and provider spec are reachable and compatible
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 0
    And stdout is a schema-valid report with compatible=true and errors==[]
    And stdout is a canon 1.1 report (schema_version, validator, interaction, consumer.name, generated_at)
    And uncovered provider operations are listed in uncovered_operations[]

  Scenario: Consumer incompatible with provider (primary verdict)
    # scenario 2 (happy-class: the incompatible verdict) — canon 1.1: exit 1 carries
    # errors[].subject (the netlist edge granularity), proven by one representative rule (R1).
    Given a config whose scope includes an operation the provider does not expose
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 1
    And stdout report compatible=false with errors[0].subject naming the operation

  Scenario: Config not found, unreadable, malformed YAML, schema-invalid, or spec source not exactly-one
    # scenario 2 (CONFIG_ERROR) — Extension 1a
    Given a config file that cannot be read or is schema-invalid
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 2

  Scenario: consumed-contract or provider spec file missing / unreadable at its path
    # scenario 3 (FILE_NOT_FOUND) — Extension 2a
    Given a config pointing at a consumed-contract path that does not exist
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "FILE_NOT_FOUND"

  Scenario: consumed-contract or provider spec unparseable (invalid OpenAPI / YAML)
    # scenario 4 (PARSE_ERROR) — Extension 2b
    Given a config pointing at a provider spec that is not valid OpenAPI
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "PARSE_ERROR"

  Scenario: Provider spec_url unreachable (HTTP failure / non-2xx / connection refused)
    # scenario 5 (HTTP_ERROR) — Extension 3a
    Given a config with spec_url pointing at a stub that returns 503
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "HTTP_ERROR"

  Scenario: Provider spec_url fetch exceeds settings.timeout seconds
    # scenario 6 (TIMEOUT_ERROR) — Extension 3b
    Given a config with spec_url pointing at a stub that stalls past settings.timeout
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "TIMEOUT_ERROR"
