# Changelog

## v1.1.1 — 2026-11-01

- 修复订阅节点删除后因上游地址或凭据变化而重新出现的问题，并完善批量删除、恢复与选择操作。
- 修复读取并保存配置时，脱敏占位符覆盖原始订阅凭据的问题；空订阅响应保留已有有效缓存。
- 规范 Hysteria2 混淆配置，清理旧指纹字段，避免生成无效配置。
- 通过有界解析缓存、节点深拷贝优化和分批 YAML 编码减少重复处理与内存分配；节点列表直接携带校验结果，减少前端重复请求。
- 在同一台 arm64 路由器的 1000 节点测试中，列表响应中位数由 331 ms 降至 72 ms，配置生成由 4.25 s 降至 1.43 s。
- 修复 portable OpenWrt IPK 内文件的属主，统一为 root。
- 已通过 Go 测试、竞态检测、静态检查、前端运行时测试和真实路由器浏览器回归。

## 2026-04-26

- 将 AnyTLS 从 experimental 改为 V1 必选协议。
- 将 WireGuard 从 experimental 改为 V1 必选协议。
- 增加 `uri_anytls.go`、`uri_wireguard.go`、`wgconf.go` 任务要求。
- 增加 AnyTLS parser 字段映射、renderer 输出要求和 golden tests。
- 增加 WireGuard URI 与标准配置片段 parser 要求、renderer 输出要求和 golden tests。
- 更新验收标准：AnyTLS / WireGuard 不允许推迟到 V2，不允许只保留 Raw。
