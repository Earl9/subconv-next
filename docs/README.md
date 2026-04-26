# SubConv Next OpenWrt V1：Codex 执行文档包

## 项目目标

实现一个 **OpenWrt 优先** 的现代订阅转换工具，暂定名 `subconv-next`。

第一版只做这些事：

1. 在 OpenWrt 上运行一个本地 HTTP daemon。
2. 通过 UCI 保存订阅源、模板、刷新策略和输出配置。
3. 通过 LuCI 页面管理订阅源、手动刷新、查看日志、下载 Mihomo YAML。
4. 支持现代节点解析：`ss://`、`vmess://`、`vless://`、`trojan://`、`hysteria2://` / `hy2://`、`tuic://`。
5. 支持导入 Clash/Mihomo YAML 和 Base64 订阅。
6. 输出 Mihomo YAML。
7. 不依赖 subconverter。
8. 不做代理内核，不提供节点，不转发流量，只做配置生成。

## V1 技术选型

- 后端：Go 单二进制 daemon。
- 配置：OpenWrt UCI，配置文件 `/etc/config/subconv_next`。
- 服务管理：OpenWrt procd init script。
- Web UI：LuCI JavaScript app。
- 本地 API：daemon 监听 `127.0.0.1:9876`。
- 输出文件：`/var/run/subconv-next/mihomo.yaml`。
- 缓存目录：`/var/run/subconv-next/cache/`。
- 持久状态：`/etc/subconv-next/state.json`，只存非敏感状态。
- 日志：syslog + `/var/log/subconv-next.log` 可选。

## 文档执行顺序

Codex 应按以下顺序实现：

1. `00-codex-master-instructions.md`
2. `01-product-scope-openwrt-v1.md`
3. `02-repo-structure.md`
4. `03-openwrt-package.md`
5. `04-daemon-api.md`
6. `05-uci-config-schema.md`
7. `06-subscription-fetch-cache.md`
8. `07-node-ir.md`
9. `08-protocol-parsers.md`
10. `09-mihomo-renderer.md`
11. `10-luci-app.md`
12. `11-security-and-reliability.md`
13. `12-tests-and-acceptance.md`
14. `13-build-release-ci.md`
15. `14-cloudflare-workers-future.md`

## 交付物

V1 完成时仓库应包含：

```text
subconv-next/
├── cmd/subconv-next/
├── internal/
│   ├── api/
│   ├── config/
│   ├── fetcher/
│   ├── parser/
│   ├── renderer/
│   ├── scheduler/
│   └── storage/
├── package/openwrt/subconv-next/
├── package/openwrt/luci-app-subconv-next/
├── templates/
├── testdata/
├── docs/
├── go.mod
├── Makefile
└── README.md
```

## 非目标

V1 不做：

- 节点测速。
- 节点连通性检测。
- 自动安装 Mihomo。
- 自动修改 OpenWrt 防火墙、透明代理、TUN、DNS 劫持。
- 多用户账户。
- 远程云端同步。
- Surge、Loon、Quantumult X、Shadowrocket 输出。
- Cloudflare Workers 部署。

这些放到 V2+。
