package store

import _ "embed"

//go:embed schema.sql
var schema string

// TODO: modernc.org/sqlite で接続管理・起動時に schema.sql を自動適用する
