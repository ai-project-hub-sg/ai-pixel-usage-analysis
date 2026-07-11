CREATE TABLE accounts (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  current_host TEXT,
  last_sync_at TEXT,
  last_error TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE usage_records (
  id INTEGER PRIMARY KEY,
  account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  upstream_id INTEGER NOT NULL,
  request_id TEXT,
  api_key_id INTEGER,
  upstream_account_id INTEGER,
  model TEXT,
  inbound_endpoint TEXT,
  upstream_endpoint TEXT,
  request_type TEXT,
  stream INTEGER,
  billing_mode TEXT,
  input_tokens INTEGER NOT NULL DEFAULT 0,
  output_tokens INTEGER NOT NULL DEFAULT 0,
  cache_creation_tokens INTEGER NOT NULL DEFAULT 0,
  cache_read_tokens INTEGER NOT NULL DEFAULT 0,
  total_cost TEXT NOT NULL DEFAULT '0',
  actual_cost TEXT NOT NULL DEFAULT '0',
  duration_ms INTEGER,
  first_token_ms INTEGER,
  created_at TEXT NOT NULL,
  raw_json TEXT NOT NULL CHECK(json_valid(raw_json)),
  UNIQUE(account_id, upstream_id)
);

CREATE INDEX usage_records_account_created ON usage_records(account_id, created_at);

CREATE TABLE balance_ledger_entries (
  id INTEGER PRIMARY KEY,
  account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  upstream_id INTEGER NOT NULL,
  direction TEXT NOT NULL CHECK(direction IN ('credit','debit')),
  amount TEXT NOT NULL,
  balance_after TEXT NOT NULL,
  reason TEXT NOT NULL,
  business_category TEXT NOT NULL,
  ref_type TEXT,
  ref_id INTEGER,
  remark_text TEXT NOT NULL,
  search_text TEXT NOT NULL,
  extracted_json TEXT NOT NULL CHECK(json_valid(extracted_json)),
  metadata_json TEXT NOT NULL CHECK(json_valid(metadata_json)),
  created_at TEXT NOT NULL,
  raw_json TEXT NOT NULL CHECK(json_valid(raw_json)),
  UNIQUE(account_id, upstream_id)
);

CREATE INDEX ledger_account_created ON balance_ledger_entries(account_id, created_at);
CREATE INDEX ledger_reason ON balance_ledger_entries(reason, business_category, direction);

CREATE TABLE sync_cursors (
  account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  data_type TEXT NOT NULL,
  cursor_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY(account_id, data_type)
);

CREATE TABLE sync_runs (
  id INTEGER PRIMARY KEY,
  account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  data_type TEXT NOT NULL,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  status TEXT NOT NULL,
  read_count INTEGER NOT NULL DEFAULT 0,
  write_count INTEGER NOT NULL DEFAULT 0,
  error_summary TEXT
);

CREATE TABLE upstream_health (
  account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  host_url TEXT NOT NULL,
  last_success_at TEXT,
  last_failure_at TEXT,
  consecutive_failures INTEGER NOT NULL DEFAULT 0,
  next_probe_at TEXT,
  PRIMARY KEY(account_id, host_url)
);

CREATE TABLE dashboard_users (
  id INTEGER PRIMARY KEY CHECK(id=1),
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE web_sessions (
  token_hash BLOB PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES dashboard_users(id) ON DELETE CASCADE,
  created_at TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  revoked_at TEXT
);

CREATE INDEX web_sessions_expiry ON web_sessions(expires_at);
