# 07. NodeIR 中间表示

## 目标

任何输入格式都先解析成 `NodeIR`，再渲染成 Mihomo YAML。

Parser 禁止直接生成 YAML。

Renderer 禁止直接解析 URI。

## Go 类型

Codex 创建：

```go
package model

type Protocol string

const (
    ProtocolSS        Protocol = "ss"
    ProtocolVMess     Protocol = "vmess"
    ProtocolVLESS     Protocol = "vless"
    ProtocolTrojan    Protocol = "trojan"
    ProtocolHysteria2 Protocol = "hysteria2"
    ProtocolTUIC      Protocol = "tuic"
    ProtocolAnyTLS    Protocol = "anytls"
    ProtocolWireGuard Protocol = "wireguard"
    ProtocolHTTP      Protocol = "http"
    ProtocolSOCKS5    Protocol = "socks5"
)

type NodeIR struct {
    ID        string                 `json:"id"`
    Name      string                 `json:"name"`
    Type      Protocol               `json:"type"`
    Server    string                 `json:"server"`
    Port      int                    `json:"port,omitempty"`
    Auth      Auth                   `json:"auth,omitempty"`
    TLS       TLSOptions             `json:"tls,omitempty"`
    Transport TransportOptions       `json:"transport,omitempty"`
    WireGuard *WireGuardOptions      `json:"wireguard,omitempty"`
    UDP       *bool                  `json:"udp,omitempty"`
    Tags      []string               `json:"tags,omitempty"`
    Source    SourceInfo             `json:"source,omitempty"`
    Raw       map[string]interface{} `json:"raw,omitempty"`
    Warnings  []string               `json:"warnings,omitempty"`
}

type Auth struct {
    UUID         string `json:"uuid,omitempty"`
    Password     string `json:"password,omitempty"`
    Username     string `json:"username,omitempty"`
    Token        string `json:"token,omitempty"`
    PrivateKey   string `json:"private_key,omitempty"`
    PublicKey    string `json:"public_key,omitempty"`
    PreSharedKey string `json:"pre_shared_key,omitempty"`
}

type TLSOptions struct {
    Enabled           bool            `json:"enabled,omitempty"`
    SNI               string          `json:"sni,omitempty"`
    ALPN              []string        `json:"alpn,omitempty"`
    Insecure          bool            `json:"insecure,omitempty"`
    Fingerprint       string          `json:"fingerprint,omitempty"`
    ClientFingerprint string          `json:"client_fingerprint,omitempty"`
    Reality           *RealityOptions `json:"reality,omitempty"`
    ECH               *ECHOptions     `json:"ech,omitempty"`
}

type RealityOptions struct {
    PublicKey string `json:"public_key,omitempty"`
    ShortID   string `json:"short_id,omitempty"`
    SpiderX   string `json:"spider_x,omitempty"`
}

type ECHOptions struct {
    Enabled bool   `json:"enabled,omitempty"`
    Config  string `json:"config,omitempty"`
}

type TransportOptions struct {
    Network     string            `json:"network,omitempty"`
    Path        string            `json:"path,omitempty"`
    Host        string            `json:"host,omitempty"`
    ServiceName string            `json:"service_name,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
}


type WireGuardOptions struct {
    IP                  string     `json:"ip,omitempty"`
    IPv6                string     `json:"ipv6,omitempty"`
    AllowedIPs          []string   `json:"allowed_ips,omitempty"`
    Reserved            []int      `json:"reserved,omitempty"`
    ReservedString      string     `json:"reserved_string,omitempty"`
    MTU                 int        `json:"mtu,omitempty"`
    PersistentKeepalive  int        `json:"persistent_keepalive,omitempty"`
    RemoteDNSResolve    bool       `json:"remote_dns_resolve,omitempty"`
    DNS                 []string   `json:"dns,omitempty"`
    Peers               []WGPeer   `json:"peers,omitempty"`
    AmneziaWG           map[string]interface{} `json:"amnezia_wg,omitempty"`
}

type WGPeer struct {
    Server       string   `json:"server,omitempty"`
    Port         int      `json:"port,omitempty"`
    PublicKey    string   `json:"public_key,omitempty"`
    PreSharedKey string   `json:"pre_shared_key,omitempty"`
    AllowedIPs   []string `json:"allowed_ips,omitempty"`
    Reserved     []int    `json:"reserved,omitempty"`
}

type SourceInfo struct {
    Name string `json:"name,omitempty"`
    Kind string `json:"kind,omitempty"`
}
```

## ID 生成

ID 不要随机。使用稳定 hash：

```text
sha256(type + server + port + name + auth fingerprint)
```

这样 dedupe 稳定。

## Normalization

解析后执行：

1. trim name。
2. name 为空则用 `type-server-port`。
3. 清理控制字符。
4. server lowercase，除非是特殊域名。
5. tags 从名称识别地区：HK、JP、US、SG、TW、KR、DE、GB、FR、CA、AU。
6. 去重。
7. 对危险字段加入 warnings，不直接 panic。

## Dedupe 规则

优先：

```text
type + server + port + uuid/password + transport + sni
```

如果重复：

- 保留第一个节点。
- 合并 tags。
- 追加 warning：`duplicate skipped from source xxx`。

## Parser Result

```go
type ParseResult struct {
    Nodes    []model.NodeIR `json:"nodes"`
    Warnings []string       `json:"warnings"`
    Errors   []ParseError   `json:"errors"`
}
```

单个节点解析失败不应导致整个订阅失败。
