# 06. 订阅拉取、缓存与 SSRF 防护

## Fetcher 输入

```go
type Source struct {
    Name      string
    URL       string
    UserAgent string
    Enabled   bool
}
```

## Fetcher 输出

```go
type FetchedSubscription struct {
    Name        string
    URL         string
    Content     []byte
    ContentType string
    FetchedAt   time.Time
    FromCache   bool
}
```

## 限制

默认限制：

| 项目 | 默认 |
|---|---:|
| fetch timeout | 15s |
| max redirects | 3 |
| max body size | 5 MiB |
| allowed scheme | http, https |
| blocked hosts | localhost, private IP, link-local IP, multicast |
| cache TTL | refresh_interval |

## SSRF 防护

必须阻止：

- `127.0.0.0/8`
- `10.0.0.0/8`
- `172.16.0.0/12`
- `192.168.0.0/16`
- `169.254.0.0/16`
- `::1/128`
- `fc00::/7`
- `fe80::/10`
- `localhost`
- `.local`
- OpenWrt 路由器自身 LAN IP，若能检测

流程：

```text
Parse URL
  ↓
Validate scheme
  ↓
Resolve host
  ↓
Reject private/link-local/loopback/multicast
  ↓
HTTP request with timeout
  ↓
LimitReader(max bytes)
  ↓
Store cache
```

## DNS Rebinding 防护

HTTP client 的 DialContext 应复用验证后的 IP。

不要只在请求前解析一次 host，然后让默认 client 再解析一次。

## 缓存

缓存路径：

```text
/var/run/subconv-next/cache/<sha256(url)>.body
/var/run/subconv-next/cache/<sha256(url)>.meta.json
```

meta：

```json
{
  "url_hash": "sha256...",
  "fetched_at": "2026-04-26T00:00:00Z",
  "status_code": 200,
  "etag": "",
  "last_modified": "",
  "size": 12345
}
```

## 错误策略

如果拉取失败：

1. 有未过期缓存：使用缓存并返回 warning。
2. 有过期缓存：使用缓存并返回 warning。
3. 无缓存：该订阅失败，但其他订阅继续。
4. 全部失败：不覆盖上次成功生成的 YAML。

## User-Agent

默认：

```text
SubConvNext/0.1 OpenWrt
```

## 不记录敏感 URL

日志中 URL 只显示：

```text
https://example.com/***
```

不要记录 query token。
