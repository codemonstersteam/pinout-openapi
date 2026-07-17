// Package provider — как приобретается и разбирается спека провайдера, и как из неё
// извлекается одна операция. Этот файл: DeriveProviderOperation (ticket 10, R1) —
// чистая навигация paths[path][method] уже распарсенной ProviderSpec.Doc
// (*openapi3.T, kin-openapi) → { requires, provides } по телу и параметрам
// (path/query/header). Никакого I/O здесь (io: none) — приобретение спеки живёт в
// SpecLoader.Load (loader.go, io: http, отдельный тикет).
package provider

import (
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"pinout-openapi/internal/validate/domain"
)

// EnumerateOperations перечисляет ВСЕ операции (path+method), объявленные в спеке
// провайдера — полная поверхность провайдера, из которой CompareContracts вычитает
// сконфигурированный scope (cfg.operations), чтобы получить uncovered_operations[]
// (contracts.md §CompareContracts; CONTEXT.md «Uncovered operation»). Как и
// DeriveProviderOperation — чистая навигация по уже распарсенной spec.Doc
// (*openapi3.T), io: none. Метод нормализуется в нижний регистр (config.schema.json
// method enum — lower-case), чтобы вычитание scope совпадало по ключу с OperationRef
// консьюмера. Результат детерминирован (отсортирован по path, затем method): порядок
// обхода map в kin-openapi не гарантирован, а uncovered_operations[] должен быть
// воспроизводим. nil / не-*openapi3.T / бес-путёвый документ → пусто.
func EnumerateOperations(spec domain.ProviderSpec) []domain.OperationRef {
	doc, ok := spec.Doc.(*openapi3.T)
	if !ok || doc == nil || doc.Paths == nil {
		return nil
	}

	var refs []domain.OperationRef
	for path, item := range doc.Paths.Map() {
		if item == nil {
			continue
		}
		for method := range item.Operations() {
			refs = append(refs, domain.OperationRef{Path: path, Method: strings.ToLower(method)})
		}
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Path != refs[j].Path {
			return refs[i].Path < refs[j].Path
		}
		return refs[i].Method < refs[j].Method
	})
	return refs
}

// DeriveProviderOperation навигирует paths[ref.Path][ref.Method] в уже распарсенной
// spec.Doc (*openapi3.T) и строит ProviderOperation{Requires, Provides}:
//   - Requires — обязательные поля тела запроса (application/json, schema.Required) +
//     обязательные параметры path/query/header (контравариант — что провайдер требует
//     от консьюмера);
//   - Provides — поля тела успешного ответа (application/json, 2xx) (ковариант — что
//     провайдер отдаёт консьюмеру).
//
// Второе возвращаемое значение — R1: false, если операция отсутствует в спеке
// провайдера (путь не найден ИЛИ метод для найденного пути не объявлен). Это
// НЕ ошибка трубы — честный вердикт, который CompareOperation сворачивает в
// OP_NOT_IN_PROVIDER (contracts.md §DeriveProviderOperation).
func DeriveProviderOperation(spec domain.ProviderSpec, ref domain.OperationRef) (domain.ProviderOperation, bool) {
	doc, ok := spec.Doc.(*openapi3.T)
	if !ok || doc == nil || doc.Paths == nil {
		return domain.ProviderOperation{}, false
	}

	pathItem := doc.Paths.Find(ref.Path)
	if pathItem == nil {
		return domain.ProviderOperation{}, false
	}

	op := pathItem.GetOperation(strings.ToUpper(ref.Method))
	if op == nil {
		return domain.ProviderOperation{}, false
	}

	return domain.ProviderOperation{
		Requires: requiredRequestFields(op),
		Provides: providedResponseFields(op),
	}, true
}

// requiredRequestFields — обязательные поля тела запроса (application/json) +
// обязательные параметры path/query/header, одной картой "имя" → "имя типа".
func requiredRequestFields(op *openapi3.Operation) map[string]string {
	fields := map[string]string{}

	for _, paramRef := range op.Parameters {
		if paramRef == nil || paramRef.Value == nil || !paramRef.Value.Required {
			continue
		}
		switch paramRef.Value.In {
		case "path", "query", "header":
			fields[paramRef.Value.Name] = schemaTypeName(paramRef.Value.Schema)
		}
	}

	bodySchema := jsonSchema(requestBodyContent(op))
	if bodySchema != nil {
		for _, name := range bodySchema.Required {
			fields[name] = schemaTypeName(bodySchema.Properties[name])
		}
	}

	return fields
}

// providedResponseFields — поля тела успешного (2xx) ответа (application/json), одной
// картой "имя" → "имя типа".
func providedResponseFields(op *openapi3.Operation) map[string]string {
	fields := map[string]string{}

	schema := jsonSchema(successResponseContent(op))
	if schema == nil {
		return fields
	}
	for name, propRef := range schema.Properties {
		fields[name] = schemaTypeName(propRef)
	}

	return fields
}

// requestBodyContent — Content тела запроса операции, либо nil, если тело не объявлено.
func requestBodyContent(op *openapi3.Operation) openapi3.Content {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return nil
	}
	return op.RequestBody.Value.Content
}

// successResponseContent — Content первого объявленного 2xx-ответа (200..204, затем
// любой прочий код с префиксом "2"), либо nil, если такого ответа нет.
func successResponseContent(op *openapi3.Operation) openapi3.Content {
	if op.Responses == nil {
		return nil
	}
	for _, code := range []int{200, 201, 202, 203, 204} {
		if respRef := op.Responses.Status(code); respRef != nil && respRef.Value != nil {
			return respRef.Value.Content
		}
	}
	for status, respRef := range op.Responses.Map() {
		if strings.HasPrefix(status, "2") && respRef != nil && respRef.Value != nil {
			return respRef.Value.Content
		}
	}
	return nil
}

// jsonSchema — схема application/json из Content, либо nil.
func jsonSchema(content openapi3.Content) *openapi3.Schema {
	if content == nil {
		return nil
	}
	media := content.Get("application/json")
	if media == nil || media.Schema == nil {
		return nil
	}
	return media.Schema.Value
}

// schemaTypeName — первое имя типа схемы (string/integer/number/boolean/array/object),
// либо "" если схема/тип не заданы.
func schemaTypeName(ref *openapi3.SchemaRef) string {
	if ref == nil || ref.Value == nil || ref.Value.Type == nil {
		return ""
	}
	types := ref.Value.Type.Slice()
	if len(types) == 0 {
		return ""
	}
	return types[0]
}
