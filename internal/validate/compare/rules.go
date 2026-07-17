// Package compare — слайс slice-01-validate: доменное ядро forward-совместимости
// (contracts.md §CompareOperation; module-tree.md "compare"). Портировано ДОСЛОВНО из
// sandbox/check.mjs::compareOp (sandbox/ALGORITHM.md) на Go — не переизобретено.
package compare

import (
	"fmt"

	"pinout-openapi/internal/validate/domain"
)

// CompareOperation — contracts.md §CompareOperation:
// CompareOperation(consumedOp, providerOp|NotPresent) -> []Violation. Ядро сравнения
// по body+параметрам (path/query/header — та же плоская форма map[поле]тип, поэтому
// эквивалентны на этом уровне, lesson D1): R1 providerOp отсутствует (nil, NotPresent)
// ⇒ OP_NOT_IN_PROVIDER; R2 requires(provider) ⊆ sends(consumer) иначе
// MISSING_REQUIRED_REQUEST_FIELD; R3 reads(consumer) ⊆ provides(provider) иначе
// READS_FIELD_NOT_PROVIDED; R4 общие поля — совпадение типов (typesMatch) иначе
// TYPE_MISMATCH. Нарушение — ЗНАЧЕНИЕ, не ошибка: падающего антецедента нет.
//
// providerOp == nil кодирует "NotPresent" (R1) — сигнал DeriveProviderOperation
// (ticket 10, provider-пакет), без зависимости compare от provider-пакета.
func CompareOperation(consumedOp domain.ConsumedOperation, providerOp *domain.ProviderOperation) []domain.Violation {
	at := fmt.Sprintf("%s %s", consumedOp.Ref.Method, consumedOp.Ref.Path)

	if providerOp == nil {
		return []domain.Violation{{
			Code:     domain.CodeOpNotInProvider,
			Message:  "operation not present in provider spec",
			Location: at,
		}}
	}

	var violations []domain.Violation

	// R2, контравариантно: всё обязательное у поставщика потребитель обязан слать
	// (body-поле или параметр path/query/header — одна и та же плоская форма).
	for field := range providerOp.Requires {
		if _, sent := consumedOp.Sends[field]; !sent {
			violations = append(violations, domain.Violation{
				Code:     domain.CodeMissingRequiredRequestField,
				Message:  fmt.Sprintf("provider requires %q, consumer does not send it", field),
				Location: at,
				Details:  field,
			})
		}
	}

	// R4 (request-сторона): типы полей, что шлём и что поставщик знает по имени.
	for field, sentType := range consumedOp.Sends {
		if provType, known := providerOp.Requires[field]; known && !typesMatch(sentType, provType) {
			violations = append(violations, domain.Violation{
				Code:     domain.CodeTypeMismatch,
				Message:  fmt.Sprintf("request field %q: consumer sends %s, provider expects %s", field, sentType, provType),
				Location: at,
				Details:  field,
			})
		}
	}

	// R3, ковариантно: каждое читаемое потребителем поле ДОЛЖНО быть у поставщика
	// (ловит удаление поля); R4 (response-сторона): совпадение типа для общих полей.
	for field, readType := range consumedOp.Reads {
		provType, provided := providerOp.Provides[field]
		if !provided {
			violations = append(violations, domain.Violation{
				Code:     domain.CodeReadsFieldNotProvided,
				Message:  fmt.Sprintf("consumer reads %q, provider does not provide it", field),
				Location: at,
				Details:  field,
			})
			continue
		}
		if !typesMatch(readType, provType) {
			violations = append(violations, domain.Violation{
				Code:     domain.CodeTypeMismatch,
				Message:  fmt.Sprintf("response field %q: consumer reads %s, provider provides %s", field, readType, provType),
				Location: at,
				Details:  field,
			})
		}
	}

	return violations
}

// typesMatch — R4 type equality: имена типов (∈ {string,integer,number,boolean,array,
// object}, FRD Data dictionary B) сравниваются как значения — совпадение по имени типа,
// не по идентичности объекта.
func typesMatch(consumerType, providerType string) bool {
	return consumerType == providerType
}
