# SubConv Next OpenWrt / LuCI Architecture Analysis

分析日期：2026-07-11
分析范围：当前工作树中的 OpenWrt 服务包、LuCI 页面、UCI、procd/init、RPC ACL、IPK 脚本及发布流程。
阶段约束：本阶段只分析现状，不修改运行代码。

## 0. 结论摘要

当前 OpenWrt 集成是一个可工作的轻量包装层：UCI 保存基础启动参数，procd 管理 Go 服务，单个 LuCI JavaScript View 读取 UCI、查询 `service.list` 并调用 `luci.setInitAction`。

它尚不是完整的服务管理 Dashboard，主要限制如下：

- 所有功能集中在一个 `index.js`，没有 Dashboard、Service、Network、Storage、Logs、About 子页面。
- 没有专用 LuCI RPC 后端，只能获得“是否运行”的布尔状态，无法可靠获得 PID、启动时间、错误状态、端口占用、目录权限和服务日志。
- LuCI 管理 UCI 启动参数，SubConv Next Web UI 管理 `/etc/subconv-next/config.json`，实际存在两套配置边界。
- “启用服务”UCI 值与 `/etc/rc.d` 自动启动状态是两个状态，当前页面没有统一表达。
- 当前操作按钮通过 DOM 节点的 `outerHTML` 返回，事件处理器不会随 HTML 序列化，按钮存在失效风险。
- 状态区使用 `rawhtml` 拼接 UCI 值，缺少 HTML 转义，存在持久化脚本注入风险。
- 主包内嵌 LuCI 文件，同时仓库又维护独立 `luci-app-subconv-next` 包，包职责重复。
- 当前升级脚本绕开 OpenWrt conffile 机制，采用 preinst/postinst 临时备份恢复 UCI，解决了近期安装问题，但不是长期标准方案。

推荐目标是保留 LuCI 原生 JavaScript View、UCI 和 procd，同时增加最小权限的专用 rpcd 对象 `subconv-next`，将页面拆成六个 View，并明确区分“服务启动配置”和“应用业务配置”。

## 1. 当前架构说明

### 1.1 目录与职责

| 路径 | 当前职责 |
| --- | --- |
| `openwrt/luci-app-subconv-next/Makefile` | 标准 LuCI 包定义，架构为 `all`，依赖 `luci-base` 和 `subconv-next` |
| `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/index.js` | 唯一 LuCI JavaScript View，包含状态、配置表单和服务操作 |
| `openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json` | 注册 `admin/services/subconv-next` 菜单入口 |
| `openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json` | 授权读取/写入 UCI，并调用 `service.list`、`luci.setInitAction` |
| `openwrt/subconv-next/Makefile` | OpenWrt Buildroot 核心包；当前同时安装 daemon、init、UCI 和全部 LuCI 文件 |
| `openwrt/subconv-next/files/subconv-next.config` | 最小 UCI 默认配置 |
| `openwrt/subconv-next/files/subconv-next.init` | rc.common/procd 服务脚本 |
| `scripts/package-openwrt-ipk-sdk.sh` | 使用 SDK `ipkg-build` 手工组装 all-in-one IPK，并验证内容和权限 |
| `scripts/package-openwrt-luci-ipk-sdk.sh` | 手工组装独立 LuCI IPK |
| `scripts/package-openwrt-ipk-portable.sh` | 创建临时兼容 `ipkg-build` 后调用主打包脚本 |
| `.github/workflows/auto-release.yml` | 构建多架构二进制；当前只发布 arm64 all-in-one IPK |

### 1.2 LuCI Controller 状态

当前没有 Lua Controller，也没有 ucode Controller。

这是现代 LuCI JavaScript 应用的合法结构：菜单 JSON 直接将 URL 映射到 JavaScript View，因此不需要传统的 `luasrc/controller/*.lua`。后续重构应继续使用菜单 JSON，不建议为了形式补回 Lua Controller。

### 1.3 LuCI View 与 CSS

当前只有一个 View：

```text
view/subconv-next/index.js
```

它使用：

- `form`：生成 UCI 表单。
- `poll`：轮询服务运行状态。
- `rpc`：调用 ubus RPC。
- `uci`：读取和保存 `subconv-next` 配置。
- `ui`：模态框和通知。
- `view`：LuCI 页面基类。

当前没有独立 CSS 文件。页面完全依赖 LuCI 默认 CBI/按钮样式，因此缺少 Dashboard 布局、状态卡片、日志查看器和响应式导航所需的局部样式。

### 1.4 RPC 与权限

前端声明两个 RPC：

```text
service.list(name)
luci.setInitAction(name, action)
```

ACL 允许：

- 读取和写入 UCI 包 `subconv-next`。
- 调用 `service.list`。
- 调用 `luci.setInitAction`。

当前没有 `subconv-next` 专用 ubus/rpcd 对象，也没有文件、日志、端口或版本检测 RPC。

### 1.5 UCI 配置

当前 UCI 只有一个命名 section：

```uci
config subconv-next 'main'
    option enabled '1'
    option host '0.0.0.0'
    option port '9876'
    option data_dir '/etc/subconv-next/data'
    option public_base_url ''
    option log_level 'info'
```

这些字段只负责 OpenWrt 服务包装层。订阅源、规则、节点和渲染配置仍由 SubConv Next Web UI 的 JSON 工作区管理。

### 1.6 init.d / procd

`/etc/init.d/subconv-next` 使用 `USE_PROCD=1`，主要行为是：

1. 加载 UCI `subconv-next.main`。
2. `enabled != 1` 时直接返回，不启动实例。
3. 创建 `data_dir`。
4. 执行：

   ```sh
   /usr/bin/subconv-next serve --config /etc/subconv-next/config.json
   ```

5. 通过环境变量覆盖监听地址、端口、数据目录、日志级别和 Public Base URL。
6. 监听 UCI 文件变化并通过 reload trigger 重启。
7. 启用 procd respawn，并将 stdout/stderr 发送到 logd。

程序在 JSON 配置不存在时使用默认配置，因此首次安装可以仅依赖 UCI 环境变量启动。Web UI 后续会写入 `/etc/subconv-next/config.json`。

## 2. 页面入口

### 2.1 当前入口

菜单节点：

```text
admin/services/subconv-next
```

浏览器地址：

```text
http://router-ip/cgi-bin/luci/admin/services/subconv-next
```

View 映射：

```text
subconv-next/index
```

最终静态文件：

```text
/www/luci-static/resources/view/subconv-next/index.js
```

### 2.2 当前页面内容

当前页面是单个 `form.Map`，包含：

- Running/Stopped 状态。
- Port、Data Directory、Web UI URL。
- enabled、host、port、data_dir、public_base_url、log_level。
- Start、Stop、Restart、Open Web UI。
- Save & Apply 后自动 restart。

### 2.3 目标菜单结构建议

```text
SubConv Next
├── Dashboard
├── Service
├── Network
├── Storage
├── Logs
└── About
```

推荐为每个页面建立独立 menu node 和独立 JS View，父节点默认跳转 Dashboard。不要在单个 View 内自行实现第二套餐导航状态。

## 3. 数据流

### 3.1 页面加载

```text
Browser
  │
  ├── uci.load('subconv-next')
  │     └── rpcd UCI API -> /etc/config/subconv-next
  │
  └── service.list('subconv-next')
        └── ubus service -> procd instance state
```

`service.list` 结果被压缩为单个布尔值，只要任一 instance 的 `running` 为 true 就显示 Running。

### 3.2 配置保存

```text
LuCI form
  -> UCI staging
  -> uci save/commit/apply
  -> /etc/config/subconv-next
  -> procd UCI reload trigger
  -> init reload_service
  -> restart
```

页面另外覆盖 `handleSaveApply()`，保存完成后再次调用 `restart`。因此存在 reload trigger 与显式 restart 重复触发的可能。

### 3.3 服务控制

```text
Start / Stop / Restart
  -> luci.setInitAction('subconv-next', action)
  -> /etc/init.d/subconv-next <action>
  -> procd
```

当前调用完成后整页刷新，没有检查动作后的真实状态，也没有等待状态收敛。

### 3.4 Web UI

```text
LuCI Open Web UI
  -> 根据 UCI host/port/public_base_url 拼 URL
  -> 浏览器直接访问 SubConv Next HTTP 服务
```

SubConv Next Web UI 不是 LuCI 内嵌页面，不经过 LuCI 登录会话。项目 README 也明确说明 V1 没有内置登录，因此监听 `0.0.0.0` 时应视为局域网暴露服务。

### 3.5 日志

当前存在两种日志来源：

1. procd 将 stdout/stderr 写入 OpenWrt logd，可通过 `logread` 查看。
2. SubConv Next 将应用事件写入 `${data_dir}/logs/app.log`，并维护轮转文件。

现有 LuCI 页面没有读取任一日志来源。应用的 `/api/logs` 是工作区日志接口，需要 workspace 参数，不适合作为 OpenWrt 服务级日志的唯一来源。

## 4. 配置读取流程

### 4.1 OpenWrt 启动配置

UCI 字段由 init 脚本读取并转成环境变量：

| UCI | 环境变量 | 程序配置 |
| --- | --- | --- |
| `host` | `SUBCONV_HOST` | `Service.ListenAddr` |
| `port` | `SUBCONV_PORT` | `Service.ListenPort` |
| `data_dir` | `SUBCONV_DATA_DIR` | state/cache/output 路径根目录 |
| `public_base_url` | `SUBCONV_PUBLIC_BASE_URL` | `Service.PublicBaseURL` |
| `log_level` | `SUBCONV_LOG_LEVEL` | service/render log level |

优先级是：显式 CLI flag > 环境变量 > JSON/default config。

### 4.2 应用业务配置

`/etc/subconv-next/config.json` 由 SubConv Next Web UI/API 使用，包含订阅、规则、工作区和渲染配置。它不是当前 UCI 文件的自动镜像。

因此配置边界应定义为：

- UCI：daemon 生命周期、监听网络、数据目录、日志级别。
- JSON/workspace：订阅转换业务配置。

后续 LuCI 页面不应尝试把完整业务配置复制进 UCI，也不应直接修改 JSON 文件。

### 4.3 自动启动的双状态问题

当前有两个相关但不同的状态：

- UCI `enabled=1`：init 的 `start_service()` 是否创建 procd instance。
- `/etc/rc.d/S95subconv-next`：服务是否在系统启动时被调用。

包安装后无条件执行 init `enable`，而 LuCI 的“Enable service”只修改 UCI。目标 Service 页面应明确合并或同时展示这两个状态，避免“UCI 开启但 rc.d 未启用”或相反。

## 5. 服务控制流程

### 5.1 当前状态读取

```text
service.list -> instances -> running -> Running / Stopped
```

当前不展示：

- PID。
- 实际命令行。
- 启动时间和 uptime。
- 最近退出码。
- respawn/crash 状态。
- 监听端口是否真正建立。
- `/healthz` 是否响应。

### 5.2 当前动作执行

`luci.setInitAction` 可执行 start、stop、restart，也可用于 enable、disable。现有 ACL 在 read 和 write 区域都声明了 `setInitAction`，权限范围应收紧到 write。

### 5.3 推荐状态模型

专用 RPC `subconv-next.status` 应返回真实、结构化数据：

```json
{
  "state": "running|stopped|error",
  "running": true,
  "enabled": true,
  "autostart": true,
  "pid": 1234,
  "started_at": "2026-07-11T00:00:00Z",
  "uptime_seconds": 3600,
  "listen_address": "0.0.0.0",
  "port": 9876,
  "port_listening": true,
  "health_ok": true,
  "version": "1.0.7-13",
  "last_error": ""
}
```

数据必须来自 procd、`/proc`、init.d、真实端口和本地 health endpoint，不在前端硬编码。

## 6. 当前问题分析

### 6.1 高优先级

#### A. 操作按钮事件可能丢失

当前 `actions.cfgvalue()` 创建带 `click` 处理器的 DOM 节点，然后返回 `.outerHTML`。事件处理器不会被序列化进 HTML，Start/Stop/Restart 按钮可能无法执行。

建议直接返回 DOM 节点，或在 `render()` 完成后绑定事件，不使用 `outerHTML`。

#### B. raw HTML 注入风险

`renderStatus()` 将 UCI 的 port、data_dir 和 URL 直接拼入字符串，并设置 `rawhtml=true`。管理员可控 UCI 值未经转义，可能形成持久化 HTML/脚本注入。

建议全部使用 `E()` 创建文本节点和属性，并使用 `URL` API 生成链接。

#### C. 服务状态过于简化

只区分 Running/Stopped，无法满足 Error 黄色状态，也无法识别“进程存在但端口未监听”或“端口监听但 health 失败”。

#### D. 配置与自动启动语义不一致

UCI enabled 和 init enable/disable 没有统一，页面上的“Enable service”并不等价于“自动启动”。

#### E. 包职责重复

核心包安装 LuCI 文件且依赖 `luci-base`，独立 `luci-app-subconv-next` 又安装同一文件。两包同时安装会产生文件所有权冲突和升级不确定性。

### 6.2 中优先级

#### F. Web UI URL 生成不可靠

- 跟随 LuCI 的 HTTPS scheme，但 SubConv Next 通常只提供 HTTP。
- 特定 IPv6 地址没有加方括号。
- `public_base_url` 是发布链接基址，不一定是管理 Web UI 地址。
- 监听 `0.0.0.0`/`::` 时替换浏览器主机是合理的，但未验证端口可达。

#### G. Save & Apply 可能重复重启

UCI reload trigger 已会 restart，页面又在 `handleSaveApply()` 中显式 restart。

#### H. 缺少专用 RPC

端口检测、目录检测、版本、PID、日志读取、日志清理和下载无法通过当前 RPC 安全实现。直接授予任意 shell/文件权限不符合最小权限原则。

#### I. 日志“清空”定义不明确

logd 是全局环形日志，不能安全地只清空一个服务。建议：

- 实时查看：读取 `logread -e subconv-next`。
- 清空：只截断应用自己的 `${data_dir}/logs/app.log*`，或仅清空前端显示缓冲区。
- 下载：由专用 RPC 返回脱敏日志文本，前端生成下载文件。

#### J. Storage 修改不迁移数据

修改 `data_dir` 后 init 只创建新目录，不会迁移原工作区、发布文件和日志。页面必须显示明确警告，并在未来决定是否提供显式迁移操作。

### 6.3 打包与维护问题

#### K. 非标准 conffile 升级策略

当前包故意不生成 `CONTROL/conffiles`，改用固定 `/tmp/subconv-next.config.keep` 在 preinst/postinst 备份恢复。

优点：避免近期 `/etc/config/subconv-next` conffile 状态不一致导致的 opkg 安装错误。
风险：中断安装可能遗留备份；固定临时路径缺少事务标识；默认配置升级无法通过标准 `-opkg` 文件提示用户；行为偏离 OpenWrt 包惯例。

长期建议恢复标准 conffile，并使用 `/etc/uci-defaults/` 做字段级迁移，不删除用户配置。迁移前应专门测试从目前无 conffile 元数据版本升级的路径。

#### L. 三套构建路径可能漂移

- OpenWrt Buildroot Makefile。
- SDK 手工组包脚本。
- portable 手工 `ipkg-build`。

版本号、依赖、维护脚本和文件列表在多个位置重复维护。当前 release 使用 portable 路径，而不是标准 `make package/.../compile`。

#### M. 版本与仓库信息不统一

- Makefile、脚本中手工维护 `1.0.7-13`。
- release workflow 动态生成 `1.0.x`，IPK release 固定为 `-1`。
- 核心 Makefile URL 指向 `github.com/subconv-next/subconv-next`，README/API 指向 `github.com/Earl9/subconv-next`。
- 程序只嵌入 version，没有 build time，About 页面暂时无法真实展示构建时间。

#### N. release 未发布独立 LuCI 包

独立 LuCI 打包脚本存在，但 release 只发布 all-in-one `subconv-next_*.ipk`。目标包结构需要先确定，否则后续 Dashboard 文件可能继续被重复打包。

## 7. 推荐重构方案

### 7.1 目标分层

```text
LuCI menu + JavaScript Views + scoped CSS
                    │
                    ├── LuCI UCI API
                    │     └── /etc/config/subconv-next
                    │
                    └── LuCI RPC
                          └── ubus object: subconv-next
                                ├── status
                                ├── service_action
                                ├── autostart
                                ├── check_port
                                ├── check_storage
                                ├── logs
                                ├── clear_logs
                                └── about
                                      │
                                      ├── procd / init.d
                                      ├── /proc
                                      ├── logread / app.log
                                      ├── filesystem checks
                                      └── subconv-next version
```

### 7.2 推荐文件结构

```text
openwrt/luci-app-subconv-next/
├── Makefile
├── htdocs/luci-static/resources/
│   ├── view/subconv-next/
│   │   ├── dashboard.js
│   │   ├── service.js
│   │   ├── network.js
│   │   ├── storage.js
│   │   ├── logs.js
│   │   └── about.js
│   └── subconv-next/dashboard.css
└── root/
    ├── usr/libexec/rpcd/subconv-next
    ├── usr/share/luci/menu.d/luci-app-subconv-next.json
    └── usr/share/rpcd/acl.d/luci-app-subconv-next.json
```

rpcd 实现可以使用 OpenWrt 自带的 shell/ucode 能力，但必须暴露固定方法，禁止给前端任意命令执行权限。

### 7.3 页面职责

| 页面 | 数据来源 | 写操作 |
| --- | --- | --- |
| Dashboard | `subconv-next.status` | start/stop/restart，Open Web UI |
| Service | UCI + status RPC | enabled、autostart、服务动作 |
| Network | UCI + `check_port` | host、port、public_base_url |
| Storage | UCI + `check_storage` | data_dir；不隐式迁移数据 |
| Logs | `logs` RPC | 自动刷新、清空 app log、下载 |
| About | `about` RPC + package metadata | 无 |

### 7.4 状态判定

- Running：procd instance 运行、端口监听且 health 正常。
- Stopped：无运行中的 procd instance。
- Error：进程存在但端口/health 异常，或 procd 最近退出/反复 respawn。

前端只负责颜色和展示，不负责猜测状态。

### 7.5 UCI 保存策略

- 所有配置修改继续使用 LuCI `form`/`uci` API。
- 禁止直接编辑 `/etc/config/subconv-next`。
- Save：仅保存/提交。
- Save & Apply：提交后只触发一次 reload/restart，并等待 RPC 状态收敛。
- autostart 通过受限 RPC 调用 init enable/disable，并与 UCI enabled 的语义在 UI 中明确区分。

### 7.6 包结构建议

推荐恢复标准拆包：

- `subconv-next`：二进制、init、UCI、默认/迁移脚本；不依赖 LuCI。
- `luci-app-subconv-next`：View、CSS、menu、ACL、rpcd helper；依赖 `subconv-next` 和 `luci-base`。

Release 同时发布两个 IPK。若必须保留单文件安装体验，可额外提供明确命名的 bundle/meta 包，而不是让核心包和 LuCI 包重复拥有相同文件。

### 7.7 分阶段实施顺序

1. 基础设施：专用 rpcd、ACL、菜单拆分、共享 CSS/工具模块。
2. Dashboard：真实状态和快捷动作。
3. Service：enabled、autostart、PID、启动时间。
4. Network：host、port、public URL、端口检测。
5. Storage：存在性、目录类型、读写权限、空间信息。
6. Logs：实时刷新、脱敏、清空 app log、下载。
7. About：version、build time、GitHub、License。
8. 打包统一：标准 Buildroot/SDK 流程、升级迁移、双 IPK 发布。

## 8. 修改前文件清单、原因与风险

第二阶段设计完成后，预计代码阶段会涉及以下文件。实际修改前应再次确认清单。

| 文件/目录 | 修改原因 | 主要风险 |
| --- | --- | --- |
| `menu.d/luci-app-subconv-next.json` | 拆分六个页面入口 | 菜单路径兼容和默认跳转 |
| `view/subconv-next/*.js` | 实现 Dashboard 与各管理模块 | LuCI 版本 API 差异、轮询清理 |
| `subconv-next/dashboard.css` | 局部现代化布局 | 与不同 LuCI 主题冲突 |
| `rpcd/acl.d/luci-app-subconv-next.json` | 新 RPC 最小权限 | 权限过宽或漏授权 |
| `usr/libexec/rpcd/subconv-next` | 真实状态、日志、端口、存储检测 | 命令注入、路径校验、性能 |
| `subconv-next.init` | 提供更完整 procd 元数据或一致重载行为 | 启停回归、respawn 行为 |
| `subconv-next.config` | 必要的新 UCI 字段 | 旧配置升级兼容 |
| 两个 OpenWrt Makefile | 明确包边界和安装文件 | 文件所有权、依赖和升级路径 |
| IPK 打包脚本和 release workflow | 统一产物与版本 | opkg 升级、架构命名、发布失败 |
| `cmd/subconv-next/main.go` | 若需真实 build time，则增加 ldflags 字段 | 二进制版本兼容 |

## 9. 第一阶段验证结论

本阶段完成了以下只读验证：

- 核对全部 OpenWrt/LuCI 文件清单。
- 核对菜单、ACL、JS View、UCI 和 init 脚本。
- 核对程序配置缺失回退及环境变量覆盖流程。
- 核对 Buildroot、SDK、portable 和 GitHub Release 打包路径。
- 解包检查现有 `subconv-next_1.0.7-13_aarch64_generic.ipk` 的 control/data 内容。
- 确认现有 IPK 包含 daemon、init、UCI、LuCI menu/ACL/View，且 control 中没有 conffiles。

本阶段没有修改运行代码、UCI、init、LuCI 页面或 IPK 内容。
