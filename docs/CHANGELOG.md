# Changelog

## 2026-04-26

- 将 AnyTLS 从 experimental 改为 V1 必选协议。
- 将 WireGuard 从 experimental 改为 V1 必选协议。
- 增加 `uri_anytls.go`、`uri_wireguard.go`、`wgconf.go` 任务要求。
- 增加 AnyTLS parser 字段映射、renderer 输出要求和 golden tests。
- 增加 WireGuard URI 与标准配置片段 parser 要求、renderer 输出要求和 golden tests。
- 更新验收标准：AnyTLS / WireGuard 不允许推迟到 V2，不允许只保留 Raw。
