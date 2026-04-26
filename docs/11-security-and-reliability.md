# 11. 安全与可靠性

## 安全边界

`subconv-next` 只做订阅解析和配置生成。

禁止实现：

- 代理服务
- 节点转发
- 端口转发
- 透明代理
- 防火墙规则修改
- DNS 劫持
- 绕过检测
- 扫描内网

## 网络安全

订阅拉取必须：

- 只允许 http/https。
- 禁止 localhost/private/link-local/multicast。
- 限制 body 大小。
- 限制 redirect。
- 限制超时。
- 禁止在日志输出完整 URL。
- 禁止跟随跳转到内网地址。

## 本地服务暴露

默认：

```text
listen_addr=127.0.0.1
allow_lan=0
```

只有用户显式设置 `allow_lan=1` 才允许局域网访问。

即使 allow_lan=1，也不要监听 WAN 接口检测逻辑；只提供警告。

## 日志脱敏

需要脱敏：

- 订阅 URL query
- token
- password
- uuid 可保留前后 4 位
- private-key
- public-key 可保留，但不建议在日志输出

示例：

```text
https://example.com/api/v1/client/subscribe?token=***
```

## 文件写入

- 运行缓存写 `/var/run/subconv-next`。
- 不要频繁写 `/etc`。
- 只有状态变更时才写 `/etc/subconv-next/state.json`。
- 输出 YAML 默认写 `/var/run`，避免频繁写 flash。

## 崩溃恢复

- daemon 启动时如果输出文件不存在，自动生成一次。
- 刷新失败时不覆盖旧 YAML。
- parser 局部失败不影响其他订阅。
- 单个订阅失败不影响其他订阅。

## 资源限制

默认：

```text
max_subscription_bytes = 5 MiB
max_nodes = 5000
max_inline_bytes = 1 MiB
fetch_timeout = 15s
render_timeout = 10s
```

如果超过限制，返回明确错误。

## 并发控制

刷新操作必须有锁。

如果已有刷新运行，第二个 `/api/refresh` 返回：

```json
{
  "ok": false,
  "error": {
    "code": "REFRESH_IN_PROGRESS",
    "message": "refresh is already running"
  }
}
```

## 配置校验

daemon 启动时校验：

- 端口范围 1–65535。
- 模板必须是 lite/standard/full。
- output_path 必须是绝对路径。
- cache_dir 必须是绝对路径。
- URL scheme 必须合法。
