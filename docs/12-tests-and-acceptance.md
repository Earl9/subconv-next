# 12. 测试与验收

## 测试分层

### Unit Tests

必须覆盖：

- UCI parser
- Base64 detector
- URI parser
- YAML parser
- NodeIR normalize
- dedupe
- Mihomo renderer
- SSRF IP 判断
- atomic write

### Golden Tests

必须覆盖：

- lite template
- standard template
- VLESS Reality
- Hysteria2
- TUIC
- AnyTLS
- WireGuard URI
- WireGuard standard config
- duplicate name rename
- empty region group not generated

### API Tests

使用 `httptest` 覆盖：

- `/healthz`
- `/api/status`
- `/api/parse`
- `/api/generate`
- `/api/refresh`
- `/sub/mihomo.yaml`

## Testdata 结构

```text
testdata/
├── config/
│   ├── basic.json
│   └── basic.uci
├── nodes/
│   ├── ss.txt
│   ├── vmess.txt
│   ├── vless-reality.txt
│   ├── trojan-ws.txt
│   ├── hy2.txt
│   ├── tuic-v5.txt
│   ├── anytls.txt
│   ├── wireguard-uri.txt
│   └── wireguard-conf.conf
├── subscriptions/
│   ├── mixed-uri-list.txt
│   ├── base64-mixed.txt
│   └── mihomo.yaml
└── golden/
    ├── lite-basic.yaml
    ├── standard-vless-reality.yaml
    ├── standard-hy2-tuic.yaml
    ├── standard-anytls.yaml
    └── standard-wireguard.yaml
```

## 必须通过的命令

```sh
go test ./...
go test ./internal/parser -run TestParse
go test ./internal/renderer -run Golden
go run ./cmd/subconv-next version
```

## 本地集成测试

```sh
go run ./cmd/subconv-next serve --config ./testdata/config/basic.json
curl -fsS http://127.0.0.1:9876/healthz
curl -fsS http://127.0.0.1:9876/api/status
curl -fsS http://127.0.0.1:9876/sub/mihomo.yaml
```

## OpenWrt 验收

在 OpenWrt 设备或容器中：

```sh
opkg install ./subconv-next_*.ipk
/etc/init.d/subconv-next enable
/etc/init.d/subconv-next start
curl -fsS http://127.0.0.1:9876/healthz
uci show subconv_next
logread | grep subconv
```

## LuCI 验收

- 菜单存在。
- 能保存配置。
- 能添加多个 subscription。
- 能点击刷新。
- 能下载 YAML。
- 日志页不泄露完整 token。

## 性能验收

在低端设备目标：

- 1000 个节点解析与渲染 < 5 秒。
- daemon idle RSS 尽量 < 30 MB。
- 订阅刷新过程中 RSS 尽量 < 80 MB。
- 输出 YAML 大小合理。

## 回归测试要求

每新增一个协议字段映射，必须新增：

1. parser test。
2. renderer golden。
3. 文档说明。


## AnyTLS / WireGuard V1 强制验收

V1 不能把 AnyTLS 或 WireGuard 标记为 experimental。以下测试必须通过：

```sh
go test ./internal/parser -run 'AnyTLS|WireGuard'
go test ./internal/renderer -run 'AnyTLS|WireGuard|Golden'
```

验收点：

- `anytls://` 能解析为 `type=anytls`。
- AnyTLS renderer 输出 `type: anytls`、`password`、`sni`、`client-fingerprint`、`udp`。
- AnyTLS 不输出 `reality-opts`。
- `wireguard://` 能解析为 `type=wireguard`。
- WireGuard 标准 `[Interface]` / `[Peer]` 配置能解析。
- WireGuard renderer 输出 `private-key`、`public-key`、`ip`、`allowed-ips`、`udp`。
