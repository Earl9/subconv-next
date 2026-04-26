# 15. Codex 任务拆分清单

## Task 1：初始化项目

- 创建 Go module。
- 创建目录结构。
- 创建 `cmd/subconv-next/main.go`。
- 实现 `version` 命令。
- 增加根 Makefile。
- 验收：`go test ./...` 通过。

## Task 2：配置加载

- 实现 JSON config loader。
- 实现 UCI subset loader。
- 增加 testdata config。
- 验收：JSON 和 UCI 解析结果一致。

## Task 3：HTTP daemon

- 实现 serve 命令。
- 实现 `/healthz`。
- 实现 `/api/status`。
- 实现内存 runtime status。
- 验收：curl 成功。

## Task 4：NodeIR 与 normalize

- 实现 `model.NodeIR`。
- 实现 normalize。
- 实现 dedupe。
- 验收：unit tests。

## Task 5：基础 parser

- 实现 detect。
- 实现 base64 decode。
- 实现 URI list parser。
- 实现 YAML proxies parser。
- 验收：mixed subscription 可解析。

## Task 6：协议 parser

依次实现：

1. ss
2. trojan
3. vless
4. hysteria2
5. tuic
6. anytls
7. wireguard URI
8. WireGuard standard config
9. vmess

每个协议必须有 table-driven test。AnyTLS 和 WireGuard 是 V1 必须支持项，不允许推迟到 V2。

## Task 7：Mihomo renderer

- 实现 `RenderMihomo(nodes, opts)`。
- 实现 lite template。
- 实现 standard template。
- 实现 proxy-groups。
- 实现 region groups。
- 实现 golden tests。

## Task 8：fetcher 与 cache

- 实现 URL fetch。
- 实现 SSRF guard。
- 实现 max body size。
- 实现 cache。
- 实现失败回退缓存。
- 验收：httptest 覆盖 redirect/private IP。

## Task 9：refresh pipeline

- 串联 config → fetch → parse → render → atomic write。
- 实现 `/api/refresh`。
- 实现 `/sub/mihomo.yaml`。
- 验收：本地 integration test。

## Task 10：OpenWrt 包

- 创建 `package/openwrt/subconv-next`。
- 添加 Makefile。
- 添加 init script。
- 添加默认 UCI。
- 添加模板文件。
- 验收：OpenWrt SDK 能打包，或至少文件结构正确。

## Task 11：LuCI App

- 创建 `luci-app-subconv-next`。
- 添加 menu.d。
- 添加 rpcd ACL。
- 添加 overview/subscriptions/render/logs views。
- 验收：LuCI 页面能保存 UCI 并调用 API。

## Task 12：CI 与 Release

- 添加 GitHub Actions。
- 添加多架构构建。
- 添加 checksums。
- 更新 README 安装文档。
