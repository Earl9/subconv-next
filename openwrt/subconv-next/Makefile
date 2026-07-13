include $(TOPDIR)/rules.mk

PKG_NAME:=subconv-next
PKG_VERSION:=1.0.7
PKG_RELEASE:=33
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
  DEPENDS:=+ca-bundle +uci
endef

define Package/subconv-next/description
  Self-hosted subscription converter and Mihomo YAML generator.
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
	$(INSTALL_DIR) $(1)/usr/share/subconv-next
	$(INSTALL_DATA) ./files/subconv-next.config $(1)/usr/share/subconv-next/subconv-next.config
endef

define Package/subconv-next/preinst
#!/bin/sh
root="$${IPKG_INSTROOT:-}"
config_path="$$root/etc/config/subconv-next"
backup_path="$$root/tmp/subconv-next.config.keep"
persistent_backup="$$root/etc/config/subconv-next-opkg.backup"
if [ -f "$$config_path" ]; then
	mkdir -p "$$root/tmp"
	cp "$$config_path" "$$backup_path" || true
	cp "$$config_path" "$$persistent_backup" || true
	chmod 0600 "$$persistent_backup" 2>/dev/null || true
fi
exit 0
endef

define Package/subconv-next/postinst
#!/bin/sh
root="$${IPKG_INSTROOT:-}"
config_path="$$root/etc/config/subconv-next"
default_config="$$root/usr/share/subconv-next/subconv-next.config"
backup_path="$$root/tmp/subconv-next.config.keep"
persistent_backup="$$root/etc/config/subconv-next-opkg.backup"
if [ -f "$$backup_path" ]; then
	mkdir -p "$$root/etc/config"
	cp "$$backup_path" "$$config_path"
	chmod 0644 "$$config_path"
	cp "$$config_path" "$$persistent_backup" || true
	chmod 0600 "$$persistent_backup" 2>/dev/null || true
elif [ -f "$$persistent_backup" ]; then
	mkdir -p "$$root/etc/config"
	cp "$$persistent_backup" "$$config_path"
	chmod 0644 "$$config_path"
elif [ ! -e "$$config_path" ] && [ -f "$$default_config" ]; then
	mkdir -p "$$root/etc/config"
	cp "$$default_config" "$$config_path"
	chmod 0644 "$$config_path"
fi
[ ! -f "$$persistent_backup" ] || chmod 0600 "$$persistent_backup" 2>/dev/null || true
rm -f "$$backup_path"
[ -n "$${IPKG_INSTROOT}" ] && exit 0
if [ -x /etc/init.d/subconv-next ]; then
	/etc/init.d/subconv-next enable >/dev/null 2>&1 || true
	enabled="$$(uci -q get subconv-next.main.enabled)"
	[ -z "$$enabled" ] && enabled="1"
	if [ "$$enabled" = "1" ]; then
		/etc/init.d/subconv-next start >/dev/null 2>&1 || true
	fi
fi
exit 0
endef

define Package/subconv-next/prerm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] && exit 0
if [ -x /etc/init.d/subconv-next ]; then
	if [ -f /etc/config/subconv-next ]; then
		cp /etc/config/subconv-next /etc/config/subconv-next-opkg.backup >/dev/null 2>&1 || true
		chmod 0600 /etc/config/subconv-next-opkg.backup >/dev/null 2>&1 || true
	fi
	/etc/init.d/subconv-next stop >/dev/null 2>&1 || true
	/etc/init.d/subconv-next disable >/dev/null 2>&1 || true
fi
exit 0
endef

$(eval $(call BuildPackage,subconv-next))
