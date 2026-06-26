# 13. 构建、发布与 CI

## Makefile

仓库根目录创建：

```makefile
APP=subconv-next

.PHONY: test build clean run

test:
	go test ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP) ./cmd/subconv-next

run:
	go run ./cmd/subconv-next serve --config ./testdata/config/basic.json

clean:
	rm -rf bin dist
```

## 本地构建

```sh
make test
make build
./bin/subconv-next version
```

## 多架构交叉编译

先提供普通 Linux 交叉编译产物：

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/subconv-next_linux_amd64 ./cmd/subconv-next
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/subconv-next_linux_arm64 ./cmd/subconv-next
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o dist/subconv-next_linux_armv7 ./cmd/subconv-next
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o dist/subconv-next_linux_mipsle_softfloat ./cmd/subconv-next
```

## OpenWrt SDK 构建策略

V1 允许两条路线：

### 路线 A：SDK 内构建 Go

如果 SDK 和 feeds 支持 Go helper，则在 OpenWrt Makefile 中直接编译。

### 路线 B：预构建二进制打包

如果 SDK 内 Go 构建不稳定，先用 GitHub Actions 交叉编译 binary，再让 OpenWrt package 安装对应架构 binary。

Codex 优先保证路线 B 可用。

## GitHub Actions

创建 `.github/workflows/ci.yml`：

- checkout
- setup-go
- go test
- build linux amd64/arm64/armv7/mipsle
- upload artifacts

### Auto Release

`.github/workflows/auto-release.yml` 会在 `main` 每次更新后自动发布 GitHub Release：

- 自动生成版本号：读取远端最新 `v1.0.x` tag 并递增 patch，Git tag 为 `v<version>`。失败的 workflow 不会消耗版本号。
- 用该版本号构建 Linux 多架构二进制。
- 使用仓库内 portable `ipkg-build` 打包 `aarch64_generic` all-in-one OpenWrt IPK，不需要配置 OpenWrt SDK URL。
- 推送 Docker 镜像到 GHCR：`latest` 和当前版本 tag。GHCR 推送失败不会阻断 GitHub Release 和 OpenWrt IPK 上传。
- 创建 Git tag 和 GitHub Release，并上传二进制、OpenWrt IPK 与 `checksums.txt`。

GHCR 默认使用 `GITHUB_TOKEN` 推送，并在镜像中写入 `org.opencontainers.image.source` 以关联当前仓库。如果 GHCR 返回 `permission_denied: write_package`，需要在 package 设置中给当前仓库 Actions 写入权限，或配置 repository secrets。配置前自动发布仍会继续创建 Git tag、GitHub Release 和上传 OpenWrt IPK。

- `GHCR_USERNAME`：PAT 所属 GitHub 用户名，可省略并默认使用 `github.actor`。
- `GHCR_TOKEN`：classic PAT，至少包含 `write:packages` 权限。

手动触发 `workflow_dispatch` 时可以勾选 `reset_v1_versions`，发布前会尝试删除 `v1.0.1+` 的 Release/tag，再从 `v1.0.1` 重新生成。已经被 GitHub 标记为 immutable 的 Release 可能需要先在 GitHub 页面手动删除。

提交信息包含 `[skip release]` 时，只保留普通 CI，不自动发版。

## Release Artifacts

V1 release 包含：

```text
subconv-next-linux-amd64
subconv-next-linux-arm64
subconv-next-linux-armv7
subconv-next-linux-mips-softfloat
subconv-next-linux-mipsle-softfloat
subconv-next_<version>-1_aarch64_generic.ipk
checksums.txt
```

## 版本命名

```text
v0.1.0
```

CLI 输出：

```sh
subconv-next version
```

结果：

```text
subconv-next version 0.1.0 commit <sha> built <date>
```
