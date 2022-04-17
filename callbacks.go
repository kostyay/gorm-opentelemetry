// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelgorm

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type internalCtxKey string

const (
	dbTableKey        = attribute.Key("db.sql.table")
	dbRowsAffectedKey = attribute.Key("db.rows_affected")
	dbOperationKey    = semconv.DBOperationKey
	dbStatementKey    = semconv.DBStatementKey
	omitVarsKey       = internalCtxKey("omit_vars")
)

func dbTable(name string) attribute.KeyValue {
	return dbTableKey.String(name)
}

func dbStatement(stmt string) attribute.KeyValue {
	return dbStatementKey.String(stmt)
}

func dbCount(n int64) attribute.KeyValue {
	return dbRowsAffectedKey.Int64(n)
}

func dbOperation(op string) attribute.KeyValue {
	return dbOperationKey.String(op)
}

func (op *OtelPlugin) spanName(tx *gorm.DB, operation string) string {
	query := extractQuery(tx)

	operation = operationForQuery(query, operation)

	target := op.cfg.dbName
	if target == "" {
		target = tx.Dialector.Name()
	}

	if tx.Statement != nil && tx.Statement.Table != "" {
		target += "." + tx.Statement.Table
	}

	return fmt.Sprintf("%s %s", operation, target)
}

func operationForQuery(query, op string) string {
	if op != "" {
		return op
	}

	return strings.ToUpper(strings.Split(query, " ")[0])
}

func (op *OtelPlugin) before(operation string) gormHookFunc {
	return func(tx *gorm.DB) {
		tx.Statement.Context, _ = op.tracer.
			Start(tx.Statement.Context, op.spanName(tx, operation), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	}
}

func extractQuery(tx *gorm.DB) string {
	if shouldOmit, _ := tx.Statement.Context.Value(omitVarsKey).(bool); shouldOmit {
		return tx.Statement.SQL.String()
	}
	return tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...)
}

const (
	eventMaxSize = 250
	maxChunks    = 4
)

func chunkBy(val string, size int, callback func(string, ...oteltrace.EventOption)) {
	if len(val) > maxChunks*size {
		return
	}

	for i := 0; i < maxChunks*size; i += size {
		end := len(val)
		if end > size {
			end = size
		}
		callback(val[0:end])
		if end > len(val)-1 {
			break
		}
		val = val[end:]
	}
}

func (op *OtelPlugin) after(operation string) gormHookFunc {
	return func(tx *gorm.DB) {
		span := oteltrace.SpanFromContext(tx.Statement.Context)
		if !span.IsRecording() {
			// skip the reporting if not recording
			return
		}
		defer span.End()

		span.SetName(op.spanName(tx, operation))
		// Error
		if tx.Error != nil {
			span.SetStatus(codes.Error, tx.Error.Error())
		}

		// extract the db operation
		query := strings.ToValidUTF8(extractQuery(tx), "")

		// If query is longer then max size log it as chunked event, otherwise log it in attribute
		if len(query) > eventMaxSize {
			chunkBy(query, eventMaxSize, span.AddEvent)
		} else {
			span.SetAttributes(dbStatement(query))
		}

		operation = operationForQuery(query, operation)
		if tx.Statement.Table != "" {
			span.SetAttributes(dbTable(tx.Statement.Table))
		}

		span.SetAttributes(
			dbOperation(operation),
			dbCount(tx.Statement.RowsAffected),
		)
	}
}
func WithOmitVariablesFromTrace(ctx context.Context) context.Context {
	return context.WithValue(ctx, omitVarsKey, true)
}
