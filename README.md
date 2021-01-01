# gorm-opentelemetry
[![Test Status](https://github.com/kostyay/gorm-opentelemetry/workflows/Test/badge.svg)](github.com/kostyay/gorm-opentelemetry/actions)

OpenTelemetry plugin for GORM v2

Traces all queries along with the query SQL.

Usage Example:
====
```go
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

package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/kostyay/gorm-opentelemetry"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func ExampleNewPlugin() {
	// Initialize db connection
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Initialize otel plugin with options
	plugin := otelgorm.NewPlugin(
	// include any options here
	)
	err = db.Use(plugin)
	if err != nil {
		panic("failed configuring plugin")
	}

	// Migrate the schema
	err = db.AutoMigrate(&Product{})
	if err != nil {
		panic(err.Error())
	}

	// Create
	db.Create(&Product{Code: "D42", Price: 100})

	// Read
	var product Product
	db.First(&product, 1)                 // find product with integer primary key
	db.First(&product, "code = ?", "D42") // find product with code D42
}


```