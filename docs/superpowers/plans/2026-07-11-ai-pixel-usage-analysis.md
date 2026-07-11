# AI Pixel Usage Analysis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付一个支持多账户采集、SQLite 分析、24 小时登录会话、Nginx 反向代理和 Linux 单二进制部署的 AI Pixel 用量分析系统。

**Architecture:** Go 单体进程负责配置、认证、上游采集、SQLite、分析 API 和内嵌前端；React/TypeScript 前端构建后通过 `go:embed` 打入二进制。Go 仅监听 TOML 指定的本机地址与端口，公网 HTTPS 由 Nginx 提供。

**Tech Stack:** Go 1.25、标准库 `net/http`、`modernc.org/sqlite`、`go-toml/v2`、`x/crypto/argon2`、React、TypeScript、Vite、ECharts、Vitest、Testing Library、Playwright。

---

## File map

### Go entry point and composition

- `cmd/ai-pixel-usage-analysis/main.go`：命令行入口、依赖装配、启动和优雅关闭。
- `internal/app/app.go`：应用生命周期、HTTP 服务和同步调度器组合。

### Configuration and secrets

- `internal/config/config.go`：TOML 结构、解析、默认值和校验。
- `internal/config/config_test.go`：监听、时区、节点和账户配置测试。
- `internal/secrets/dotenv.go`：保留注释的 `.env` 解析与原子更新。
- `internal/secrets/dashboard.go`：网页登录凭据读取和安全生成。
- `internal/secrets/dashboard_test.go`：缺失、部分缺失、权限和保留内容测试。

### Database and authentication

- `internal/store/store.go`：SQLite 打开、PRAGMA 和事务辅助。
- `internal/store/migrations.go`、`internal/store/migrations/*.sql`：版本化数据库迁移。
- `internal/store/store_test.go`：迁移、约束和重开测试。
- `internal/auth/password.go`：Argon2id 哈希及恒定时间验证。
- `internal/auth/service.go`：管理员同步、会话创建、验证、撤销和清理。
- `internal/auth/service_test.go`：密码轮换、会话过期及撤销测试。

### Upstream collection

- `internal/upstream/types.go`：登录、请求记录、余额流水和分页 DTO。
- `internal/upstream/client.go`：`/api/v1` HTTP 客户端、登录、刷新和分页请求。
- `internal/upstream/client_test.go`：模拟上游协议测试。
- `internal/upstream/failover.go`：优先级、同权重顺序、切换和首选节点复探。
- `internal/upstream/failover_test.go`：节点行为测试。
- `internal/ledger/remark.go`：原始流水类型、业务分类和备注规范化。
- `internal/ledger/remark_test.go`：已知及未知 metadata 规则测试。
- `internal/syncer/repository.go`：请求、流水、游标和运行记录的幂等写入。
- `internal/syncer/service.go`：初始化、增量、重叠窗口和分页同步。
- `internal/syncer/scheduler.go`：每分钟调度、单账户防重入和手动同步。
- `internal/syncer/service_test.go`：初始化、恢复、重复和部分失败集成测试。

### Analytics and HTTP

- `internal/analytics/types.go`：筛选器、时间桶、KPI 和对比响应。
- `internal/analytics/query.go`：请求和流水 SQL 聚合与明细查询。
- `internal/analytics/compare.go`：昨天同期、上周同期和零基数处理。
- `internal/analytics/query_test.go`：多账户隔离、聚合和筛选测试。
- `internal/httpapi/router.go`：路由、中间件和统一 JSON 错误。
- `internal/httpapi/auth_handlers.go`：登录、退出、会话状态。
- `internal/httpapi/analytics_handlers.go`：总览、请求、流水和账户状态 API。
- `internal/httpapi/httpapi_test.go`：未登录保护、Cookie、Origin 和 API 集成测试。
- `internal/webui/embed.go`：嵌入前端和 SPA fallback。
- `internal/webui/dist/placeholder.txt`：前端首次构建前保证 embed 目录包含可嵌入文件。

### Frontend

- `web/package.json`、`web/package-lock.json`、`web/tsconfig.json`、`web/vite.config.ts`：构建与测试配置。
- `web/src/main.tsx`、`web/src/App.tsx`：入口、会话恢复和路由外壳。
- `web/src/api/client.ts`、`web/src/api/types.ts`：同源 API 客户端与类型。
- `web/src/styles.css`：设计 token、响应式布局和状态样式。
- `web/src/pages/LoginPage.tsx`：登录页面。
- `web/src/pages/OverviewPage.tsx`：KPI 与同期趋势。
- `web/src/pages/UsagePage.tsx`：请求分析与明细。
- `web/src/pages/LedgerPage.tsx`：流水类型、方向和备注分析。
- `web/src/pages/AccountsPage.tsx`：账户同步状态与手动重试。
- `web/src/components/*`：筛选栏、KPI、图表、表格、状态和分页组件。
- `web/src/**/*.test.tsx`：Vitest 组件测试。
- `web/e2e/dashboard.spec.ts`：Playwright 登录和分析流程。

### Documentation and deployment

- `docs/prd.md`、`docs/ui.md`、`docs/trd.md`、`docs/structure.md`、`docs/development.md`、`docs/acceptance.md`、`docs/packaging.md`：README 要求的交付文档。
- `deploy/nginx.conf.example`：HTTPS 反向代理示例。
- `deploy/ai-pixel-usage-analysis.service`：systemd 示例。
- `scripts/build.ps1`：当前 Windows 工作区可直接运行的测试、前端构建、Linux 构建和验收入口。
- `Makefile`：Linux 开发环境的可选快捷入口，转调与 PowerShell 脚本等价的命令。
- `README.md`：快速开始、配置和文档索引。

## Task 1: Write the required product and engineering documents first

**Files:**
- Create: `docs/prd.md`
- Create: `docs/ui.md`
- Create: `docs/trd.md`
- Create: `docs/structure.md`
- Create: `docs/development.md`
- Create: `docs/acceptance.md`
- Create: `docs/packaging.md`

- [ ] **Step 1: Write the seven documents from the approved specification**

Use these exact document responsibilities:

```text
prd.md: users, goals, v1 scope, user journeys, functional requirements, non-functional requirements, acceptance summary
ui.md: navigation, login, overview, usage, ledger, account status, filters, responsive behavior, loading/empty/error states
trd.md: deployment architecture, upstream protocol, schema, sync algorithm, API contract, auth and threat controls
structure.md: repository tree and one responsibility per package/file
development.md: prerequisites, TDD loop, commands, style rules, migrations, secrets, commit conventions
acceptance.md: executable requirement matrix mapping every v1 criterion to a command or browser scenario
packaging.md: frontend build, CGO-free Linux build, artifact contents, config, .env permissions, Nginx and systemd deployment
```

Every document must state that Nginx is the public entry point, Go binds the configured local port, dashboard credentials live only in `.env`, and SQLite stores only Argon2id salted hashes.

- [ ] **Step 2: Verify document coverage and absence of incomplete markers**

Run:

```powershell
rg -n "^#|Nginx|Argon2id|DASHBOARD_USERNAME|DASHBOARD_PASSWORD|24 小时|多账户" docs/prd.md docs/ui.md docs/trd.md docs/structure.md docs/development.md docs/acceptance.md docs/packaging.md
rg -n -i "TO[D]O|T[B]D|FIX[M]E|待[定]|待[补]充" docs
```

Expected: the first command shows matching requirements in the relevant documents; the second command exits with code 1 and no matches.

- [ ] **Step 3: Commit documentation baseline**

```powershell
git add docs/prd.md docs/ui.md docs/trd.md docs/structure.md docs/development.md docs/acceptance.md docs/packaging.md
git commit -m "docs: add product and engineering documentation"
```

## Task 2: Bootstrap Go, frontend, and build orchestration

**Files:**
- Create: `go.mod`
- Create: `cmd/ai-pixel-usage-analysis/main.go`
- Create: `internal/app/app.go`
- Create: `internal/webui/embed.go`
- Create: `internal/webui/dist/placeholder.txt`
- Create: `web/package.json`
- Create: `web/tsconfig.json`
- Create: `web/vite.config.ts`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/styles.css`
- Create: `scripts/build.ps1`
- Create: `Makefile`

- [ ] **Step 1: Initialize dependency manifests**

Run:

```powershell
go mod init github.com/ai-project-hub-sg/ai-pixel-usage-analysis
go get github.com/pelletier/go-toml/v2 modernc.org/sqlite golang.org/x/crypto
New-Item -ItemType Directory -Force web | Out-Null
Push-Location web
npm init -y
npm install react react-dom echarts
npm install -D typescript vite @vitejs/plugin-react vitest jsdom @testing-library/react @testing-library/jest-dom @testing-library/user-event playwright @types/react @types/react-dom
Pop-Location
```

Expected: `go.mod`, `go.sum`, `web/package.json`, and `web/package-lock.json` exist.

Set the frontend scripts to these exact commands:

```json
{
  "scripts": {
    "build": "vite build",
    "test": "vitest",
    "test:e2e": "playwright test"
  }
}
```

- [ ] **Step 2: Write a failing application smoke test**

Create `internal/app/app_test.go`:

```go
package app

import "testing"

func TestName(t *testing.T) {
	if Name != "ai-pixel-usage-analysis" {
		t.Fatalf("unexpected application name %q", Name)
	}
}
```

- [ ] **Step 3: Run the test and verify the red state**

Run: `go test ./internal/app -run TestName -v`

Expected: FAIL because `Name` is undefined.

- [ ] **Step 4: Add the minimal application and embedded UI skeleton**

`internal/app/app.go` starts with:

```go
package app

const Name = "ai-pixel-usage-analysis"
```

`internal/webui/embed.go` embeds `dist`, and `web/vite.config.ts` sets `build.outDir` to `../internal/webui/dist` with `emptyOutDir: true`. `scripts/build.ps1` accepts `test`, `web-build`, `build`, `build-linux`, and `verify`; `build-linux` temporarily sets `CGO_ENABLED=0`, `GOOS=linux`, and `GOARCH=amd64`, then restores the caller's environment. The optional `Makefile` exposes equivalent targets.

- [ ] **Step 5: Verify manifests and skeleton**

Run:

```powershell
go test ./internal/app -v
Push-Location web; npm run build; Pop-Location
go test ./...
```

Expected: all commands exit 0 and Vite writes `internal/webui/dist/index.html`.

- [ ] **Step 6: Commit bootstrap**

```powershell
git add go.mod go.sum cmd internal/app internal/webui web scripts/build.ps1 Makefile
git commit -m "build: bootstrap Go and embedded React application"
```

## Task 3: Parse and validate TOML configuration

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Modify: `config.toml`

- [ ] **Step 1: Write failing configuration tests**

Tests must cover defaults, configured port, host priority order with stable ties, duplicate account IDs, missing account credential keys, invalid timezone, zero hosts, and `session_ttl != 24h` rejection.

Use the public contract:

```go
cfg, err := config.Load(path)
addr := cfg.Server.Address()
hosts := cfg.OrderedHosts()
```

Expected configured address: `127.0.0.1:8080`; expected tie behavior: declaration order is preserved.

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/config -v`

Expected: FAIL because package functions and types do not exist.

- [ ] **Step 3: Implement typed configuration and validation**

Define these core types:

```go
type Config struct {
	Server   ServerConfig    `toml:"server"`
	Analysis AnalysisConfig  `toml:"analysis"`
	Auth     AuthConfig      `toml:"auth"`
	Hosts    []HostConfig    `toml:"host"`
	Accounts []AccountConfig `toml:"account"`
}

type ServerConfig struct { Host string; Port int; PublicURL string `toml:"public_url"`; SecureCookie bool `toml:"secure_cookie"` }
type AnalysisConfig struct {
	Timezone string `toml:"timezone"`
	SyncInterval string `toml:"sync_interval"`
	SyncOverlap string `toml:"sync_overlap"`
	PreferredHostProbeInterval string `toml:"preferred_host_probe_interval"`
}
type AuthConfig struct { SessionTTL string `toml:"session_ttl"` }
type HostConfig struct { URL string; Priority int }
type AccountConfig struct { ID string; Name string; EmailEnv string `toml:"email_env"`; PasswordEnv string `toml:"password_env"`; Enabled bool }
```

Apply defaults from the approved design, parse durations, load `time.Location`, normalize base URLs, and sort a copy of hosts with `sort.SliceStable`.

- [ ] **Step 4: Replace the existing TOML with the approved shape**

Keep the three existing upstream domains and `user1`/`password1`; add `[server]`, `[analysis]`, `[auth]`, `[[host]]`, and `[[account]]`. Do not add dashboard username or password fields.

- [ ] **Step 5: Verify configuration behavior**

Run:

```powershell
go test ./internal/config -v
go test ./...
```

Expected: all tests PASS.

- [ ] **Step 6: Commit configuration**

```powershell
git add internal/config config.toml
git commit -m "feat: add validated multi-account configuration"
```

## Task 4: Generate and persist dashboard credentials in `.env`

**Files:**
- Create: `internal/secrets/dotenv.go`
- Create: `internal/secrets/dashboard.go`
- Create: `internal/secrets/dashboard_test.go`

- [ ] **Step 1: Write failing secret-management tests**

Cover these cases:

```go
func TestEnsureDashboardCredentialsCreatesBoth(t *testing.T)
func TestEnsureDashboardCredentialsGeneratesOnlyMissingValue(t *testing.T)
func TestEnsureDashboardCredentialsPreservesCommentsAndAccountSecrets(t *testing.T)
func TestEnsureDashboardCredentialsDoesNotUseProcessEnvironment(t *testing.T)
func TestEnsureDashboardCredentialsRejectsUnwritableFile(t *testing.T)
```

Assert that generated passwords decode to at least 16 random bytes, existing values remain unchanged, file content contains exactly one instance of each fixed key, and Unix file mode is `0600`.

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/secrets -v`

Expected: FAIL because `EnsureDashboardCredentials` is undefined.

- [ ] **Step 3: Implement atomic `.env` preservation and generation**

Expose:

```go
type DashboardCredentials struct { Username, Password string }
type AccountCredentials struct { Email, Password string }

func EnsureDashboardCredentials(path string) (DashboardCredentials, bool, error)
func LoadAccountCredentials(path string, accounts []config.AccountConfig) (map[string]AccountCredentials, error)
```

Use `crypto/rand`, base64url without padding, a same-directory temporary file, `Sync`, `Chmod(0600)`, and `Rename`. Preserve unrelated lines and comments. Return `generated=true` when either value was created. Never log credential values.

- [ ] **Step 4: Verify red-to-green and file safety**

Run:

```powershell
go test ./internal/secrets -v
go test ./...
```

Expected: all tests PASS; generated test files contain preserved upstream secrets and secure dashboard keys.

- [ ] **Step 5: Commit secret management**

```powershell
git add internal/secrets
git commit -m "feat: securely initialize dashboard credentials"
```

## Task 5: Add SQLite schema and migrations

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/migrations.go`
- Create: `internal/store/migrations/001_initial.sql`
- Create: `internal/store/store_test.go`

- [ ] **Step 1: Write a failing migration test**

Open a temporary database with `store.Open`, then query `sqlite_master` and assert these tables exist:

```text
accounts, usage_records, balance_ledger_entries, sync_cursors, sync_runs,
upstream_health, dashboard_users, web_sessions, schema_migrations
```

Also assert unique indexes on `(account_id, upstream_id)` and foreign keys enabled.

- [ ] **Step 2: Run the migration test and verify failure**

Run: `go test ./internal/store -v`

Expected: FAIL because `store.Open` is undefined.

- [ ] **Step 3: Implement the store and complete initial schema**

`store.Open` must use the modernc SQLite driver, set WAL, foreign keys, busy timeout, one writer connection, and run embedded migrations transactionally. Monetary columns use integer micro-units or decimal text consistently; timestamps use UTC RFC3339Nano text; raw payloads use JSON text with validity checks where supported.

- [ ] **Step 4: Verify migrations and reopen behavior**

Run:

```powershell
go test ./internal/store -v
go test ./...
```

Expected: all tests PASS, including opening the same database twice without rerunning migration side effects.

- [ ] **Step 5: Commit database foundation**

```powershell
git add internal/store
git commit -m "feat: add SQLite schema and migrations"
```

## Task 6: Implement Argon2id administrator and 24-hour sessions

**Files:**
- Create: `internal/auth/password.go`
- Create: `internal/auth/service.go`
- Create: `internal/auth/service_test.go`

- [ ] **Step 1: Write failing password and session tests**

Test hash/verify success, wrong password, malformed hash, independent salts, `.env` credential synchronization, password rotation revoking old sessions, 24-hour expiry, logout revocation, and database storage containing neither plaintext password nor raw session token.

Use:

```go
svc := auth.NewService(db, clock)
changed, err := svc.SyncDashboardUser(ctx, creds)
session, err := svc.Login(ctx, username, password)
user, err := svc.Authenticate(ctx, session.Token)
```

The injected clock satisfies `type Clock interface { Now() time.Time }`.

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/auth -v`

Expected: FAIL because the service is undefined.

- [ ] **Step 3: Implement password encoding and session hashing**

Use Argon2id with a versioned PHC-style string containing memory, iterations, parallelism, salt, and hash. Generate salts and 32-byte session tokens with `crypto/rand`; store only SHA-256 session token hashes. Use injected clocks so expiry tests do not sleep.

- [ ] **Step 4: Verify authentication behavior**

Run:

```powershell
go test ./internal/auth -v
go test ./...
```

Expected: all tests PASS; session expiry occurs exactly at the configured 24-hour boundary.

- [ ] **Step 5: Commit authentication core**

```powershell
git add internal/auth
git commit -m "feat: add salted dashboard authentication and sessions"
```

## Task 7: Implement the upstream API client

**Files:**
- Create: `internal/upstream/types.go`
- Create: `internal/upstream/client.go`
- Create: `internal/upstream/client_test.go`

- [ ] **Step 1: Write failing protocol tests with `httptest`**

The fake server must assert:

- public settings are read before login;
- login submits email, password, and `login_agreement_revision`;
- Authorization uses `Bearer`;
- refresh replaces both tokens;
- usage and ledger requests send page, page size, dates, times, timezone, sort order, direction, reason, reference type, and reference ID;
- non-2xx responses become typed auth, rate-limit, client, server, or transport errors.

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/upstream -run Client -v`

Expected: FAIL because `Client` and DTOs are undefined.

- [ ] **Step 3: Implement the client contract**

Expose:

```go
type API interface {
	Login(context.Context) error
	Refresh(context.Context) error
	ListUsage(context.Context, UsageQuery) (Page[UsageRecord], error)
	ListLedger(context.Context, LedgerQuery) (Page[LedgerEntry], error)
}

type Page[T any] struct { Items []T; Total, Page, PageSize, Pages int }
```

Model all confirmed upstream fields and preserve each raw JSON object. Set explicit HTTP timeouts and cap response bodies used for error messages.

- [ ] **Step 4: Verify the client**

Run:

```powershell
go test ./internal/upstream -run Client -v
go test ./...
```

Expected: all tests PASS and test logs contain no supplied passwords or tokens.

- [ ] **Step 5: Commit upstream protocol support**

```powershell
git add internal/upstream/types.go internal/upstream/client.go internal/upstream/client_test.go
git commit -m "feat: add authenticated upstream usage client"
```

## Task 8: Add host priority, failover, and preferred-host recovery

**Files:**
- Create: `internal/upstream/failover.go`
- Create: `internal/upstream/failover_test.go`

- [ ] **Step 1: Write failing failover tests**

Cover priority order, stable equal-priority order, sticky current host, finite retry, server-error failover, authentication error without blind failover, and recovery to a preferred host after the probe interval.

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/upstream -run Failover -v`

Expected: FAIL because the failover client does not exist.

- [ ] **Step 3: Implement the account-scoped failover client**

Maintain current host and health per account. Classify errors before switching. Apply exponential backoff with injected jitter and clock so tests remain deterministic. Persist health updates through a narrow repository interface.

- [ ] **Step 4: Verify failover behavior**

Run:

```powershell
go test ./internal/upstream -run Failover -v
go test ./...
```

Expected: all tests PASS with bounded attempts and deterministic host order.

- [ ] **Step 5: Commit failover**

```powershell
git add internal/upstream/failover.go internal/upstream/failover_test.go
git commit -m "feat: add weighted upstream failover"
```

## Task 9: Normalize ledger types and remarks

**Files:**
- Create: `internal/ledger/remark.go`
- Create: `internal/ledger/remark_test.go`

- [ ] **Step 1: Write table-driven failing tests**

Use fixtures for `usage_charge`, `redeem_code`, `account_share_income`, `private_group_commission`, subscription prepay/refund, admin adjustment, and an unknown future reason. Assert original reason preservation, business category, normalized text, searchable text, request/API key/account/group IDs, period and hourly rate extraction.

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/ledger -v`

Expected: FAIL because `Normalize` is undefined.

- [ ] **Step 3: Implement deterministic normalization**

Expose:

```go
type Remark struct {
	Category string
	Text string
	SearchText string
	Fields map[string]string
}

func Normalize(reason, refType string, refID int64, metadata json.RawMessage) Remark
```

Unknown input returns category `other`, includes the raw reason in searchable text, and never discards raw metadata from the caller.

- [ ] **Step 4: Verify normalization**

Run: `go test ./internal/ledger -v`

Expected: all tests PASS.

- [ ] **Step 5: Commit ledger analysis rules**

```powershell
git add internal/ledger
git commit -m "feat: add ledger type and remark normalization"
```

## Task 10: Add idempotent repositories, initial sync, and incremental sync

**Files:**
- Create: `internal/syncer/repository.go`
- Create: `internal/syncer/service.go`
- Create: `internal/syncer/scheduler.go`
- Create: `internal/syncer/service_test.go`

- [ ] **Step 1: Write failing integration tests**

Seed a fake clock and fake paginated upstream. Verify:

- no cursor starts at the previous calendar month's first instant in the analysis timezone;
- both usage and ledger pages are fetched to completion;
- duplicate pages produce one database row per `(account_id, upstream_id)`;
- a failed page does not advance the cursor;
- restart resumes from the stored cursor minus the overlap;
- two accounts remain isolated;
- one failed account does not block another;
- the scheduler prevents account re-entry and manual retry uses the same lock.

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/syncer -v`

Expected: FAIL because repository and service types are undefined.

- [ ] **Step 3: Implement transaction-scoped UPSERT repositories**

Implement usage, ledger, cursor, run, account status, and upstream health writes. Normalize ledger remarks before insert. Update cursor only in the transaction that completes the requested interval.

- [ ] **Step 4: Implement synchronization and scheduling**

Expose:

```go
type Service interface { SyncAccount(context.Context, string) error }
type Scheduler interface { Start(context.Context); Trigger(context.Context, string) error }
```

Use one account lock per enabled account, a one-minute ticker from config, bounded parallelism, context cancellation, and a run record for every attempt.

- [ ] **Step 5: Verify synchronization**

Run:

```powershell
go test ./internal/syncer -v
go test ./...
```

Expected: all tests PASS and the race-enabled syncer test reports no races.

Run race check: `go test -race ./internal/syncer`

- [ ] **Step 6: Commit synchronization**

```powershell
git add internal/syncer
git commit -m "feat: add resilient multi-account synchronization"
```

## Task 11: Implement analytics queries and comparisons

**Files:**
- Create: `internal/analytics/types.go`
- Create: `internal/analytics/query.go`
- Create: `internal/analytics/compare.go`
- Create: `internal/analytics/query_test.go`

- [ ] **Step 1: Write failing seeded-database tests**

Seed UTC records spanning DST-independent Asia/Shanghai hour boundaries. Assert all-account totals, single-account isolation, hourly and daily buckets, yesterday and last-week matches, zero comparison bases returning null percentages, separate credit/debit/net values, and filters for raw reason, business category, direction, remark keyword, ref type, ref ID, model, API key, endpoint, request type and billing mode.

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/analytics -v`

Expected: FAIL because analytics service functions are undefined.

- [ ] **Step 3: Implement filter parsing and parameterized SQL**

All SQL values use bound parameters. Escape `%`, `_`, and the escape character for remark searches. Restrict sort keys and granularities to enumerated values.

- [ ] **Step 4: Implement comparison joining**

Represent percentage as `*float64`; return nil when the baseline is zero. Associate each primary bucket with `bucket-24h` and `bucket-7d` in the configured analysis location rather than subtracting display labels.

- [ ] **Step 5: Verify analytics**

Run:

```powershell
go test ./internal/analytics -v
go test ./...
```

Expected: all tests PASS and seeded totals exactly match hand-calculated fixtures.

- [ ] **Step 6: Commit analytics**

```powershell
git add internal/analytics
git commit -m "feat: add multi-account usage and ledger analytics"
```

## Task 12: Add protected HTTP API and application composition

**Files:**
- Create: `internal/httpapi/router.go`
- Create: `internal/httpapi/auth_handlers.go`
- Create: `internal/httpapi/analytics_handlers.go`
- Create: `internal/httpapi/httpapi_test.go`
- Modify: `internal/app/app.go`
- Modify: `cmd/ai-pixel-usage-analysis/main.go`

- [ ] **Step 1: Write failing HTTP integration tests**

Test public health endpoints, SPA redirect to login, protected analytics returning 401, successful login Cookie attributes, invalid login throttling, session recovery, logout, Origin rejection on writes, valid overview response, account status, and manual sync.

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/httpapi -v`

Expected: FAIL because `NewRouter` is undefined.

- [ ] **Step 3: Implement routes and middleware with standard `http.ServeMux`**

Register the exact routes from the design. Use a JSON envelope `{ "data": value }` for success and `{ "error": { "code": code, "message": message } }` for errors. Do not include SQL, credentials, tokens, or upstream response bodies in client errors.

- [ ] **Step 4: Compose startup in the required security order**

Startup order must be:

```text
load TOML -> ensure .env dashboard credentials -> load account credentials -> open/migrate SQLite ->
sync dashboard hash and revoke sessions if changed -> construct upstream clients -> construct syncer and analytics ->
start local HTTP listener -> start scheduler -> wait for signal -> graceful shutdown
```

Log generated credential file path without logging values. Fail startup if `.env` generation, config, database, or listener setup fails.

- [ ] **Step 5: Verify API and application build**

Run:

```powershell
go test ./internal/httpapi ./internal/app ./cmd/ai-pixel-usage-analysis -v
go test ./...
go vet ./...
```

Expected: all tests and vet exit 0.

- [ ] **Step 6: Commit HTTP service**

```powershell
git add internal/httpapi internal/app cmd/ai-pixel-usage-analysis
git commit -m "feat: expose authenticated analytics service"
```

## Task 13: Build the authenticated React application shell

**Files:**
- Create: `web/src/api/client.ts`
- Create: `web/src/api/types.ts`
- Create: `web/src/pages/LoginPage.tsx`
- Create: `web/src/components/AppShell.tsx`
- Create: `web/src/pages/LoginPage.test.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/styles.css`

- [ ] **Step 1: Write failing login and session tests**

Use Testing Library to assert session loading, unauthenticated login, generic invalid-credential error, authenticated navigation, logout, keyboard focus, label associations and mobile layout classes.

- [ ] **Step 2: Run tests and verify failure**

Run: `Push-Location web; npm test -- --run; Pop-Location`

Expected: FAIL because the login page and API client do not exist.

- [ ] **Step 3: Implement the app shell and design system**

Use same-origin `fetch` with `credentials: "same-origin"`. Define an airy dashboard with high-contrast typography, restrained blue/teal accents, semantic success/warning/error colors, 44px minimum controls, visible focus rings, responsive navigation and no externally hosted fonts or assets.

- [ ] **Step 4: Verify frontend shell**

Run:

```powershell
Push-Location web
npm test -- --run
npm run build
Pop-Location
```

Expected: tests PASS and production assets are written to `internal/webui/dist`.

- [ ] **Step 5: Commit frontend shell**

```powershell
git add web internal/webui/dist
git commit -m "feat: add authenticated dashboard shell"
```

## Task 14: Implement overview, usage, ledger, and account-status pages

**Files:**
- Create: `web/src/pages/OverviewPage.tsx`
- Create: `web/src/pages/UsagePage.tsx`
- Create: `web/src/pages/LedgerPage.tsx`
- Create: `web/src/pages/AccountsPage.tsx`
- Create: `web/src/components/GlobalFilters.tsx`
- Create: `web/src/components/KpiCard.tsx`
- Create: `web/src/components/TrendChart.tsx`
- Create: `web/src/components/DataTable.tsx`
- Create: `web/src/components/StatusNotice.tsx`
- Create: `web/src/pages/analytics.test.tsx`

- [ ] **Step 1: Write failing page tests**

Mock API responses and assert: all/single account switch, hourly/day granularity, independent yesterday/last-week toggles, null comparison display, request dimensions, separate credit/debit/net KPIs, raw ledger type filter, business category filter, direction filter, remark search, reference filters, metadata drill-down, partial-data warning, and manual sync state.

- [ ] **Step 2: Run tests and verify failure**

Run: `Push-Location web; npm test -- --run; Pop-Location`

Expected: FAIL because the pages and components do not exist.

- [ ] **Step 3: Implement shared filters, charts, and tables**

Encode filter state in URL search parameters. Abort superseded requests. ECharts series must give primary, yesterday, and last-week lines distinct stroke patterns in addition to color. Tables must support keyboard-accessible drill-down and server pagination.

- [ ] **Step 4: Implement the four pages and all data states**

Each page must render loading skeleton, empty state, partial-account warning, recoverable request error, and retry action. Ledger amounts must never combine credit and debit into a single unsigned value.

- [ ] **Step 5: Verify frontend analytics**

Run:

```powershell
Push-Location web
npm test -- --run
npm run build
Pop-Location
```

Expected: all tests PASS and the build contains no network-hosted runtime dependencies.

- [ ] **Step 6: Commit analytics UI**

```powershell
git add web internal/webui/dist
git commit -m "feat: add usage and ledger analysis dashboard"
```

## Task 15: Add Nginx, systemd, packaging, and README instructions

**Files:**
- Create: `deploy/nginx.conf.example`
- Create: `deploy/ai-pixel-usage-analysis.service`
- Modify: `docs/packaging.md`
- Modify: `docs/acceptance.md`
- Modify: `README.md`
- Modify: `scripts/build.ps1`
- Modify: `Makefile`

- [ ] **Step 1: Write deployment configurations**

Nginx must redirect HTTP to HTTPS, proxy to `127.0.0.1:8080`, set `Host`, `X-Real-IP`, `X-Forwarded-For`, and `X-Forwarded-Proto`, set bounded body size and timeouts, and add HSTS, content-type, frame and referrer protections. systemd must run as a dedicated unprivileged user, set the working directory containing `config.toml` and `.env`, restart on failure, and use filesystem hardening compatible with SQLite writes.

- [ ] **Step 2: Complete build and operator documentation**

README and packaging docs must include exact commands:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 test
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 build
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 build-linux
```

Document the three deployed files/directories: executable, `config.toml`, `.env`, plus the writable SQLite data path. State that Node is a build dependency only and Nginx is external server infrastructure.

- [ ] **Step 3: Add a packaging smoke target**

`powershell -ExecutionPolicy Bypass -File scripts/build.ps1 verify` must run frontend tests/build, Go tests, vet, Linux build, start a native Windows binary against a temporary config/database, poll `/health/ready`, and stop it cleanly. The Linux artifact is compiled separately and is not executed on Windows.

- [ ] **Step 4: Verify documentation and packaging files**

Run:

```powershell
rg -n "proxy_pass|X-Forwarded-Proto|Strict-Transport-Security" deploy/nginx.conf.example
rg -n "DASHBOARD_USERNAME|DASHBOARD_PASSWORD|0600|CGO_ENABLED=0|Nginx" README.md docs/packaging.md
git diff --check
```

Expected: all required deployment strings are present and diff check exits 0.

- [ ] **Step 5: Commit deployment material**

```powershell
git add deploy docs/packaging.md docs/acceptance.md README.md scripts/build.ps1 Makefile
git commit -m "docs: add Linux and Nginx deployment workflow"
```

## Task 16: Run browser acceptance and complete final verification

**Files:**
- Create: `web/e2e/dashboard.spec.ts`
- Create: `web/playwright.config.ts`
- Modify: `web/package.json`
- Modify: `docs/acceptance.md`

- [ ] **Step 1: Write end-to-end scenarios against deterministic fixture data**

Scenarios must log in, verify 24-hour Cookie metadata through the test server, switch all/single account, toggle both comparisons independently, filter request dimensions, filter ledger raw type/direction/remark, open metadata detail, observe partial-account warning, trigger manual sync, and log out.

- [ ] **Step 2: Run E2E tests and verify the red state before fixture wiring**

Run: `Push-Location web; npm run test:e2e; Pop-Location`

Expected: FAIL because the deterministic test server command is not yet wired.

- [ ] **Step 3: Add the deterministic test-server mode**

The Go test server must use a temporary SQLite database and fixed clock, seed two accounts plus usage and ledger fixtures, bind an ephemeral loopback port, and never load the real `.env` or call real upstream hosts.

- [ ] **Step 4: Run the complete verification matrix**

Run:

```powershell
Push-Location web
npm test -- --run
npm run build
npm run test:e2e
Pop-Location
go test ./...
go test -race ./internal/auth ./internal/syncer ./internal/analytics ./internal/httpapi
go vet ./...
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 build-linux
```

Expected: all frontend tests, E2E scenarios, Go tests, race checks, vet and Linux build exit 0.

- [ ] **Step 5: Inspect the Linux artifact and requirement checklist**

Run:

```powershell
Get-Item dist/ai-pixel-usage-analysis-linux-amd64 | Select-Object Name,Length
rg -n "PASS|命令|场景" docs/acceptance.md
git status --short
```

Expected: one Linux executable exists, every acceptance row names a verification command or scenario, and only intended final files are modified.

- [ ] **Step 6: Commit end-to-end acceptance**

```powershell
git add web/e2e web/playwright.config.ts web/package.json web/package-lock.json docs/acceptance.md
git commit -m "test: add end-to-end acceptance coverage"
```

- [ ] **Step 7: Push the development branch with the project-local SSH configuration**

Run:

```powershell
git status --short --branch
git log --oneline origin/main..HEAD
git push -u origin codex/complete-project
git ls-remote origin refs/heads/codex/complete-project
```

Expected: push succeeds without `gh`, and `ls-remote` reports the same commit as local `HEAD`.
