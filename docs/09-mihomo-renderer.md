# 09. Mihomo Renderer

## 目标

输入：

```go
[]model.NodeIR
RenderOptions
```

输出：

```yaml
mixed-port: 7890
allow-lan: false
mode: rule
log-level: info
dns: ...
proxies: ...
proxy-groups: ...
rule-providers: ...
rules: ...
```

## RenderOptions

```go
type RenderOptions struct {
    Template     string
    MixedPort    int
    AllowLAN     bool
    Mode         string
    LogLevel     string
    IPv6         bool
    DNSEnabled   bool
    EnhancedMode string
}
```

## 模板

V1 提供三套：

### lite

适合新手：

- 节点选择
- 自动选择
- DIRECT
- REJECT
- MATCH

### standard

默认：

- 节点选择
- 自动选择
- 故障转移
- 香港
- 日本
- 美国
- 新加坡
- AI
- 流媒体
- Telegram
- GitHub
- Microsoft
- Apple
- 国内直连
- 漏网之鱼

### full

V1 可先复制 standard，并预留更多规则组。

## 代理组生成

基于节点 tags 和名称 regex。

地区识别：

```text
HK: 香港|港|HK|Hong Kong
JP: 日本|日|JP|Japan
US: 美国|美|US|United States
SG: 新加坡|狮城|SG|Singapore
TW: 台湾|台|TW|Taiwan
KR: 韩国|韩|KR|Korea
```

基础组：

```yaml
proxy-groups:
  - name: 节点选择
    type: select
    proxies:
      - 自动选择
      - 故障转移
      - DIRECT
      - <all nodes>

  - name: 自动选择
    type: url-test
    proxies:
      - <all nodes>
    url: https://www.gstatic.com/generate_204
    interval: 300

  - name: 故障转移
    type: fallback
    proxies:
      - <all nodes>
    url: https://www.gstatic.com/generate_204
    interval: 300
```

如果某地区节点为空，不生成该地区组。

## 节点字段映射

### VLESS Reality

Mihomo 输出：

```yaml
- name: example
  type: vless
  server: example.com
  port: 443
  uuid: xxxx
  network: tcp
  tls: true
  udp: true
  flow: xtls-rprx-vision
  servername: example.com
  client-fingerprint: chrome
  reality-opts:
    public-key: xxx
    short-id: xxx
```

### Hysteria2

```yaml
- name: hy2
  type: hysteria2
  server: example.com
  port: 443
  password: xxx
  sni: example.com
  skip-cert-verify: false
  obfs: salamander
  obfs-password: xxx
  udp: true
```

### TUIC v5

```yaml
- name: tuic
  type: tuic
  server: example.com
  port: 443
  uuid: xxx
  password: xxx
  sni: example.com
  udp-relay-mode: native
  congestion-controller: bbr
  reduce-rtt: true
  udp: true
```


### AnyTLS

```yaml
- name: anytls
  type: anytls
  server: example.com
  port: 443
  password: xxx
  client-fingerprint: chrome
  udp: true
  idle-session-check-interval: 30
  idle-session-timeout: 30
  min-idle-session: 0
  sni: example.com
  alpn:
    - h2
    - http/1.1
  skip-cert-verify: false
```

Renderer 要求：

- `type: anytls`。
- `password` 来自 `auth.password`。
- `client-fingerprint` 来自 `tls.clientFingerprint`。
- `sni`、`alpn`、`skip-cert-verify` 来自 `tls`。
- `idle-session-*` 和 `min-idle-session` 来自 Raw。
- 不输出 `reality-opts`，因为 Mihomo 不支持 AnyTLS + Reality。

### WireGuard

简化单 peer 输出：

```yaml
- name: wg
  type: wireguard
  server: example.com
  port: 51820
  ip: 172.16.0.2
  ipv6: fd00::2
  private-key: CLIENT_PRIVATE_KEY
  public-key: SERVER_PUBLIC_KEY
  allowed-ips:
    - 0.0.0.0/0
    - ::/0
  pre-shared-key: PSK
  reserved:
    - 209
    - 98
    - 59
  persistent-keepalive: 25
  udp: true
  mtu: 1280
  remote-dns-resolve: true
  dns:
    - 1.1.1.1
```

多 peer 输出：

```yaml
- name: wg
  type: wireguard
  ip: 172.16.0.2
  ipv6: fd00::2
  private-key: CLIENT_PRIVATE_KEY
  peers:
    - server: example.com
      port: 51820
      public-key: SERVER_PUBLIC_KEY
      allowed-ips:
        - 0.0.0.0/0
  udp: true
```

Renderer 要求：

- 单 peer 可以使用简化写法。
- 多 peer 使用 `peers`。
- `allowed-ips` 默认补 `0.0.0.0/0`。
- `udp` 默认 true。
- `reserved` 支持数组和字符串两种输入；数组优先。
- 支持 `remote-dns-resolve` 与 `dns`。
- 支持透传 `amnezia-wg-option`，如果 Raw 或 WireGuardOptions 中存在。

## Rule Providers

V1 使用内置规则占位，避免依赖复杂外部规则仓库。

默认 rules：

```yaml
rules:
  - GEOSITE,private,DIRECT
  - GEOIP,private,DIRECT,no-resolve
  - GEOSITE,cn,DIRECT
  - GEOIP,CN,DIRECT
  - MATCH,节点选择
```

V1 可预留 rule-providers 配置，但不强制拉取外部规则。

## YAML 要求

- 输出顺序稳定。
- 字段名使用 Mihomo 风格 kebab-case。
- 空字段不输出。
- `proxies` 中节点顺序保持订阅顺序。
- 所有 name 必须唯一。
- 重名节点自动加后缀：`name`, `name 2`, `name 3`。
- 敏感字段只出现在 YAML，不出现在日志。

## Golden Tests

为以下情况写 golden：

```text
testdata/golden/lite-basic.yaml
testdata/golden/standard-vless-reality.yaml
testdata/golden/standard-hy2-tuic.yaml
testdata/golden/standard-anytls.yaml
testdata/golden/standard-wireguard.yaml
testdata/golden/dedupe-renamed.yaml
```

测试命令：

```sh
go test ./internal/renderer -run Golden
```
