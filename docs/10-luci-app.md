# 10. LuCI App

## 目标

创建 `luci-app-subconv-next`，在 LuCI 中提供：

```text
Services → SubConv Next
```

页面：

1. Overview
2. Subscriptions
3. Render Settings
4. Logs
5. About

## 现代 LuCI 方式

V1 使用 JavaScript view，不使用旧 Lua controller。

目录：

```text
package/openwrt/luci-app-subconv-next/root/
├── usr/share/luci/menu.d/luci-app-subconv-next.json
├── usr/share/rpcd/acl.d/luci-app-subconv-next.json
└── www/luci-static/resources/view/subconv-next/
    ├── overview.js
    ├── subscriptions.js
    ├── render.js
    ├── logs.js
    └── about.js
```

## menu.d

```json
{
  "admin/services/subconv-next": {
    "title": "SubConv Next",
    "order": 60,
    "action": {
      "type": "firstchild"
    },
    "depends": {
      "acl": [ "luci-app-subconv-next" ]
    }
  },
  "admin/services/subconv-next/overview": {
    "title": "Overview",
    "order": 10,
    "action": {
      "type": "view",
      "path": "subconv-next/overview"
    }
  },
  "admin/services/subconv-next/subscriptions": {
    "title": "Subscriptions",
    "order": 20,
    "action": {
      "type": "view",
      "path": "subconv-next/subscriptions"
    }
  },
  "admin/services/subconv-next/render": {
    "title": "Render Settings",
    "order": 30,
    "action": {
      "type": "view",
      "path": "subconv-next/render"
    }
  },
  "admin/services/subconv-next/logs": {
    "title": "Logs",
    "order": 40,
    "action": {
      "type": "view",
      "path": "subconv-next/logs"
    }
  }
}
```

## ACL

```json
{
  "luci-app-subconv-next": {
    "description": "Grant access to SubConv Next configuration and API",
    "read": {
      "uci": [ "subconv_next" ],
      "ubus": {
        "service": [ "list" ],
        "file": [ "read" ]
      }
    },
    "write": {
      "uci": [ "subconv_next" ],
      "ubus": {
        "service": [ "set", "delete", "list" ]
      }
    }
  }
}
```

V1 如果直接调用 daemon HTTP API，应只访问 localhost API，并避免暴露 token。

## Overview 页面

显示：

- 服务状态
- 版本
- 节点数量
- 启用订阅数量
- 上次刷新时间
- 上次错误
- 按钮：启动、停止、重启、刷新生成、下载 YAML

API：

```text
GET /api/status
POST /api/refresh
GET /sub/mihomo.yaml
```

## Subscriptions 页面

使用 `form.Map('subconv_next')` 管理 UCI：

- service main
- subscription sections
- inline sections

字段：

- name
- enabled
- url
- user_agent
- insecure_skip_verify

支持动态添加/删除 subscription。

## Render Settings 页面

管理：

- template
- mixed_port
- allow_lan
- mode
- log_level
- ipv6
- dns_enabled
- enhanced_mode

## Logs 页面

调用：

```text
GET http://127.0.0.1:9876/api/logs?tail=200
```

显示纯文本。

## UX 要求

- 刷新按钮点击后显示 loading。
- 成功后显示节点数量。
- 失败后显示错误 message。
- 不在页面显示完整订阅 URL query token。
- URL 输入框使用 password-like toggle 可后续做，V1 可普通输入。

## Codex 验收

- LuCI 菜单出现。
- 能新增订阅。
- 点击 Save & Apply 后写入 `/etc/config/subconv_next`。
- 能刷新生成。
- 能下载 YAML。
