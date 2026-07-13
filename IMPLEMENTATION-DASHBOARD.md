# SubConv Next Dashboard Implementation

实施日期：2026-07-11
实施范围：第四阶段 Dashboard 页面。未实现 Service、Network、Storage、Logs 或 About 页面，也未调整 `luci.subconv` RPC 架构。

## 1. 修改文件

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/dashboard.js`
  - 将 RPC 连通性占位页替换为真实 Dashboard。
  - 使用 `view.extend`、只读 `form.JSONMap + form.GridSection`、`E()` 和 `dom.content()`。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/dashboard.css`
  - 新增仅作用于 `.subconv-next-app` 的 Dashboard 样式。
  - 提供桌面三列、平板两列、手机单列布局。
- `Makefile`
  - 新增 `make package`，调用现有 portable OpenWrt IPK 脚本。
- `openwrt/subconv-next/Makefile`
  - 安装 Dashboard CSS。
  - package release 提升到 `1.0.7-14`。
- `openwrt/luci-app-subconv-next/Makefile`
  - package release 提升到 `1.0.7-14`。
- `scripts/package-openwrt-ipk-sdk.sh`
  - 打包并验证 Dashboard CSS。
  - 默认 release 提升到 `14`。
- `scripts/package-openwrt-luci-ipk-sdk.sh`
  - 独立 LuCI 包打包并验证 Dashboard CSS。
  - 默认 release 提升到 `14`。

## 2. RPC 调用

Dashboard 只消费第三阶段已有的 `luci.subconv` 接口。

页面加载和手动刷新并行调用：

```text
luci.subconv.status
luci.subconv.get_config
```

状态字段：

- `running`
- `pid`
- `version`
- `uptime`

配置字段：

- `listen`
- `port`
- `data_dir`
- `log_level`
- `public_base_url`

服务按钮调用：

```text
luci.subconv.start
luci.subconv.stop
luci.subconv.restart
```

动作期间会禁用服务按钮和刷新按钮。RPC 返回 `success: true` 时显示 `Start/Stop/Restart successful` 通知并重新读取 Dashboard 数据；失败时显示包含 RPC 错误信息的失败通知。

Stop 操作在调用 RPC 前显示确认框。确认内容说明 Web UI 和订阅刷新服务将停止，但已发布文件不会被删除。

## 3. UI 结构

页面包含三个同级区域：

1. Service Status
   - Running：绿色状态点和文字。
   - Stopped：红色状态点和文字。
   - Error：橙色状态点和 RPC 不可用说明。
   - 展示 Version、PID 和 Uptime。
2. Configuration
   - 使用只读 `form.GridSection` 展示 Listen、Port、Data Dir 和 Log Level。
   - 配置 RPC 不可用时显示 Unavailable，不伪造默认值。
3. Actions
   - Start、Stop、Restart。
   - Open Web UI。

页面顶部提供手动 Refresh。未实现后台定时轮询。

Open Web UI 按本阶段约束使用 `get_config.public_base_url`。仅当值存在且浏览器解析为 HTTP(S) URL 时显示；为空或非法时隐藏按钮并显示配置缺失状态。

所有动态内容通过 `E()`、DOM 节点属性或 `textContent` 写入。页面没有使用：

- `innerHTML`
- `rawhtml`
- jQuery
- Vue
- React
- mock 数据

响应式断点：

- `>= 1024px`：三列。
- `640-1023px`：状态和配置两列，操作区整行。
- `< 640px`：单列，操作按钮两列排列，Open Web UI 整行。

## 4. 测试结果

### 已完成

- `node --check`：Dashboard JavaScript 通过。
- `dash -n`：两套 OpenWrt 打包脚本通过。
- `git diff --check`：通过。
- 静态安全检查：Dashboard 未出现 `innerHTML`、`rawhtml`、`service.list` 或 `setInitAction`。
- `make package`：通过。
- all-in-one 包：

  ```text
  dist/subconv-next_1.0.7-14_aarch64_generic.ipk
  ```

- 独立 LuCI 包：

  ```text
  dist/luci-app-subconv-next_1.0.7-14_all.ipk
  ```

- 两种包均验证包含：

  ```text
  /usr/libexec/rpcd/luci.subconv
  /usr/share/rpcd/acl.d/luci-app-subconv-next.json
  /usr/share/luci/menu.d/luci-app-subconv-next.json
  /www/luci-static/resources/view/subconv-next/dashboard.js
  /www/luci-static/resources/subconv-next/dashboard.css
  ```

- rpcd 文件模式为 `0755`，Dashboard JS/CSS、ACL 和菜单为 `0644`。
- `go test ./...`：通过，确认现有服务和渲染代码未回归。

### 需要在真实 OpenWrt 上完成

开发机没有 `opkg`、`ubus`、rpcd、LuCI 和浏览器运行环境，因此未伪造安装或浏览器测试结果。设备端应执行：

```sh
opkg install /tmp/subconv-next_1.0.7-14_aarch64_generic.ipk
ubus -v list luci.subconv
ubus call luci.subconv status '{}'
ubus call luci.subconv get_config '{}'
```

然后访问：

```text
/cgi-bin/luci/admin/services/subconv-next
```

验证 Running、Stopped、Error 三种状态，手动刷新，三项服务动作，Stop 确认框，通知消息和 Open Web UI 显示条件。

## 5. 已知问题

- 当前 `status` RPC 只提供进程运行状态，尚无端口监听、healthz 或 respawn 诊断字段，因此 Dashboard 的 Error 仅表示 RPC/响应异常，不能区分“进程存在但健康检查失败”。这些字段需要后续 RPC 阶段明确扩展，第四阶段没有修改 RPC。
- 按本阶段要求，Open Web UI 使用 `public_base_url`。这与 `UI-DESIGN.md` 中未来使用专用 `web_ui_url` 字段的长期方案不同；当前 RPC 没有该字段，因此未自行推导监听地址 URL。
- Dashboard 动作成功后立即重新读取状态，没有后台自动轮询。procd 状态收敛较慢的设备可能短暂显示动作前状态，用户可使用 Refresh 再次读取。
- 页面未在真实 OpenWrt 主题和移动浏览器上截图验收，最终主题兼容性需要设备端确认。
