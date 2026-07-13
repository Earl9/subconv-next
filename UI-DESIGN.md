# SubConv Next LuCI UI 重构设计方案

设计日期：2026-07-11
依据：[ARCHITECTURE.md](ARCHITECTURE.md) 与 `/root/SubConv-Next-UI-DESIGN.md` 需求。
阶段约束：本阶段只输出设计，不修改 LuCI、RPC、UCI、init 或打包代码。

## 1. 设计目标

将当前单页 LuCI 配置表单升级为现代化、轻量、可诊断的服务管理 Dashboard，同时保持：

- LuCI 原生 JavaScript View。
- LuCI RPC 与 UCI API。
- procd/init.d 服务兼容。
- 现有 UCI 字段兼容。
- 无 Node.js、npm、React、Vue 或前端构建步骤。
- PC、平板和手机可用。
- 页面状态全部来自真实系统数据，不硬编码运行状态。

设计参考 OpenClash、Home Assistant 和 Portainer 的信息组织方式，但视觉和交互必须服从 OpenWrt 管理后台的轻量、紧凑和可预测性。

## 2. 设计原则

### 2.1 信息优先级

1. 服务是否健康。
2. 用户是否能执行恢复动作。
3. 当前监听和存储配置是否有效。
4. 日志是否能解释故障。
5. 版本和构建信息。

### 2.2 LuCI 原生优先

- 使用 `view`、`form`、`uci`、`rpc`、`poll`、`ui`、`dom` 等 LuCI 模块。
- 使用 menu JSON 注册页面，不增加传统 Lua Controller。
- 使用 `E()` 创建 DOM，不拼接包含 UCI 值的 raw HTML。
- 使用 LuCI 标准按钮、表单、通知和模态框。
- 自定义 CSS 只作用于 `.subconv-next-app` 命名空间。

### 2.3 操作型界面

- Dashboard 可使用少量状态摘要块。
- Service、Network、Storage 使用全宽、紧凑的表单区，不堆叠装饰性卡片。
- Logs 使用固定高度日志工具面板。
- 不使用大 Hero、渐变背景、装饰图形或嵌套卡片。
- 圆角不超过 8px。

### 2.4 安全默认

- Start、Stop、Restart、清空日志等动作必须通过固定 RPC 方法执行。
- 前端不得获得任意 shell 执行权限。
- 所有路径、端口和 URL 在 RPC 后端再次校验。
- 日志在后端脱敏并限制读取量。
- Stop 和清空日志需要确认；Start、Restart 可直接执行但必须显示结果。

## 3. 信息架构与导航

### 3.1 菜单结构

```text
Services
└── SubConv Next
    ├── Dashboard
    ├── Service
    ├── Network
    ├── Storage
    ├── Logs
    └── About
```

### 3.2 URL 规划

| 页面 | LuCI URL | View |
| --- | --- | --- |
| Dashboard | `admin/services/subconv-next/dashboard` | `subconv-next/dashboard` |
| Service | `admin/services/subconv-next/service` | `subconv-next/service` |
| Network | `admin/services/subconv-next/network` | `subconv-next/network` |
| Storage | `admin/services/subconv-next/storage` | `subconv-next/storage` |
| Logs | `admin/services/subconv-next/logs` | `subconv-next/logs` |
| About | `admin/services/subconv-next/about` | `subconv-next/about` |

保留兼容入口：

```text
admin/services/subconv-next
```

父节点通过 `firstchild` 或等效 LuCI menu action 跳转到 Dashboard，避免旧书签失效。

### 3.3 页面公共头部

每个页面顶部使用相同的紧凑标题区：

```text
SubConv Next                           [Running]
当前页面说明                           v1.0.7-13
```

- 页面名使用 LuCI 页面标题级别，不使用超大字号。
- 右侧状态来自 `luci.subconv.status`。
- 状态点击后进入 Dashboard，不承担操作。
- 移动端状态移到标题下方，避免挤压。

## 4. 全局状态模型

### 4.1 状态定义

| 状态 | 颜色 | 判定 |
| --- | --- | --- |
| Running | 绿色 | procd 运行、端口监听、本地 health 正常 |
| Stopped | 红色 | procd 无运行实例 |
| Error | 黄色 | 进程存在但端口或 health 异常，或检测到退出/respawn 异常 |
| Checking | 中性灰 | RPC 请求尚未完成 |

颜色只作为辅助信息，必须同时显示文字和状态标记。

### 4.2 状态降级

- RPC 不可用：显示 `Error`，说明“状态服务不可用”，保留 UCI 表单读取能力。
- health 超时：显示 `Error`，说明“进程运行但健康检查失败”。
- 端口检测不支持：显示 `Unknown`，不能错误标为 Stopped。
- version 命令失败：显示 `Unknown`，About 页面提供错误详情。

### 4.3 操作期间状态

执行服务动作时：

1. 禁用同组操作按钮。
2. 当前按钮显示 busy 状态。
3. RPC 返回后轮询 `status`，最长等待 10 秒。
4. 状态收敛后显示成功通知。
5. 超时则显示错误通知和最后一次状态，不整页刷新。

## 5. Dashboard 设计

### 5.1 目标

用户打开页面后，在一个视口内判断：服务是否可用、运行在哪、运行多久、应该执行什么动作。

### 5.2 桌面布局

```text
┌──────────────────────────────────────────────────────────────┐
│ SubConv Next                                      [Running]  │
│ OpenWrt service management                                  │
├──────────────────────────────────────────────────────────────┤
│ Status       Version       PID           Uptime              │
│ Running      1.0.7-13      1234          3d 04h 12m          │
├──────────────────────────────────────────────────────────────┤
│ Listen Address                 Port                          │
│ 0.0.0.0                       9876                           │
├──────────────────────────────────────────────────────────────┤
│ [Start] [Stop] [Restart]                     [Open Web UI]   │
├──────────────────────────────────────────────────────────────┤
│ Health details                                               │
│ procd: OK   port: listening   healthz: OK   autostart: on    │
└──────────────────────────────────────────────────────────────┘
```

摘要指标可使用同一行的轻量块，但不在块内嵌套卡片。

### 5.3 移动布局

- Status 和 Version 两列。
- PID 和 Uptime 两列。
- Listen Address、Port 各自一行。
- 操作按钮使用 2x2 网格，稳定按钮高度。
- Open Web UI 占整行。

### 5.4 快捷操作

| 操作 | 可用条件 | 行为 |
| --- | --- | --- |
| Start | Stopped/Error | 调用 `start`，等待 Running |
| Stop | Running/Error 且存在进程 | 确认后调用 `stop` |
| Restart | Running/Error | 调用 `restart`，等待 Running |
| Open Web UI | 配置可组成 URL | 新窗口打开；不可达时仍允许打开但显示提示 |

Open Web UI URL 不直接使用 `public_base_url`。优先使用 RPC 返回的 `web_ui_url`，后端根据 host、port、当前 LuCI 请求主机和 IPv6 规则生成本地管理地址。`public_base_url` 只作为发布链接配置展示。

### 5.5 健康详情

显示四个独立检查结果：

- procd instance。
- 监听端口。
- `/healthz`。
- 自动启动。

Error 状态下展开简短原因，例如：

```text
Process is running, but 127.0.0.1:9876 did not pass health check.
```

## 6. Service 页面设计

### 6.1 页面职责

管理 daemon 生命周期，不编辑订阅、规则或节点业务数据。

### 6.2 状态区

显示：

- Service State。
- PID。
- Started At。
- Uptime。
- Respawn 状态。
- 最近退出码或错误摘要。

### 6.3 开关定义

必须区分两个开关：

| 开关 | 数据来源 | 含义 |
| --- | --- | --- |
| Enable service | UCI `subconv-next.main.enabled` | init 被调用时是否创建服务实例 |
| Start at boot | init enable/disable | 是否存在 rc.d 自动启动链接 |

交互规则：

- 关闭 Enable service 后保存应用，当前运行实例停止。
- 开启 Enable service 后保存应用，若 Start at boot 开启则启动服务。
- Start at boot 通过 RPC 修改，不直接编辑 `/etc/rc.d`。
- 两个开关状态不一致时显示说明，不自动偷偷改另一个状态。

### 6.4 服务操作

页面底部提供 Start、Stop、Restart，与 Dashboard 共用组件和 RPC。

Stop 确认内容必须说明：Web UI 和订阅刷新服务会停止，但已有发布 YAML 文件不会被删除。

## 7. Network 页面设计

### 7.1 UCI 字段

保持现有字段名，避免升级破坏：

| UI 标签 | UCI 字段 | 验证 |
| --- | --- | --- |
| Listen Address | `host` | IPv4、IPv6 或合法主机地址；默认 `0.0.0.0` |
| Port | `port` | 1-65535；默认 9876 |
| Public Base URL | `public_base_url` | 空值或绝对 HTTP(S) URL |
| Log Level | `log_level` | debug/info/warn/error |

需求文档中的 `listen` 在 UI 数据模型中使用 `listen_address`，持久化时仍映射到现有 UCI `host`。

### 7.2 端口检测

Port 输入旁提供 `Check` 命令按钮。

检测请求：

```json
{
  "host": "0.0.0.0",
  "port": 9876
}
```

结果：

- Available：当前无冲突。
- In use by SubConv Next：服务自身正在监听，保存相同端口安全。
- In use：被其他进程占用。
- Invalid：地址或端口格式错误。
- Unknown：系统缺少检测能力。

检测只是预检查。保存应用后仍需以真实启动状态为准。

### 7.3 保存与应用

- Save：提交 UCI，不重启。
- Save & Apply：提交 UCI，触发一次 restart，并等待状态。
- 修改 host/port 时显示“管理 Web UI 地址会变化”的行内提示。
- 修改 Public Base URL 不应影响 Open Web UI 按钮地址。

## 8. Storage 页面设计

### 8.1 UCI 字段

```text
Data Directory -> subconv-next.main.data_dir
```

必须是绝对路径。默认：

```text
/etc/subconv-next/data
```

### 8.2 检测信息

状态区显示：

- Exists。
- Is Directory。
- Readable。
- Writable。
- Permission，例如 `0755`。
- Owner，例如 `root:root`。
- Available Space。
- App log 是否存在及大小。
- 工作区、发布目录是否存在。

### 8.3 检测行为

`Check` 使用当前输入值，不要求先保存。

写权限检测应在目标目录创建随机命名的零字节临时文件并立即删除。路径必须经过后端绝对路径校验，禁止 `..` 路径逃逸和 shell 拼接。

### 8.4 修改目录提示

改变 Data Directory 不自动迁移数据。保存前显示确认：

```text
Changing the data directory does not move existing workspaces,
published files, cache, or logs.
```

本阶段不设计“一键迁移”，避免在路由器上进行高风险递归复制。

## 9. Logs 页面设计

### 9.1 布局

```text
Logs                                   [Auto refresh: on]
[Source: Service ▼] [Lines: 200 ▼] [Refresh] [Clear] [Download]
┌──────────────────────────────────────────────────────────────┐
│ 2026-07-11T... subconv-next[1234]: ...                       │
│ ...                                                          │
└──────────────────────────────────────────────────────────────┘
Showing 200 lines · last update 10:32:11 · not truncated
```

日志区域是固定高度、可滚动的工具面板，不放在装饰卡片中。

### 9.2 日志来源

| Source | 数据 |
| --- | --- |
| Service | `logread` 中与 subconv-next 相关的 stdout/stderr |
| Application | `${data_dir}/logs/app.log` 和轮转文件 |

默认选择 Service。日志后端统一脱敏 token、URL 密钥、Authorization、Cookie、password、uuid 和私钥字段。

### 9.3 实时与自动刷新

- 默认读取 200 行。
- 可选 100、200、500、1000 行。
- 单次 RPC 最大 1000 行、256 KiB，先达到任一限制即截断。
- 自动刷新默认开启，间隔 3 秒。
- 页面不可见时暂停自动刷新。
- 用户滚动到日志底部时自动跟随；向上查看历史时不强制跳到底部。
- 离开页面时移除 poll 任务。

### 9.4 手动刷新

Refresh 使用熟悉的刷新图标或 LuCI 标准命令按钮，执行期间保持日志内容，完成后原位更新，不闪白。

### 9.5 清空日志

- Service/logd：不能按单服务安全清空，因此 Clear 只清空当前页面显示缓冲区，并明确不会清除系统日志。
- Application：确认后截断 `app.log` 并删除受控轮转文件 `app.log.1` 至 `app.log.3`。
- 禁止允许前端传任意文件路径。

### 9.6 下载日志

- 下载当前选择的来源。
- RPC 返回最多 512 KiB 的脱敏文本和 `truncated` 标志。
- 前端使用 Blob 下载 `subconv-next-logs-YYYYMMDD-HHMMSS.txt`。
- 超过限制时文件首行写明已截断。

## 10. About 页面设计

### 10.1 显示内容

| 字段 | 来源 |
| --- | --- |
| Version | `/usr/bin/subconv-next version` |
| Build Time | 编译 ldflags 注入，RPC 返回 |
| Package Version | opkg status 或编译常量 |
| LuCI Package Version | opkg status |
| GitHub | `https://github.com/Earl9/subconv-next` |
| License | MIT |
| OpenWrt | `ubus call system board` 的 release 信息 |
| Architecture | `ubus call system board` 或 `uname -m` |

当前程序没有 build time。实现阶段应增加只读构建变量，例如：

```text
main.buildTime
```

由 Go ldflags 注入，禁止前端写死。

### 10.2 操作

- Open GitHub。
- Open Releases。
- Copy diagnostic summary。

诊断摘要只包含版本、状态、OpenWrt 和配置路径，不包含订阅 URL、token 或业务配置。

## 11. RPC 设计

### 11.1 命名

按设计需求采用 ubus object：

```text
luci.subconv
```

LuCI 声明示例：

```javascript
rpc.declare({
    object: 'luci.subconv',
    method: 'status',
    expect: { '': {} }
});
```

这将 `ARCHITECTURE.md` 中暂定的 `subconv-next.status` 名称收敛为需求指定的 `luci.subconv.status`。ACL 只授权下列固定方法。

### 11.2 status

```text
luci.subconv.status
```

请求：无参数。

响应：

```json
{
  "success": true,
  "state": "running",
  "running": true,
  "enabled": true,
  "autostart": true,
  "pid": 1234,
  "started_at": "2026-07-11T06:00:00Z",
  "uptime_seconds": 3600,
  "version": "1.0.7-13",
  "listen_address": "0.0.0.0",
  "port": 9876,
  "port_listening": true,
  "health_ok": true,
  "web_ui_url": "http://192.168.1.1:9876/",
  "respawn": true,
  "last_exit_code": 0,
  "message": ""
}
```

### 11.3 服务控制

```text
luci.subconv.start
luci.subconv.stop
luci.subconv.restart
```

请求：无参数。

响应：

```json
{
  "success": true,
  "message": "operation success"
}
```

RPC 后端只能操作固定服务名 `/etc/init.d/subconv-next`，不接收 service name 参数。

### 11.4 autostart

```text
luci.subconv.autostart
```

请求：

```json
{
  "enabled": true
}
```

后端只允许布尔值，并执行固定 init enable/disable。

### 11.5 check_port

```text
luci.subconv.check_port
```

请求：

```json
{
  "host": "0.0.0.0",
  "port": 9876
}
```

响应：

```json
{
  "success": true,
  "state": "self",
  "available": false,
  "owned_by_service": true,
  "message": "Port is currently used by SubConv Next"
}
```

### 11.6 check_storage

```text
luci.subconv.check_storage
```

请求：

```json
{
  "path": "/etc/subconv-next/data"
}
```

响应：

```json
{
  "success": true,
  "exists": true,
  "is_directory": true,
  "readable": true,
  "writable": true,
  "mode": "0755",
  "owner": "root",
  "group": "root",
  "available_bytes": 104857600,
  "app_log_exists": true,
  "app_log_bytes": 4096,
  "workspace_exists": true,
  "published_exists": true,
  "message": ""
}
```

### 11.7 logs

```text
luci.subconv.logs
```

请求：

```json
{
  "source": "service",
  "tail": 200,
  "max_bytes": 262144
}
```

响应：

```json
{
  "success": true,
  "source": "service",
  "lines": ["..."],
  "truncated": false,
  "updated_at": "2026-07-11T07:00:00Z",
  "message": ""
}
```

### 11.8 clear_logs

```text
luci.subconv.clear_logs
```

请求仅允许：

```json
{
  "source": "application"
}
```

Service source 不执行系统日志删除，返回可解释错误。

### 11.9 about

```text
luci.subconv.about
```

响应：

```json
{
  "success": true,
  "version": "1.0.7-13",
  "build_time": "2026-07-11T07:00:00Z",
  "package_version": "1.0.7-13",
  "luci_package_version": "1.0.7-13",
  "github_url": "https://github.com/Earl9/subconv-next",
  "license": "MIT",
  "openwrt_release": "25.12.2",
  "architecture": "aarch64_generic"
}
```

### 11.10 错误模型

所有 RPC 业务错误使用一致结构：

```json
{
  "success": false,
  "code": "PORT_IN_USE",
  "message": "Port 9876 is already in use"
}
```

前端展示 `message`，诊断详情可展示 `code`。不得把 shell 命令、敏感路径内容或订阅凭据原样返回。

## 12. UCI 数据模型

### 12.1 保持兼容的字段

```javascript
{
  enabled: true,
  listen_address: '0.0.0.0', // maps to UCI host
  port: 9876,
  data_dir: '/etc/subconv-next/data',
  log_level: 'info',
  public_base_url: ''
}
```

运行状态、日志、PID、uptime、统计、缓存和历史记录禁止写入 UCI。

### 12.2 保存边界

- UCI API 负责读取、修改和提交以上配置。
- RPC 只做检测和生命周期动作。
- rpcd 不直接代替 UCI 保存，避免两条写入路径。
- SubConv Next JSON 业务配置继续由 Web UI 管理。

## 13. CSS 与视觉规范

### 13.1 CSS 作用域

所有规则放在：

```css
.subconv-next-app { ... }
```

不覆盖全局 `body`、`.cbi-map`、`.btn` 或 LuCI 主题变量。

### 13.2 颜色

优先使用 LuCI 主题颜色和 CSS 自定义变量，提供后备值：

| 用途 | 建议后备值 |
| --- | --- |
| Running | `#198754` |
| Stopped | `#c0392b` |
| Error | `#b7791f` |
| Text | `#202428` |
| Muted | `#66727d` |
| Border | `#d7dde3` |
| Surface | `#ffffff` |
| Tool surface | `#111820` |

不使用大面积单一蓝色、紫色渐变或米色主题。

### 13.3 尺寸

- 内容最大宽度跟随 LuCI 主容器，不另造全屏应用壳。
- 区块间距：16px。
- 表单行最小高度：44px。
- 命令按钮最小高度：36px。
- 圆角：状态块 6px，工具面板 6px。
- 日志面板：桌面 `min-height: 420px`，移动 `min-height: 320px`。
- 字体不随 viewport 宽度缩放。
- Letter spacing 保持 0。

### 13.4 响应式断点

| 宽度 | 布局 |
| --- | --- |
| `>= 1024px` | Dashboard 4 列摘要，操作横排 |
| `640-1023px` | Dashboard 2 列摘要 |
| `< 640px` | 单列配置，摘要 2 列，操作 2x2 |

长路径、URL 和日志必须允许换行或水平滚动，不能溢出父容器。

### 13.5 状态与反馈

- Loading：局部 skeleton 或 LuCI spinning，不清空已有内容。
- Empty：简短说明，不显示功能教程。
- Error：行内错误 + LuCI notification。
- Success：LuCI notification，2-4 秒后自动消失。
- Disabled：保持文字可读并说明依赖条件。

## 14. 可访问性

- 所有状态同时有文字，不只依赖颜色。
- 按钮使用明确文本命令。
- 表单控件有 `<label>`。
- 日志区使用 `aria-live="polite"`，但自动刷新不逐行朗读。
- 模态框焦点可返回触发按钮。
- 键盘可操作所有按钮、开关和选择器。
- 外部链接带 `target="_blank"` 和 `rel="noopener noreferrer"`。

## 15. 文件规划

采用现代 LuCI 结构，不创建 `luasrc/controller`、`luasrc/model` 或 Lua View：

```text
openwrt/luci-app-subconv-next/
├── Makefile
├── htdocs/luci-static/resources/
│   ├── subconv-next/
│   │   ├── common.js
│   │   ├── rpc.js
│   │   └── dashboard.css
│   └── view/subconv-next/
│       ├── dashboard.js
│       ├── service.js
│       ├── network.js
│       ├── storage.js
│       ├── logs.js
│       └── about.js
└── root/
    ├── usr/libexec/rpcd/luci.subconv
    ├── usr/share/luci/menu.d/luci-app-subconv-next.json
    └── usr/share/rpcd/acl.d/luci-app-subconv-next.json
```

共享模块职责：

- `rpc.js`：全部 RPC declaration。
- `common.js`：状态格式化、duration、byte size、通知和服务动作等待。
- `dashboard.css`：全局 scoped LuCI 应用样式。

## 16. 开发顺序

1. RPC 层与最小 ACL。
2. 菜单和共享模块。
3. Dashboard。
4. Service。
5. Network。
6. Storage。
7. Logs。
8. About。
9. 包边界、升级迁移和 IPK 验证。

每个模块完成后单独验证，不等待全部页面结束后一次测试。

## 17. 修改前文件、原因与风险

本设计阶段只新增 `UI-DESIGN.md`，不修改以下代码。进入实现阶段前应再次输出实际修改清单。

| 计划文件 | 原因 | 风险 |
| --- | --- | --- |
| menu JSON | 六页导航和兼容入口 | 菜单缓存、旧 URL |
| ACL JSON | 授权固定 RPC | 权限过宽或漏授权 |
| rpcd helper | 真实系统数据和动作 | 输入校验、命令注入、性能 |
| 六个 JS View | 页面实现 | LuCI API 版本差异、poll 泄漏 |
| common.js/rpc.js | 避免重复 | 模块加载路径兼容 |
| scoped CSS | Dashboard 和响应式布局 | LuCI 主题兼容 |
| UCI config/init | 自动启动和状态一致性 | 旧配置与启停回归 |
| Makefile/打包脚本 | 安装新文件 | 文件所有权、升级路径 |
| Go build variables | Build Time | ldflags 与 release 一致性 |

## 18. 验收标准

### 18.1 页面

- 六个页面均可从 LuCI 菜单进入。
- 旧入口跳转 Dashboard。
- PC、平板、手机无文字溢出和控件重叠。
- 页面不依赖外部 CDN、npm 或运行时网络资源。

### 18.2 RPC

- `luci.subconv.status` 返回真实状态、PID、uptime 和 version。
- start/stop/restart 能正确控制固定服务。
- 非授权用户无法调用写方法。
- 参数非法时返回结构化错误。
- 日志和路径结果不泄露敏感数据。

### 18.3 UCI

- 配置通过 LuCI UCI API 保存和 commit。
- 现有 `host`、`port`、`data_dir`、`public_base_url`、`log_level`、`enabled` 保持兼容。
- Save & Apply 只触发一次服务重载或重启。

### 18.4 服务

- `/etc/init.d/subconv-next status` 与 Dashboard 状态一致。
- Start、Stop、Restart 后页面无需整页刷新即可更新。
- 自动启动状态与 init enable/disable 一致。
- Error 状态能解释端口或 health 异常。

### 18.5 日志

- 自动刷新和手动刷新正常。
- 单次读取不超过 1000 行和 256 KiB。
- 清空 Application 日志不影响系统其他服务。
- 下载内容经过脱敏并有大小上限。

### 18.6 构建和安装

- OpenWrt package 编译通过。
- 生成可安装 IPK。
- 升级保留现有 UCI 配置。
- 安装后 rpcd、LuCI menu/module 缓存正确刷新。
- 浏览器访问 `admin/services/subconv-next` 正常进入 Dashboard。

## 19. 第二阶段结论

本方案将 UI 重构限定在 LuCI 原生技术栈内，并把复杂系统操作收敛到专用、最小权限的 `luci.subconv` RPC。六个页面各自承担单一职责，UCI 只保存启动配置，运行状态和日志不落入 UCI。

下一阶段应先实现 RPC 基础设施和 ACL，再实现 Dashboard。未完成 RPC 前，不应在前端伪造 PID、uptime、端口或目录状态。
