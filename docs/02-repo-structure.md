# 02. 仓库结构

## 目标结构

```text
subconv-next/
├── cmd/
│   └── subconv-next/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   └── middleware.go
│   ├── config/
│   │   ├── config.go
│   │   ├── uci.go
│   │   └── json.go
│   ├── fetcher/
│   │   ├── fetcher.go
│   │   ├── ssrf.go
│   │   └── cache.go
│   ├── parser/
│   │   ├── detect.go
│   │   ├── base64.go
│   │   ├── yaml.go
│   │   ├── uri_ss.go
│   │   ├── uri_vmess.go
│   │   ├── uri_vless.go
│   │   ├── uri_trojan.go
│   │   ├── uri_hy2.go
│   │   ├── uri_tuic.go
│   │   ├── uri_anytls.go
│   │   ├── uri_wireguard.go
│   │   ├── wgconf.go
│   │   └── parser_test.go
│   ├── model/
│   │   └── node.go
│   ├── renderer/
│   │   ├── mihomo.go
│   │   ├── groups.go
│   │   ├── rules.go
│   │   └── golden_test.go
│   ├── scheduler/
│   │   └── scheduler.go
│   └── storage/
│       ├── state.go
│       └── atomic_file.go
├── package/
│   └── openwrt/
│       ├── subconv-next/
│       │   ├── Makefile
│       │   └── files/
│       │       ├── etc/config/subconv_next
│       │       ├── etc/init.d/subconv-next
│       │       └── usr/share/subconv-next/templates/
│       └── luci-app-subconv-next/
│           ├── Makefile
│           └── root/
│               ├── usr/share/luci/menu.d/luci-app-subconv-next.json
│               ├── usr/share/rpcd/acl.d/luci-app-subconv-next.json
│               └── www/luci-static/resources/view/subconv-next/
├── templates/
│   ├── lite.yaml
│   ├── standard.yaml
│   └── full.yaml
├── testdata/
│   ├── config/
│   ├── subscriptions/
│   ├── nodes/
│   └── golden/
├── docs/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 包职责

### `cmd/subconv-next`

负责 CLI 入口：

```sh
subconv-next serve --config /etc/config/subconv_next
subconv-next generate --config /etc/config/subconv_next --out /tmp/mihomo.yaml
subconv-next parse --input ./sub.txt
subconv-next version
```

### `internal/model`

只放跨模块共享模型，例如 `NodeIR`、`Config`、`RenderOptions`。

### `internal/parser`

负责将任何输入解析成 `[]model.NodeIR`。

禁止 parser 直接输出 Mihomo YAML。

### `internal/renderer`

负责将 `[]NodeIR` 和模板选项渲染成 Mihomo YAML。

禁止 renderer 拉取网络订阅。

### `internal/fetcher`

负责订阅下载、SSRF 防护、大小限制、超时和缓存。

禁止 fetcher 解析节点。

### `internal/api`

本地 HTTP API。

### `package/openwrt`

OpenWrt 打包文件。不要把业务逻辑写在 shell 里。
