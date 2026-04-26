# 05. UCI 配置 Schema

## 文件路径

```text
/etc/config/subconv_next
```

## section: `service`

示例：

```uci
config service 'main'
	option enabled '1'
	option listen_addr '127.0.0.1'
	option listen_port '9876'
	option log_level 'info'
	option template 'standard'
	option output_path '/var/run/subconv-next/mihomo.yaml'
	option cache_dir '/var/run/subconv-next/cache'
	option state_path '/etc/subconv-next/state.json'
	option refresh_interval '3600'
	option max_subscription_bytes '5242880'
	option fetch_timeout_seconds '15'
	option allow_lan '0'
```

字段说明：

| 字段 | 类型 | 默认 | 说明 |
|---|---|---:|---|
| enabled | bool | 1 | 是否启用服务 |
| listen_addr | string | 127.0.0.1 | 监听地址 |
| listen_port | int | 9876 | 监听端口 |
| log_level | enum | info | debug/info/warn/error |
| template | enum | standard | lite/standard/full |
| output_path | path | /var/run/subconv-next/mihomo.yaml | 输出文件 |
| cache_dir | path | /var/run/subconv-next/cache | 缓存目录 |
| state_path | path | /etc/subconv-next/state.json | 状态文件 |
| refresh_interval | int | 3600 | 自动刷新秒数 |
| max_subscription_bytes | int | 5242880 | 单订阅最大字节 |
| fetch_timeout_seconds | int | 15 | fetch 超时 |
| allow_lan | bool | 0 | 是否允许局域网访问 |

## section: `subscription`

可多个。

```uci
config subscription
	option name 'airport-a'
	option enabled '1'
	option url 'https://example.com/sub'
	option user_agent 'SubConvNext/0.1 OpenWrt'
	option insecure_skip_verify '0'
```

字段说明：

| 字段 | 类型 | 默认 | 说明 |
|---|---|---:|---|
| name | string | required | 订阅名称 |
| enabled | bool | 1 | 是否启用 |
| url | string | required | HTTP/HTTPS URL |
| user_agent | string | SubConvNext/0.1 OpenWrt | 拉取 UA |
| insecure_skip_verify | bool | 0 | 是否跳过 TLS 验证，默认禁止 |

## section: `inline`

用于粘贴节点或订阅内容。

```uci
config inline
	option name 'manual'
	option enabled '0'
	option content ''
```

## section: `render`

```uci
config render 'mihomo'
	option mixed_port '7890'
	option allow_lan '0'
	option mode 'rule'
	option log_level 'info'
	option ipv6 '0'
	option dns_enabled '1'
	option enhanced_mode 'fake-ip'
```

## UCI 解析实现建议

Go 里不要完整实现 UCI 全语法。V1 支持以下 subset 即可：

```text
config <type> ['name']
    option <key> '<value>'
    list <key> '<value>'
```

必须支持：

- 单引号
- 双引号
- 无引号简单值
- 注释行 `#`
- 空行

不需要支持复杂 include。

## 本地测试配置

为非 OpenWrt 环境提供 JSON 配置：

```json
{
  "service": {
    "listen_addr": "127.0.0.1",
    "listen_port": 9876,
    "template": "standard"
  },
  "subscriptions": [],
  "inline": []
}
```

Codex 应让 `--config` 根据扩展名自动判断 UCI 或 JSON。
