-- analyses テーブル
CREATE TABLE IF NOT EXISTS analyses (
    id         TEXT PRIMARY KEY,
    owner      TEXT NOT NULL,
    repo       TEXT NOT NULL,
    pr_number  INTEGER NOT NULL,
    result     TEXT,    -- JSON
    created_at DATETIME NOT NULL
);

-- jobs テーブル
CREATE TABLE IF NOT EXISTS jobs (
    id          TEXT PRIMARY KEY,
    analysis_id TEXT NOT NULL,
    status      TEXT NOT NULL,
    error       TEXT,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

-- sessions テーブル（OAuth セッション・アクセストークン）
CREATE TABLE IF NOT EXISTS sessions (
    id            TEXT PRIMARY KEY,
    github_token  TEXT NOT NULL,
    user_login    TEXT NOT NULL,
    created_at    DATETIME NOT NULL
);
