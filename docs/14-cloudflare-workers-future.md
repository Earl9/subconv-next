# 14. Cloudflare Workers 后续路线，非 V1

## 状态

Cloudflare Workers 支持放到 V2 或 V3，不影响 OpenWrt V1。

## 原则

OpenWrt V1 的核心 parser 测试用例必须可复用。

Workers 版可以选择：

1. 用 TypeScript 重写 parser/renderer，并复用 JSON golden tests。
2. 后续把 Go/Rust core 编译成 Wasm，但不在 V1 做。

## Workers 版范围

Workers 版只做：

- 订阅 URL 输入。
- 解析现代节点。
- 输出 Mihomo YAML。
- KV 缓存。
- 可选 D1 保存配置。

Workers 版不做：

- 节点测速。
- TCP/UDP 节点连通性测试。
- 本地路由器集成。

## 与 OpenWrt 复用方式

必须复用：

```text
testdata/nodes/*.txt
testdata/subscriptions/*.txt
testdata/golden/*.yaml
```

这样即使用不同语言实现，也能保证输出一致。

## API 对齐

Workers API 尽量对齐 OpenWrt daemon：

```text
POST /api/parse
POST /api/generate
GET /sub/:token
```

## 为什么不放 V1

OpenWrt V1 要解决的是：

- 本地 UCI 配置
- procd 服务
- LuCI 页面
- 本地缓存
- 路由器资源限制

Workers 是完全不同的运行环境，提前混入会拖慢第一版上线。
