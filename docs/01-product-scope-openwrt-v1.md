# 01. 产品范围：OpenWrt V1

## 一句话定位

`subconv-next` 是运行在 OpenWrt 上的现代订阅转换与 Mihomo 配置生成工具。

它不是代理软件，也不是机场，也不是流量转发器。

## 用户场景

用户已经有自己的订阅链接或节点链接，希望在 OpenWrt 上生成 Mihomo 可用配置，并通过 LuCI 管理。

典型流程：

1. 用户安装 `subconv-next` 和 `luci-app-subconv-next`。
2. 用户进入 LuCI → Services → SubConv Next。
3. 添加一个或多个订阅源。
4. 选择模板：lite、standard、full。
5. 点击“刷新并生成”。
6. 得到本地订阅地址：

```text
http://127.0.0.1:9876/sub/mihomo.yaml
```

或在局域网使用：

```text
http://192.168.1.1:9876/sub/mihomo.yaml
```

默认 V1 只监听本机，允许用户手动开启 LAN 监听。

## V1 功能

### 订阅源

支持：

- 远程 HTTP/HTTPS 订阅 URL。
- 本地粘贴的 Base64 订阅。
- 本地粘贴的节点 URI。
- Clash/Mihomo YAML。

### 节点协议

V1 必须支持：

- Shadowsocks：`ss://`
- VMess：`vmess://`
- VLESS：`vless://`
- Trojan：`trojan://`
- Hysteria2：`hysteria2://`、`hy2://`
- TUIC v5：`tuic://`
- AnyTLS：`anytls://`
- WireGuard：`wireguard://` 以及 WireGuard 标准配置片段

注意：AnyTLS 和 WireGuard 是 V1 强制支持项，不允许标记为 experimental，不允许只在 Raw 中保留而不输出 Mihomo 可用配置。

### 输出格式

V1 只输出：

- Mihomo YAML。

不输出：

- Surge
- Loon
- Quantumult X
- Shadowrocket
- sing-box JSON

sing-box JSON 放 V2。

## OpenWrt 约束

- 默认监听 `127.0.0.1:9876`。
- 默认不向 WAN 暴露。
- 默认不写大量日志到 flash。
- 运行时缓存放 `/var/run/subconv-next`。
- 配置放 `/etc/config/subconv_next`。
- 持久状态尽量少写。
- 支持 `reload_service`。
- 支持 `restart`。
- 支持开机启动。
- 低内存设备上可运行。

## 不做的事

- 不安装或管理 Mihomo 内核。
- 不修改防火墙。
- 不创建 TUN。
- 不做透明代理。
- 不接管 DNS。
- 不测试节点真实可用性。
- 不抓取公共规则站点之外的未知远程配置。
