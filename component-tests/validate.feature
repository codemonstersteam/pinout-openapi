Feature: Валидация совместимости синхронного контракта
  Как CI потребителя, я хочу узнать до merge, совместим ли потребитель
  с master-OpenAPI поставщика, по конфигурации contract-tests.yaml.

  # Контракт-первый: режимы отказа взяты из README/контракта CLI, не из кода.
  # N_тестов = 1 (happy) + 5 (различимых режимов отказа) = 6.

  Scenario: Совместимая пара контрактов
    Given конфиг указывает на спеку потребителя и master-спеку поставщика
    And операция потребителя присутствует у поставщика и схемы совместимы
    When запускаю "validate ./contract-tests.yaml"
    Then exit code 0
    And вывод содержит "✅ Контракты совместимы"
    And создан JSON-отчёт в общем формате

  Scenario: Операции потребителя нет у поставщика
    Given операция потребителя отсутствует в master-спеке поставщика
    When запускаю "validate ./contract-tests.yaml"
    Then exit code 1
    And отчёт содержит ошибку с кодом "OPERATION_NOT_FOUND"

  Scenario: Несовместимый request
    Given поставщик требует обязательный параметр, которого нет у потребителя
    When запускаю "validate ./contract-tests.yaml"
    Then exit code 1
    And отчёт содержит ошибку с кодом "REQUEST_INCOMPATIBLE"

  Scenario: Несовместимый response
    Given потребитель ожидает поле ответа, которого нет в схеме поставщика
    When запускаю "validate ./contract-tests.yaml"
    Then exit code 1
    And отчёт содержит ошибку с кодом "RESPONSE_INCOMPATIBLE"

  Scenario: Конфигурация отсутствует или невалидна
    Given путь к contract-tests.yaml не существует
    When запускаю "validate ./contract-tests.yaml"
    Then exit code 2
    And вывод содержит код ошибки "CONFIG_NOT_FOUND"

  Scenario: Спека поставщика недоступна
    Given spec_url поставщика недоступен в пределах timeout
    When запускаю "validate ./contract-tests.yaml"
    Then exit code 3
    And вывод содержит код ошибки "SPEC_UNREACHABLE"
