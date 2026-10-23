# 10. LuCI 管理壳

## 当前定位

LuCI 只提供 OpenWrt 服务管理壳，不重写 SubConv Next 主 Web UI。默认随 all-in-one `subconv-next` 主包一起安装，普通用户不需要单独安装 `luci-app-subconv-next`。

菜单路径：

```text
Services / SubConv Next
```

## 安装文件

```text
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/overview.js
```

源码路径：

```text
openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json
openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json
openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/overview.js
```

## 页面能力

- 显示 Running / Stopped 状态。
- 显示端口、数据目录和 Web UI URL。
- 配置 `enabled`、`host`、`port`、`data_dir`、`public_base_url`、`log_level`。
- 提供 Start、Stop、Restart、Open Web UI 按钮。
- Save & Apply 后重启 `subconv-next` 使配置生效。

## ACL

ACL 允许 LuCI 页面读取和写入 `subconv-next` UCI 配置，并通过 ubus 查询/控制 `subconv-next` 服务。页面不显示完整订阅 URL、发布 token 或节点密钥。

## 单包发布

`openwrt/luci-app-subconv-next/` 保存 LuCI 源文件；默认 IPK 和 APK 构建会把这些文件直接合并到 `subconv-next` 主包中，不发布独立的 LuCI 包。

## 验收

- 安装 all-in-one 主包后 LuCI 出现 `Services / SubConv Next`。
- 页面可打开且不报 ACL 错误。
- 点击 Open Web UI 可打开 `http://router-ip:9876`。
- 修改端口并 Save & Apply 后，`curl http://127.0.0.1:{port}/healthz` 正常。
