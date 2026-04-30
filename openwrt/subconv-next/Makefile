include $(TOPDIR)/rules.mk

PKG_NAME:=subconv-next
PKG_VERSION:=1.0.0
PKG_RELEASE:=3
PKG_LICENSE:=MIT
PKG_MAINTAINER:=SubConv Next contributors

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)
PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

SUBCONV_NEXT_SRC ?= $(TOPDIR)/../subconv-next
SUBCONV_NEXT_VERSION ?= $(PKG_VERSION)-$(PKG_RELEASE)

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

GO_TARGET_ARCH:=$(GO_ARCH)
ifeq ($(GO_TARGET_ARCH),)
  ifeq ($(ARCH),aarch64)
    GO_TARGET_ARCH:=arm64
  else ifeq ($(ARCH),x86_64)
    GO_TARGET_ARCH:=amd64
  else ifeq ($(ARCH),i386)
    GO_TARGET_ARCH:=386
  else
    GO_TARGET_ARCH:=$(ARCH)
  endif
endif

define Package/subconv-next
  SECTION:=net
  CATEGORY:=Network
  TITLE:=SubConv Next Mihomo subscription generator
  URL:=https://github.com/subconv-next/subconv-next
  DEPENDS:=+ca-bundle +luci-base +rpcd +uci
endef

define Package/subconv-next/description
  Self-hosted subscription converter and Mihomo YAML generator.
endef

define Package/subconv-next/conffiles
/etc/config/subconv-next
endef

define Build/Prepare
	test -d "$(SUBCONV_NEXT_SRC)/cmd/subconv-next" || (echo "SUBCONV_NEXT_SRC must point to the subconv-next repository" >&2; exit 1)
	mkdir -p $(PKG_BUILD_DIR)
	tar -C $(SUBCONV_NEXT_SRC) \
		--exclude .git \
		--exclude data \
		--exclude dist \
		--exclude .cache \
		--exclude ./subconv-next \
		-cf - . | tar -C $(PKG_BUILD_DIR) -xf -
endef

define Build/Compile
	mkdir -p $(PKG_BUILD_DIR)/.cache/go-build $(PKG_BUILD_DIR)/.cache/go-mod
	(cd $(PKG_BUILD_DIR); \
		GO_BIN="$(STAGING_DIR_HOSTPKG)/bin/go"; \
		if [ ! -x "$$$$GO_BIN" ]; then \
			GO_BIN="$$$$(ls $(STAGING_DIR_HOSTPKG)/bin/go[0-9]* 2>/dev/null | sort -V | tail -n 1)"; \
		fi; \
		[ -x "$$$$GO_BIN" ] || (echo "OpenWrt Go host compiler not found in $(STAGING_DIR_HOSTPKG)/bin" >&2; exit 1); \
		CGO_ENABLED=0 \
		GOOS=linux \
		GOARCH=$(GO_TARGET_ARCH) \
		GOARM=$(GO_ARM) \
		GOMIPS=$(GO_MIPS) \
		GOMIPS64=$(GO_MIPS64) \
		GOCACHE=$(PKG_BUILD_DIR)/.cache/go-build \
		GOMODCACHE=$(PKG_BUILD_DIR)/.cache/go-mod \
		"$$$$GO_BIN" build -trimpath -ldflags="-s -w -X main.version=$(SUBCONV_NEXT_VERSION)" -o subconv-next ./cmd/subconv-next)
endef

define Package/subconv-next/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/subconv-next $(1)/usr/bin/subconv-next
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/subconv-next.init $(1)/etc/init.d/subconv-next
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/subconv-next.config $(1)/etc/config/subconv-next
	$(INSTALL_DIR) $(1)/etc/subconv-next/data
	$(INSTALL_DIR) $(1)/usr/share/luci/menu.d
	$(INSTALL_DATA) $(SUBCONV_NEXT_SRC)/openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json $(1)/usr/share/luci/menu.d/luci-app-subconv-next.json
	$(INSTALL_DIR) $(1)/usr/share/rpcd/acl.d
	$(INSTALL_DATA) $(SUBCONV_NEXT_SRC)/openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json $(1)/usr/share/rpcd/acl.d/luci-app-subconv-next.json
	$(INSTALL_DIR) $(1)/www/luci-static/resources/view/subconv-next
	$(INSTALL_DATA) $(SUBCONV_NEXT_SRC)/openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/index.js $(1)/www/luci-static/resources/view/subconv-next/index.js
endef

define Package/subconv-next/postinst
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] && exit 0
if [ -x /etc/init.d/subconv-next ]; then
	/etc/init.d/subconv-next enable || true
	enabled="$$(uci -q get subconv-next.main.enabled)"
	[ -z "$$enabled" ] && enabled="1"
	if [ "$$enabled" = "1" ]; then
		/etc/init.d/subconv-next start || true
	fi
fi
rm -f /tmp/luci-indexcache
rm -rf /tmp/luci-modulecache
/etc/init.d/rpcd restart >/dev/null 2>&1 || true
/etc/init.d/uhttpd restart >/dev/null 2>&1 || true
exit 0
endef

define Package/subconv-next/prerm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] && exit 0
if [ -x /etc/init.d/subconv-next ]; then
	/etc/init.d/subconv-next stop || true
	/etc/init.d/subconv-next disable || true
fi
exit 0
endef

$(eval $(call BuildPackage,subconv-next))
