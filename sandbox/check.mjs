#!/usr/bin/env node
// Референс-чекер pinout-openapi (песочница). Демонстрирует АЛГОРИТМ, а не продовый инструмент:
// прод-версия — Go + kin-openapi. Спеки — YAML (как в реальности); парсим js-yaml.
// Запуск: `npm install && node check.mjs`. Логика — в ALGORITHM.md. Функциональный стиль (чистые функции).
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import yaml from "js-yaml";

const HERE = dirname(fileURLToPath(import.meta.url));
const readYaml = (p) => yaml.load(readFileSync(resolve(HERE, p), "utf8"));

// --- inferType :: value -> jsonSchemaType ------------------------------------
// В JSON нет "integer": целое число трактуем как integer (совпадает с OpenAPI-типом провайдера).
const inferType = (v) =>
  typeof v === "number" ? (Number.isInteger(v) ? "integer" : "number")
  : Array.isArray(v)    ? "array"
  : typeof v;

// --- shapeOf :: objectExample -> { field: type } -----------------------------
const shapeOf = (obj) =>
  Object.fromEntries(Object.entries(obj ?? {}).map(([k, v]) => [k, inferType(v)]));

// --- reconstruct :: stub -> consumedContract ---------------------------------
// ИЗ конкретных примеров стаба ВОССТАНАВЛИВАЕМ форму: что потребитель ШЛЁТ и что ЧИТАЕТ.
const reconstruct = (stub) =>
  stub.interactions.map((it) => ({
    path: it.operation.path,
    method: it.operation.method,
    sends: shapeOf(it.request?.body),      // request-поля (контравариантно)
    reads: shapeOf(it.response?.body),     // response-поля (ковариантно)
  }));

// --- providerOp :: (spec, path, method) -> { requestReq, requestProps, responseProps } | null
const jsonSchema = (mediaHolder) =>
  mediaHolder?.content?.["application/json"]?.schema ?? { properties: {}, required: [] };

const providerOp = (spec, path, method) => {
  const op = spec.paths?.[path]?.[method];
  if (!op) return null;
  const req = jsonSchema(op.requestBody);
  const res = jsonSchema(op.responses?.["200"]);
  return {
    requestRequired: req.required ?? [],
    requestProps: req.properties ?? {},
    responseProps: res.properties ?? {},
  };
};

// --- compareOp :: (consumedOp, providerSpec) -> [error] -----------------------
// Ядро: "требует ⊆ шлёт" (контравариантно) + "читает ⊆ отдаёт" (ковариантно) + совпадение типов.
const compareOp = (c, spec) => {
  const p = providerOp(spec, c.path, c.method);
  const at = `${c.method.toUpperCase()} ${c.path}`;
  if (!p) return [{ code: "OP_NOT_IN_PROVIDER", at, detail: "операции нет у поставщика" }];

  const errors = [];

  // REQUEST, контравариантно: всё обязательное у поставщика потребитель обязан слать.
  for (const field of p.requestRequired)
    if (!(field in c.sends))
      errors.push({ code: "MISSING_REQUIRED_REQUEST_FIELD", at, detail: `поставщик требует '${field}', потребитель его не шлёт` });

  // REQUEST-типы для полей, что шлём и что поставщик знает.
  for (const [field, t] of Object.entries(c.sends))
    if (p.requestProps[field] && p.requestProps[field].type !== t)
      errors.push({ code: "TYPE_MISMATCH", at, detail: `request '${field}': шлём ${t}, поставщик ждёт ${p.requestProps[field].type}` });

  // RESPONSE, ковариантно: каждое читаемое поле ДОЛЖНО быть у поставщика (ловит удаление).
  for (const [field, t] of Object.entries(c.reads)) {
    const prov = p.responseProps[field];
    if (!prov)
      errors.push({ code: "READS_FIELD_NOT_PROVIDED", at, detail: `потребитель читает '${field}', поставщик его не отдаёт` });
    else if (prov.type !== t)
      errors.push({ code: "TYPE_MISMATCH", at, detail: `response '${field}': читаем ${t}, поставщик отдаёт ${prov.type}` });
  }
  return errors;
};

// --- check :: (consumedContract, providerSpec) -> report ---------------------
const check = (consumed, spec) => {
  const errors = consumed.flatMap((c) => compareOp(c, spec));
  return { compatible: errors.length === 0, errors };
};

// --- прогон: v1 + 4 сценария --------------------------------------------------
const stub = readYaml("consumer/component-tests/provider-stub.yaml");
const consumed = reconstruct(stub);

const runs = [
  ["v1  (базовая)",                        "provider/openapi.v1.yaml"],
  ["v2a +optional response (promoCode)",   "scenarios/v2a-added-optional-response.yaml"],
  ["v2b -response field (currency убрали)", "scenarios/v2b-removed-response-field.yaml"],
  ["v2c +required request (idempotency)",  "scenarios/v2c-added-required-request.yaml"],
  ["v2d type-drift (balance→string)",      "scenarios/v2d-type-drift.yaml"],
];

console.log("consumed-contract, восстановленный из стаба:");
for (const c of consumed)
  console.log(`  ${c.method.toUpperCase()} ${c.path}  шлёт=${JSON.stringify(c.sends)}  читает=${JSON.stringify(c.reads)}`);
console.log("");

let failures = 0;
for (const [label, specPath] of runs) {
  const { compatible, errors } = check(consumed, readYaml(specPath));
  console.log(`${compatible ? "✅ СОВМЕСТИМО  " : "❌ НЕСОВМЕСТИМО"}  ${label}`);
  for (const e of errors) console.log(`     • ${e.code} @ ${e.at}: ${e.detail}`);
  if (!compatible) failures++;
}
console.log(`\nИтог: ${runs.length - failures}/${runs.length} совместимых прогонов.`);
