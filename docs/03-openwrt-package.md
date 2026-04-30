# 03. OpenWrt 包设计

## 当前方案

V1.1 默认发布一个 all-in-one IPK：

```text
subconv-next_1.0.0-3_aarch64_generic.ipk
```

该包同时包含：

- `/usr/bin/subconv-next`
- `/etc/init.d/subconv-next`
- `/etc/config/subconv-next`
- `/etc/subconv-next/data`
- `/usr/share/luci/menu.d/luci-app-subconv-next.json`
- `/usr/share/rpcd/acl.d/luci-app-subconv-next.json`
- `/www/luci-static/resources/view/subconv-next/index.js`

安装后 `postinst` 会自动 enable 服务，并在 UCI `enabled=1` 或未设置时自动 start。脚本会清理 LuCI 缓存并重启 `rpcd`/`uhttpd`，使 `Services / SubConv Next` 菜单尽快出现。

卸载前 `prerm` 会 stop/disable 服务，但不会删除 `/etc/subconv-next/data`。

## 依赖

主包依赖：

```text
ca-bundle, luci-base, rpcd, uci
```

OpenWrt SDK Makefile 中对应：

```makefile
DEPENDS:=+ca-bundle +luci-base +rpcd +uci
```

## 打包方式

Kwrt/opkg 设备默认使用 SDK 自带 `scripts/ipkg-build`：

```sh
SDK_DIR=/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64 \
VERSION=1.0.0 \
RELEASE=3 \
ARCH=aarch64_generic \
./scripts/package-openwrt-ipk-sdk.sh
```

输出：

```text
dist/subconv-next_1.0.0-3_aarch64_generic.ipk
```

不要把 APK 改名为 IPK，也不要恢复旧的手工 `ar` 打包路径；Kwrt/opkg 已验证手工 `ar` 包会被判定为 malformed。

## 可选拆分包

仓库仍保留 `openwrt/luci-app-subconv-next/`，用于高级场景下单独构建 LuCI 包。但默认发布产物是 all-in-one 主包，普通用户不需要单独安装 `luci-app-subconv-next`。

## 验收

- `opkg install /tmp/subconv-next_1.0.0-3_aarch64_generic.ipk` 成功。
- 安装后 `/etc/init.d/subconv-next status` 显示服务已启动。
- `curl http://127.0.0.1:9876/healthz` 返回 OK。
- LuCI 出现 `Services / SubConv Next`。
- LuCI 页面能修改端口并重启服务。
- 卸载包不会删除 `/etc/subconv-next/data`。
