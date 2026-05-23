package store

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// Open は SQLite DB を開き、起動時に schema.sql を適用する。
// dsn は modernc.org/sqlite の DSN もしくは単純なファイルパス。
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	// 解析ワーカーをプール化（複数並列実行）すると、複数 goroutine が同じ DB へ
	// 同時書き込みする。SQLite はライタが1つに限られ、コネクションを複数開くと
	// "database is locked" (SQLITE_BUSY) が発生し得る。コネクションを1本に固定して
	// 全 SQL を直列化することで競合を根本的に避ける。DB アクセスは解析本体（秒オーダ）
	// に比べ極小・低頻度なので直列化のコストは無視できる。
	// 副次効果として、modernc.org/sqlite の :memory: はコネクション単位で別 DB に
	// なるため、テストの :memory: 利用も1本固定で安定する。
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// migrate は既存 DB に対して後付けカラム追加などの冪等な migration を適用する。
func migrate(db *sql.DB) error {
	if err := addColumnIfNotExists(db, "jobs", "pr_base_sha", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(db, "jobs", "phase", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	// analyses に PR メタ情報（タイトル・ref・SHA）を後付けする。
	// GitHub 定義行リンク / PR メタ表示は解析結果から復元するためここに永続化する。
	for _, col := range []string{"pr_title", "pr_base_ref", "pr_head_ref", "pr_base_sha", "pr_head_sha"} {
		if err := addColumnIfNotExists(db, "analyses", col, "TEXT NOT NULL DEFAULT ''"); err != nil {
			return err
		}
	}
	return nil
}

func addColumnIfNotExists(db *sql.DB, table, column, definition string) error {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("pragma table_info %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("scan table_info row: %w", err)
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate table_info: %w", err)
	}
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	return err
}
