# AI Pixel Usage Analysis

多账户 AI Pixel 请求用量与余额流水分析服务。Go 每分钟采集上游数据到 SQLite，React 前端提供全账户汇总、单账户查看、昨天/上周同期、流水类型和备注筛选。生产环境由 Nginx 提供公网 HTTPS，Go 只监听 config.toml 指定的本机端口。

## 功能

- 多账户独立认证、节点优先级、故障切换和增量游标
- 请求、Token、成本、端点和模型分析
- 余额 credit/debit 分离、原始流水类型、业务分类和备注检索
- 小时/日聚合，昨天同期与上周同期
- Argon2id 登录保护和 24 小时 HttpOnly 会话
- Linux CGO-free 单二进制，前端已内嵌

## 配置

config.toml 保存监听端口、时区、上游节点和账户元数据，不保存网页登录凭据。每个 account 的 email_env/password_env 指向 .env 中的键。

.env 示例只展示键名：

~~~dotenv
user1="account@example.com"
password1="upstream-password"
DASHBOARD_USERNAME="dashboard-admin"
DASHBOARD_PASSWORD="dashboard-password"
~~~

如果后两个键缺失，启动时会安全随机生成并写回 .env。Linux 上 .env 必须为 0600；SQLite 仅保存 Argon2id 独立盐哈希。不要提交 .env。

## 开发与测试

~~~powershell
npm --prefix web ci
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 test
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 build
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 build-linux
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 verify
npm --prefix web run test:e2e
~~~

本机运行：

~~~powershell
go run ./cmd/ai-pixel-usage-analysis -config config.toml -env .env -database data/analysis.db
~~~

## Linux 部署

1. 执行 build-linux，复制 dist/ai-pixel-usage-analysis-linux-amd64、config.toml 和 .env 到服务器。
2. 创建专用 ai-pixel 用户和可写 data 目录，设置 .env 权限 0600。
3. 参考 deploy/ai-pixel-usage-analysis.service 安装 systemd。
4. 参考 deploy/nginx.conf.example 配置域名、证书和 Go 端口。
5. 检查 /health/ready 后再开放访问。

服务器不需要 Node.js、Go 或 SQLite 命令。Nginx 属于服务器入口基础设施，不打入应用二进制。

## 文档

- [PRD](docs/prd.md)
- [UI](docs/ui.md)
- [TRD](docs/trd.md)
- [结构](docs/structure.md)
- [开发规范](docs/development.md)
- [验收](docs/acceptance.md)
- [打包部署](docs/packaging.md)
- [设计规格](docs/superpowers/specs/2026-07-11-ai-pixel-usage-analysis-design.md)
- [实施计划](docs/superpowers/plans/2026-07-11-ai-pixel-usage-analysis.md)

## Git

本项目使用仓库 local Git 身份及 ~/.ssh 中的项目专用密钥，直接执行 git push；不要使用 gh 代替项目账户。
