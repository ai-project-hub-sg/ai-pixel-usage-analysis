# 打包与部署

~~~powershell
pwsh -File scripts/build.ps1 test
pwsh -File scripts/build.ps1 build-linux
~~~

脚本构建 React 后使用 CGO_ENABLED=0、GOOS=linux、GOARCH=amd64 生成 dist/ai-pixel-usage-analysis-linux-amd64。Node 是构建依赖，服务器不需要 Node 或 SQLite 命令。

部署可执行文件、config.toml、权限 0600 的 .env 和可写 SQLite 数据目录。若 .env 缺少 DASHBOARD_USERNAME 或 DASHBOARD_PASSWORD，服务生成缺失值并原子写回，管理员直接读取文件，日志不打印值。

Nginx 是公网 HTTPS 入口，代理到 config.toml 中 Go 的 127.0.0.1 端口。使用 deploy 目录中的 Nginx 和 systemd 示例。升级前备份二进制、配置、.env 和 SQLite；替换后启动并检查 /health/ready，迁移失败则使用备份回滚。
