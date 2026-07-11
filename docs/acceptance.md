# 验收方案

## 需求矩阵

| 要求 | 自动化命令或场景 | 通过条件 |
|---|---|---|
| 多账户初始化与增量恢复 | `go test ./internal/syncer -v` | 从分析时区上月第一天开始，完整分页后推进游标，重叠窗口 UPSERT 无重复 |
| 汇总与单账户隔离 | `go test ./internal/analytics -v`；Playwright“账户”切换 | 全部账户为汇总值，选择单账户后查询带 `account_id` 且无串数据 |
| 两种同期 | analytics 固定时钟测试；Playwright 两个独立开关 | 昨天同期与上周同期分别对应正确时间桶，可独立显示/隐藏 |
| 请求维度 | `npm test -- --run`；Playwright 请求分析场景 | 账户、模型、端点、API Key ID 参数和结果有效 |
| 流水与备注分析 | `go test ./internal/ledger ./internal/analytics -v`；Playwright 流水场景 | credit/debit 分离，原始类型、业务分类、方向、备注、引用类型/ID 可筛选，元数据可展开 |
| 网页凭据 | `go test ./internal/secrets ./internal/auth -v` | 缺失键安全随机生成并原子写回 `.env`；Linux 权限 0600；SQLite 仅存独立盐 Argon2id 哈希 |
| 登录与会话 | `go test ./internal/httpapi ./internal/auth -v`；Playwright 登录/退出 | 未登录 API 返回 401，Cookie 为 HttpOnly/SameSite=Lax 且 Max-Age=86400，退出撤销，登录失败限速 |
| 部分账户失败与同步状态 | syncer 状态回归测试；Playwright 告警/手动同步 | 单账户失败不阻断其他账户；页面显示警告；节点、最后成功时间和脱敏错误落库 |
| Nginx 公网入口 | `rg -n "proxy_pass|X-Forwarded-Proto|Strict-Transport-Security" deploy/nginx.conf.example` | HTTPS 由 Nginx 提供，Go 仅监听 `config.toml` 的本机地址和端口 |
| 单二进制与健康检查 | `powershell -ExecutionPolicy Bypass -File scripts/build.ps1 verify` | 内嵌前端，生成 CGO-free Linux amd64 文件；真实 Windows 二进制使用临时配置/数据库后 `/health/ready` 返回 200 |

## 浏览器场景

`npm --prefix web run test:e2e` 启动隔离的 `cmd/e2e-server`。该服务使用临时 SQLite、固定时钟和两账户固定数据，不读取真实 `.env`，不创建真实上游客户端。桌面 Chromium 和 Pixel 7 视口均执行：

1. 登录并检查 24 小时 Cookie 属性。
2. 检查部分账户失败告警，切换全部/单账户和小时/日粒度。
3. 独立切换昨天同期与上周同期。
4. 提交请求模型、端点和 API Key 筛选，并核对实际请求 URL。
5. 提交流水原始类型、分类、方向、备注和引用筛选，展开规范化元数据。
6. 手动同步账户、检查完成状态并退出登录。
7. 检查页面标题、非空内容、无框架错误覆盖层、除未登录会话探测产生的预期 401 资源日志外无 console warning/error，并保存桌面/移动截图到系统临时目录。

本会话未提供 Browser 插件，因此按前端测试规范使用仓库内 Playwright 回退。

## 2026-07-11 验证记录

- `npm --prefix web test -- --run`：2 个文件、5 个测试通过。
- `npm --prefix web run test:e2e`：desktop-chromium 与 mobile-chromium 共 2 个场景通过。
- `go test ./...`：全部 Go 包通过。
- `go vet ./...`：退出码 0。
- `scripts/build.ps1 verify`：退出码 0，包含前端测试/构建、Go 测试/vet、Windows 原生健康冒烟和 Linux 构建。
- `dist/ai-pixel-usage-analysis-linux-amd64`：16,401,485 字节。
- `git diff --check`：必须在提交前退出码 0。

Race 检查在当前 Windows 工作机未执行成功：显式设置 `CGO_ENABLED=1` 后，Go 报告 `C compiler "gcc" not found`。这属于本机编译器依赖缺失；应在安装 GCC 的 Windows 构建机或 Linux CI 上补跑：

~~~powershell
$env:CGO_ENABLED = "1"
go test -race ./internal/auth ./internal/syncer ./internal/analytics ./internal/httpapi
~~~
