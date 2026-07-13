# SubConv Next Service Page Implementation

实施日期：2026-07-11
实施范围：第五阶段 Service 页面。未实现 Network、Storage、Logs 或 About 页面。

## 1. 修改文件

### 新增

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/api.js`
  - 集中声明 `luci.subconv` RPC。
  - Dashboard 和 Service 页面不再分别声明相同接口。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/service.js`
  - Service 页面入口。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/service.css`
  - Service 页面响应式布局、诊断行和开关样式。

### 变更

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/dashboard.js`
  - 改为通过共享 `subconv-next.api` 模块调用 RPC。
- `openwrt/luci-app-subconv-next/root/usr/libexec/rpcd/luci.subconv`
  - 新增 `get_autostart` 和 `set_autostart`。
- `openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json`
  - read 增加 `get_autostart`。
  - write 增加 `set_autostart`。
- `openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json`
  - 新增 `admin/services/subconv-next/service`。
- `openwrt/subconv-next/Makefile`
- `openwrt/luci-app-subconv-next/Makefile`
- `scripts/package-openwrt-ipk-sdk.sh`
- `scripts/package-openwrt-luci-ipk-sdk.sh`
  - 安装并验证共享 API、Service View 和 CSS。
  - OpenWrt package release 提升到 `1.0.7-15`。

## 2. RPC 变化

### 共享前端 API

`subconv-next/api.js` 暴露：

```text
api.service.status()
api.service.start()
api.service.stop()
api.service.restart()
api.config.get()
api.config.save(values)
api.system.getAutostart()
api.system.setAutostart(enabled)
api.settle(promise)
```

`config.save()` 使用对象参数，只发送实际提供的字段，便于 Network 和 Storage 页面复用部分更新接口。

### `get_autostart`

调用固定 init 脚本的状态检查：

```text
/etc/init.d/subconv-next enabled
```

返回：

```json
{
  "available": true,
  "enabled": true
}
```

`available: false` 表示 init 脚本不存在或不可执行。

### `set_autostart`

输入：

```json
{
  "enabled": true
}
```

只允许布尔值，并映射到固定命令：

```text
/etc/init.d/subconv-next enable
/etc/init.d/subconv-next disable
```

接口不接受服务名、动作名或 shell 命令。

### ACL

read：

```text
status
get_config
get_autostart
```

write：

```text
start
stop
restart
set_config
set_autostart
```

未开放通用 init、shell、文件或直接 UCI 权限。

## 3. 页面说明

入口：

```text
/cgi-bin/luci/admin/services/subconv-next/service
```

页面包含三个同级区域。

### Runtime Status

展示真实 RPC 数据：

- Running、Stopped 或 Error。
- Version。
- PID。
- Uptime。
- Runtime、配置和系统启动三个 RPC 连通性诊断。

### Startup

两个开关保持独立：

1. Enabled in config
   - 数据源：`get_config.enabled`。
   - 写入：`set_config({ enabled })`。
   - 关闭后调用 `stop`，确保当前运行实例停止。
   - 开启后，仅当 Start on boot 已启用时调用 `start`。
2. Start on boot
   - 数据源：`get_autostart.enabled`。
   - 写入：`set_autostart(enabled)`。
   - 只管理 rc.d 开机启动链接，不自动修改 UCI enabled。

两个状态不一致时页面显示提示，不自动合并：

- UCI enabled 开启但 Start on boot 关闭。
- Start on boot 开启但 UCI enabled 关闭。

### Actions

- Start：成功显示 `Service started`。
- Stop：先显示 `Are you sure?` 确认框，成功显示 `Service stopped`。
- Restart：直接调用，成功显示 `Service restarted`。

操作期间禁用页面开关、服务按钮和刷新按钮。操作完成后重新读取全部状态，不进行后台自动轮询。

动态内容均通过 `E()`、`dom.content()` 或 `textContent` 安全更新。没有使用 `innerHTML`、`rawhtml`、jQuery 或第三方框架。

## 4. 测试结果

### rpcd 模拟测试

使用 SDK `jshn` 和 BusyBox ash 验证：

- `list` 精确包含八个方法：

  ```text
  status
  start
  stop
  restart
  get_config
  set_config
  get_autostart
  set_autostart
  ```

- `get_autostart` 正确读取启用和禁用状态。
- `set_autostart(true)` 只执行 `enable`。
- `set_autostart(false)` 只执行 `disable`。
- 非布尔输入返回失败，并且不执行 init 脚本。
- 原有 status、配置和服务动作接口仍保留。

### 静态和回归测试

- Dashboard、Service 和共享 API 通过 `node --check`。
- rpcd 和打包脚本通过 shell 语法检查。
- ACL 和菜单 JSON 通过 `jq` 解析。
- Service 和 Dashboard View 中没有直接 `rpc.declare`。
- View 中没有 `innerHTML`、`rawhtml`、`setInitAction` 或直接 init/UCI 调用。
- `git diff --check` 通过。
- `go test ./...` 通过。

### 构建结果

`make package` 通过：

```text
dist/subconv-next_1.0.7-15_aarch64_generic.ipk
```

独立 LuCI 包构建通过：

```text
dist/luci-app-subconv-next_1.0.7-15_all.ipk
```

两个包均包含：

```text
/usr/libexec/rpcd/luci.subconv
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/usr/share/luci/menu.d/luci-app-subconv-next.json
/www/luci-static/resources/subconv-next/api.js
/www/luci-static/resources/subconv-next/dashboard.css
/www/luci-static/resources/subconv-next/service.css
/www/luci-static/resources/view/subconv-next/dashboard.js
/www/luci-static/resources/view/subconv-next/service.js
```

rpcd 模式为 `0755`，其他 LuCI 资源为 `0644`。

### 真实 OpenWrt 验收

开发机没有 `opkg`、`ubus`、rpcd、LuCI 或浏览器运行环境，因此未伪造设备安装和浏览器结果。设备端应执行：

```sh
opkg install /tmp/subconv-next_1.0.7-15_aarch64_generic.ipk
ubus -v list luci.subconv
ubus call luci.subconv status '{}'
ubus call luci.subconv get_autostart '{}'
```

浏览器访问：

```text
/cgi-bin/luci/admin/services/subconv-next/service
```

重点验证两个开关的四种组合、Start/Stop/Restart、Stop 确认、手动刷新和移动端布局。

## 5. 已知问题

- 当前 `status` RPC 没有 Started At、respawn 状态、最近退出码或 healthz 诊断，因此 Service 页面只能展示现有运行数据和 RPC 连通性。第五阶段没有扩展这些不在目标内的诊断接口。
- 服务动作后立即重新读取状态。procd 状态收敛较慢时可能短暂显示旧状态，用户可手动刷新。
- `Enabled in config` 是即时提交开关，不提供单独的 Save/Apply 阶段。关闭时会提交 UCI 后停止服务；开启且 Start on boot 已启用时会提交 UCI 后启动服务。
- 页面尚未在真实 OpenWrt 主题、平板和手机浏览器中截图验收。
