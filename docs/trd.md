# 技术设计文档

## 架构

React/TypeScript 经 Vite 构建后由 go:embed 打入 Go。Go 提供 HTTP API、SQLite、认证、调度和上游客户端。Nginx 终止公网 HTTPS 并代理到 config.toml 指定的本机端口。

## 配置与安全

TOML 只保存非敏感 server、analysis、auth 会话时长、host 和 account 元数据。网页登录凭据仅来自 .env；缺失项通过 crypto/rand、同目录临时文件、fsync、0600 和 rename 安全生成。启动时同步为 SQLite dashboard_users 的 Argon2id 独立盐哈希，变更凭据会撤销旧会话。上游令牌只存内存。

## 上游与同步

客户端调用 /api/v1/settings/public、auth/login、auth/refresh、usage、usage/stats、usage/balance-ledger 及 stats。节点按优先级升序和同权重声明顺序选择，分类错误后有限重试与切换。

无游标账户从分析时区上月第一天同步。增量每分钟从成功游标减重叠窗口读取，完整事务成功后推进游标。同账户防重入，多账户受控并行。

## 数据与分析

SQLite 包含 accounts、usage_records、balance_ledger_entries、sync_cursors、sync_runs、upstream_health、dashboard_users、web_sessions、schema_migrations。业务记录唯一键为账户和上游 ID，已知字段结构化并保留原始 JSON。

流水保留原始 reason，生成业务分类和可检索备注。时间以 UTC 保存，按配置时区聚合；小时桶关联前 24 小时和前 7 天，零基数百分比为 null。SQL 参数化，排序和粒度使用白名单。

## HTTP

健康检查不泄露数据，其余分析 API 需要会话。Cookie 为 HttpOnly、SameSite=Lax、生产 Secure、24 小时。写接口校验 Origin/CSRF，登录限速，Go 只信任本机 Nginx 代理头。
