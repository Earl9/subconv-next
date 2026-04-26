# 08. 协议解析器

## 输入检测

`parser.Detect(content []byte)` 返回：

```go
type InputKind string

const (
    InputKindURIList InputKind = "uri_list"
    InputKindBase64  InputKind = "base64"
    InputKindYAML    InputKind = "yaml"
    InputKindUnknown InputKind = "unknown"
)
```

检测顺序：

1. YAML：包含 `proxies:` 或可解析为 YAML map。
2. URI list：包含 `://` 且按行分隔。
3. Base64：尝试 base64 decode 后重新 detect。
4. Unknown。

## URI 通用规则

- `#fragment` 作为节点名。
- query params 大小写兼容。
- URL decode 所有参数。
- 对未知参数放入 `Raw`。
- 保留 parser warnings。

## Shadowsocks `ss://`

支持格式：

```text
ss://method:password@host:port#name
ss://base64(method:password)@host:port#name
ss://base64(method:password@host:port)#name
```

NodeIR：

```text
type=ss
server
port
auth.password
raw.method
```

Mihomo renderer 输出：

```yaml
type: ss
cipher: <method>
password: <password>
udp: true
```

## VMess `vmess://`

VMess 常见为 base64 JSON。

字段映射：

| VMess | NodeIR |
|---|---|
| ps | name |
| add | server |
| port | port |
| id | auth.uuid |
| aid | raw.alterId |
| net | transport.network |
| type | raw.headerType |
| host | transport.host |
| path | transport.path |
| tls | tls.enabled |
| sni | tls.sni |
| alpn | tls.alpn |
| fp | tls.clientFingerprint |

## VLESS `vless://`

典型：

```text
vless://uuid@host:port?encryption=none&security=reality&sni=example.com&fp=chrome&pbk=xxx&sid=xxx&type=tcp&flow=xtls-rprx-vision#name
```

字段映射：

| Query | NodeIR |
|---|---|
| uuid username | auth.uuid |
| security=tls/reality | tls.enabled |
| sni/servername | tls.sni |
| fp | tls.clientFingerprint |
| pbk/publicKey | tls.reality.publicKey |
| sid/shortId | tls.reality.shortId |
| spx/spiderX | tls.reality.spiderX |
| type/network | transport.network |
| path | transport.path |
| host | transport.host |
| serviceName | transport.serviceName |
| flow | raw.flow |
| packetEncoding | raw.packetEncoding |

## Trojan `trojan://`

典型：

```text
trojan://password@host:443?sni=example.com&type=ws&host=cdn.com&path=/path#name
```

字段映射：

| Query | NodeIR |
|---|---|
| username | auth.password |
| sni | tls.sni |
| alpn | tls.alpn |
| allowInsecure / insecure | tls.insecure |
| type | transport.network |
| host | transport.host |
| path | transport.path |

## Hysteria2 `hy2://` / `hysteria2://`

典型：

```text
hy2://password@host:443?sni=example.com&insecure=1&obfs=salamander&obfs-password=xxx#name
```

字段映射：

| Query | NodeIR |
|---|---|
| username | auth.password |
| sni | tls.sni |
| insecure / skip-cert-verify | tls.insecure |
| alpn | tls.alpn |
| obfs | raw.obfs |
| obfs-password / obfsPassword | raw.obfsPassword |
| ports | raw.ports |
| hop-interval | raw.hopInterval |
| up | raw.up |
| down | raw.down |

## TUIC v5 `tuic://`

典型：

```text
tuic://uuid:password@host:443?sni=example.com&congestion_control=bbr&udp_relay_mode=native#name
```

字段映射：

| Query | NodeIR |
|---|---|
| username | auth.uuid |
| password | auth.password |
| sni | tls.sni |
| alpn | tls.alpn |
| allow_insecure / insecure | tls.insecure |
| congestion_control / congestion-controller | raw.congestionController |
| udp_relay_mode / udp-relay-mode | raw.udpRelayMode |
| reduce_rtt / reduce-rtt | raw.reduceRTT |


## AnyTLS `anytls://`

典型：

```text
anytls://password@host:443?sni=example.com&insecure=1&alpn=h2,http/1.1&client-fingerprint=chrome#name
```

字段映射：

| Query | NodeIR |
|---|---|
| username | auth.password |
| sni / servername | tls.sni |
| alpn | tls.alpn |
| insecure / skip-cert-verify / allowInsecure | tls.insecure |
| client-fingerprint / fp | tls.clientFingerprint |
| idle-session-check-interval | raw.idleSessionCheckInterval |
| idle-session-timeout | raw.idleSessionTimeout |
| min-idle-session | raw.minIdleSession |
| ech / ech-config | tls.ech.config |

要求：

- `tls.enabled` 默认为 true。
- `udp` 默认为 true。
- Mihomo 不支持 AnyTLS + Reality 组合，因此 parser 如遇 `security=reality`、`pbk`、`sid` 等 Reality 参数，必须加入 warning，并且 renderer 不输出 `reality-opts`。
- 未知 query 参数保留在 Raw。

## WireGuard `wireguard://`

V1 必须支持两种输入：

1. URI 形式。
2. WireGuard 标准配置片段。

### URI 形式

推荐接受以下 URI 形态：

```text
wireguard://private-key@host:51820?public-key=xxx&ip=172.16.0.2&ipv6=fd00::2&allowed-ips=0.0.0.0/0,::/0&reserved=209,98,59&mtu=1280#name
```

字段映射：

| Query | NodeIR |
|---|---|
| username | auth.privateKey |
| public-key / peer-public-key | auth.publicKey |
| pre-shared-key / preshared-key / psk | auth.preSharedKey |
| ip / address | wireguard.ip |
| ipv6 | wireguard.ipv6 |
| allowed-ips / allowedIPs | wireguard.allowedIPs |
| reserved | wireguard.reserved 或 wireguard.reservedString |
| persistent-keepalive | wireguard.persistentKeepalive |
| mtu | wireguard.mtu |
| remote-dns-resolve | wireguard.remoteDNSResolve |
| dns | wireguard.dns |

### WireGuard 标准配置片段

必须支持解析：

```ini
[Interface]
Address = 172.16.0.2/32, fd00::2/128
PrivateKey = CLIENT_PRIVATE_KEY
DNS = 1.1.1.1
MTU = 1280

[Peer]
PublicKey = SERVER_PUBLIC_KEY
PresharedKey = PSK
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = example.com:51820
PersistentKeepalive = 25
```

映射：

| WG Config | NodeIR |
|---|---|
| Interface.PrivateKey | auth.privateKey |
| Interface.Address IPv4 | wireguard.ip |
| Interface.Address IPv6 | wireguard.ipv6 |
| Interface.DNS | wireguard.dns |
| Interface.MTU | wireguard.mtu |
| Peer.Endpoint host | server |
| Peer.Endpoint port | port |
| Peer.PublicKey | auth.publicKey |
| Peer.PresharedKey | auth.preSharedKey |
| Peer.AllowedIPs | wireguard.allowedIPs |
| Peer.PersistentKeepalive | wireguard.persistentKeepalive |

限制：

- V1 至少支持单 peer。
- 如果配置中有多个 peer，可以生成多个 NodeIR，也可以生成一个带 `wireguard.peers` 的 NodeIR；但 Mihomo renderer 必须能输出可用 YAML。
- `allowed-ips` 默认值为 `['0.0.0.0/0']`，除非输入显式提供。
- `udp` 默认为 true。

## Clash/Mihomo YAML

解析 `proxies` 数组。尽量保留原字段到 `Raw`。

支持字段：

- name
- type
- server
- port
- uuid
- password
- username
- cipher
- udp
- tls
- sni
- servername
- client-fingerprint
- reality-opts
- network
- ws-opts
- grpc-opts
- h2-opts
- skip-cert-verify

## 错误处理

Parser 返回 warning，不要因单个坏节点终止全局解析。

错误示例：

```json
{
  "line": 12,
  "kind": "INVALID_URI",
  "message": "missing host"
}
```

## 测试用例

为每种协议建立：

```text
testdata/nodes/ss.txt
testdata/nodes/vmess.txt
testdata/nodes/vless-reality.txt
testdata/nodes/trojan-ws.txt
testdata/nodes/hy2.txt
testdata/nodes/tuic-v5.txt
testdata/nodes/anytls.txt
testdata/nodes/wireguard-uri.txt
testdata/nodes/wireguard-conf.conf
```

每个测试断言：

- name
- type
- server
- port
- key auth field
- key tls field
- key transport field
