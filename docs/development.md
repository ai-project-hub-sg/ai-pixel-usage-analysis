# 开发规范

使用 Go 1.25、Node.js 24 和 npm 11；Node 只用于构建。每项行为先写失败测试，再最小实现并验证绿色。Go 使用表驱动、httptest 和临时 SQLite；前端使用 Vitest/Testing Library；关键流程使用 Playwright。

~~~powershell
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 test
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 build
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 build-linux
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 verify
~~~

- Go 文件单一职责，错误不得包含秘密。
- SQL 参数化，排序使用白名单，金额不以二进制浮点持久化，时间以 UTC 保存。
- 数据库只追加迁移；.env 永不提交且 Linux 权限 0600；config.toml 不保存网页登录字段。
- 提交前执行 Go 测试、vet、前端测试/构建，并对并发包执行 race 检查。
- 只用项目本地 Git/SSH 推送，不使用 gh。
