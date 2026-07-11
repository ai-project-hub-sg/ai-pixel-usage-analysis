# 项目结构

~~~text
cmd/ai-pixel-usage-analysis/  进程入口
internal/app/                 生命周期和装配
internal/config/              TOML
internal/secrets/             .env
internal/store/               SQLite 迁移
internal/auth/                Argon2id 与会话
internal/upstream/            上游协议和切换
internal/ledger/              流水与备注规范化
internal/syncer/              同步、游标、调度
internal/analytics/           聚合、筛选、同期
internal/httpapi/             路由和 handler
internal/webui/               内嵌前端
web/src/                      React
web/e2e/                      浏览器验收
deploy/                       Nginx 和 systemd
scripts/                      构建验证
docs/                         文档
~~~

配置不访问数据库，上游客户端不写 SQLite，同步器通过窄接口组合二者，HTTP handler 不包含 SQL，前端统一通过同源 API 客户端访问。
