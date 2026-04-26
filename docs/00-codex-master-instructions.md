# 00. Codex 总执行说明

## 角色

你是代码执行 agent。你要在当前仓库中实现 `subconv-next`，优先完成 OpenWrt V1。

## 执行原则

1. 不要问产品方向问题，按本文档实现。
2. 每次变更后运行相关测试。
3. 优先让核心库可在普通 Linux 上跑通，再接入 OpenWrt 包。
4. 不要引入大型 Web 框架。
5. 不要引入 subconverter。
6. 不要实现代理、隧道、流量转发、绕过、防火墙接管功能。
7. 所有敏感信息日志脱敏。
8. 默认只监听 `127.0.0.1`。
9. 所有网络拉取必须带 SSRF 防护、大小限制、超时限制。
10. 输出 YAML 必须 deterministic，同样输入得到同样输出。

## 推荐实现顺序

### Phase 1：普通 Linux 可跑

- 初始化 Go module。
- 实现 `subconv-next serve`。
- 实现 `/healthz`。
- 实现 UCI-like JSON 配置加载，方便本地测试。
- 实现 parser IR。
- 实现至少 `ss://`、`trojan://`、`vless://`、`hy2://`、`anytls://`、`wireguard://` parser。
- 实现 Mihomo renderer。
- 实现 golden tests。

### Phase 2：OpenWrt 服务

- 增加 `/etc/config/subconv_next` 读取。
- 增加 procd init script。
- 增加 OpenWrt package Makefile。
- 增加默认模板和安装路径。

### Phase 3：LuCI

- 增加 LuCI app 包。
- 实现订阅源管理、刷新按钮、状态页、日志页、下载配置。
- 增加 rpcd ACL。

### Phase 4：发布

- 增加 GitHub Actions 构建。
- 增加 release artifacts。
- 增加安装说明。

## 代码风格

- Go 代码使用标准库优先。
- YAML 使用 `gopkg.in/yaml.v3`。
- HTTP server 使用标准库 `net/http`。
- CLI 可用标准库 `flag`，不要一开始引入 Cobra。
- 测试使用 Go 标准 testing。
- 所有 parser 必须有 table-driven tests，V1 强制协议包括 AnyTLS 和 WireGuard。
- 所有 renderer 必须有 golden tests。

## 完成标准

当以下命令通过时，才算完成 V1：

```sh
go test ./...
go run ./cmd/subconv-next serve --config ./testdata/config/basic.json
curl -fsS http://127.0.0.1:9876/healthz
curl -fsS http://127.0.0.1:9876/sub/mihomo.yaml
```

OpenWrt 包层面的完成标准：

```sh
# 在 OpenWrt SDK 中
./scripts/feeds update -a
./scripts/feeds install -a
make menuconfig
make package/subconv-next/compile V=s
make package/luci-app-subconv-next/compile V=s
```
