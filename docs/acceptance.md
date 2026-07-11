# 验收方案

| 要求 | 验证 | 通过条件 |
|---|---|---|
| 多账户初始化 | syncer 集成测试 | 两账户从上月第一天同步 |
| 增量恢复 | syncer 测试 | 无重复，失败不推进，重启续传 |
| 汇总/单账户 | analytics 与 E2E | 汇总为各账户之和且无串数据 |
| 两种同期 | 固定时钟测试 | 昨天和上周对应桶准确 |
| 请求维度 | API/E2E | 模型、Key、端点、类型筛选有效 |
| 流水分析 | ledger/analytics/E2E | credit、debit、类型、备注和引用有效 |
| 网页凭据 | secrets/auth 测试 | 缺失项写回 0600，SQLite 仅 Argon2id 哈希 |
| 登录 | auth/HTTP 测试 | 未登录拒绝，24 小时失效，退出撤销 |
| 部分失败 | syncer/E2E | 其他账户更新并显示警告 |
| Nginx | 配置与探测 | Go 监听配置端口，转发正确 |
| 单二进制 | build-linux | CGO-free Linux 文件生成 |

最终执行前端测试、构建、E2E、Go 测试、race、vet、Linux 构建、健康检查和 git diff --check，并记录退出码和失败数。
