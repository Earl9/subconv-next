# 04. Daemon 与本地 API

## 监听策略

默认：

```text
127.0.0.1:9876
```

当 UCI `allow_lan=1` 时才允许监听：

```text
0.0.0.0:9876
```

严禁默认暴露到 WAN。

## CLI

必须实现：

```sh
subconv-next serve --config /etc/config/subconv_next
subconv-next generate --config /etc/config/subconv_next --out /tmp/mihomo.yaml
subconv-next parse --input ./input.txt --json
subconv-next version
```

## API 列表

### `GET /healthz`

返回：

```json
{
  "ok": true,
  "version": "0.1.0",
  "uptime_seconds": 12
}
```

### `GET /api/status`

返回：

```json
{
  "running": true,
  "last_refresh_at": "2026-04-26T00:00:00Z",
  "last_success_at": "2026-04-26T00:00:00Z",
  "node_count": 42,
  "enabled_subscription_count": 2,
  "output_path": "/var/run/subconv-next/mihomo.yaml",
  "last_error": ""
}
```

### `POST /api/refresh`

触发立即刷新。

返回：

```json
{
  "ok": true,
  "node_count": 42,
  "output_path": "/var/run/subconv-next/mihomo.yaml"
}
```

### `POST /api/parse`

请求：

```json
{
  "content": "vless://...",
  "content_type": "auto"
}
```

返回：

```json
{
  "ok": true,
  "nodes": []
}
```

### `POST /api/generate`

请求：

```json
{
  "nodes": [],
  "template": "standard"
}
```

返回：

```json
{
  "ok": true,
  "yaml": "mixed-port: 7890\n..."
}
```

### `GET /sub/mihomo.yaml`

返回当前生成的 Mihomo YAML。

如果未生成，尝试即时生成一次。

### `GET /api/logs?tail=200`

返回最近日志。V1 可从内存 ring buffer 返回，不要求读取系统 logread。

## API 错误格式

统一：

```json
{
  "ok": false,
  "error": {
    "code": "FETCH_TIMEOUT",
    "message": "subscription fetch timed out"
  }
}
```

## 内部流程

刷新流程：

```text
Load UCI config
  ↓
Fetch enabled subscriptions
  ↓
Detect input type
  ↓
Parse to NodeIR[]
  ↓
Normalize and dedupe
  ↓
Render Mihomo YAML
  ↓
Atomic write output file
  ↓
Update runtime status
```

## Atomic Write

写文件必须使用：

1. 写到临时文件。
2. fsync。
3. rename 覆盖目标文件。

Go 函数建议：

```go
func AtomicWriteFile(path string, data []byte, perm fs.FileMode) error
```
