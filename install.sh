#!/bin/sh

set -eu

REPOSITORY="github.com/pixingzoudaiyuexing/linode-tool"
SOURCE_URL="https://${REPOSITORY}/archive/refs/heads/main.tar.gz"
TARGET="/usr/local/bin/linode-tool"
GO_ROOT="/usr/local/lib/linode-tool/go"
GO_RELEASE="1.23.12"
MIN_GO_MAJOR=1
MIN_GO_MINOR=23

fail() {
	printf '%s\n' "错误: $*" >&2
	exit 1
}

run_privileged() {
	if [ "$AS_ROOT" -eq 1 ]; then
		"$@"
	else
		sudo "$@"
	fi
}

download() {
	url=$1
	destination=$2
	if command -v curl >/dev/null 2>&1; then
		curl -fL --retry 3 --connect-timeout 15 -o "$destination" "$url"
	elif command -v wget >/dev/null 2>&1; then
		wget -q --tries=3 --timeout=15 -O "$destination" "$url"
	else
		fail "未找到 curl 或 wget，无法下载 Go。"
	fi
}

go_meets_minimum() {
	go_version=$($1 version 2>/dev/null | awk '{print $3}' | sed 's/^go//' || true)
	[ -n "$go_version" ] || return 1
	go_major=$(printf '%s\n' "$go_version" | awk -F. '{print $1}')
	go_minor=$(printf '%s\n' "$go_version" | awk -F. '{print $2}')
	case "$go_major" in
		''|*[!0-9]*) return 1 ;;
	esac
	case "$go_minor" in
		''|*[!0-9]*) return 1 ;;
	esac
	[ "$go_major" -gt "$MIN_GO_MAJOR" ] || {
		[ "$go_major" -eq "$MIN_GO_MAJOR" ] && [ "$go_minor" -ge "$MIN_GO_MINOR" ]
	}
}

printf '%s\n' "开始安装 linode-tool..."

if [ ! -r /etc/os-release ]; then
	fail "无法读取 /etc/os-release，当前系统不是受支持的 Debian/Ubuntu 环境。"
fi

# shellcheck disable=SC1091
. /etc/os-release
case "${ID:-}" in
	debian|ubuntu)
		;;
	*)
		fail "当前系统 (${ID:-未知}) 不受支持，仅支持 Debian 和 Ubuntu。"
		;;
esac

command -v install >/dev/null 2>&1 || fail "未找到 install 命令，无法写入 ${TARGET}。"
command -v tar >/dev/null 2>&1 || fail "未找到 tar 命令，无法安装 Go。"
command -v sha256sum >/dev/null 2>&1 || fail "未找到 sha256sum 命令，无法校验 Go 安装包。"

if [ "$(id -u)" -eq 0 ]; then
	AS_ROOT=1
else
	AS_ROOT=0
	command -v sudo >/dev/null 2>&1 || fail "安装需要 root 权限，且未找到 sudo。请使用 root 运行或安装 sudo。"
fi

BUILD_DIR=$(mktemp -d "${TMPDIR:-/tmp}/linode-tool.XXXXXX") || fail "无法创建临时构建目录。"
trap 'rm -rf "$BUILD_DIR"' EXIT HUP INT TERM
mkdir -p "$BUILD_DIR/bin" || fail "无法创建临时 Go 安装目录。"

GO_BIN=$(command -v go 2>/dev/null || true)
if [ -z "$GO_BIN" ] && [ -x "$GO_ROOT/bin/go" ]; then
	GO_BIN="$GO_ROOT/bin/go"
fi

if [ -n "$GO_BIN" ] && go_meets_minimum "$GO_BIN"; then
	GO_VERSION=$($GO_BIN version | awk '{print $3}' | sed 's/^go//')
	printf '%s\n' "使用现有 Go ${GO_VERSION}。"
else
	case "$(uname -m)" in
		x86_64|amd64)
			GO_ARCH="amd64"
			GO_SHA256="d3847fef834e9db11bf64e3fb34db9c04db14e068eeb064f49af747010454f90"
			;;
		aarch64|arm64)
			GO_ARCH="arm64"
			GO_SHA256="52ce172f96e21da53b1ae9079808560d49b02ac86cecfa457217597f9bc28ab3"
			;;
		*)
			fail "不支持的 CPU 架构: $(uname -m)，目前仅支持 amd64 和 arm64。"
			;;
	esac

	GO_ARCHIVE="go${GO_RELEASE}.linux-${GO_ARCH}.tar.gz"
	GO_URL="https://go.dev/dl/${GO_ARCHIVE}"
	GO_ARCHIVE_PATH="${BUILD_DIR}/${GO_ARCHIVE}"
	printf '%s\n' "未找到 Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+，正在下载并安装 Go ${GO_RELEASE} (${GO_ARCH})..."
	if ! download "$GO_URL" "$GO_ARCHIVE_PATH"; then
		fail "Go 下载失败: ${GO_URL}"
	fi

	ACTUAL_SHA256=$(sha256sum "$GO_ARCHIVE_PATH" | awk '{print $1}')
	[ "$ACTUAL_SHA256" = "$GO_SHA256" ] || fail "Go 安装包 SHA-256 校验失败，已停止安装。"

	run_privileged install -d "$(dirname "$GO_ROOT")" || fail "无法创建 Go 安装目录。"
	run_privileged rm -rf "$GO_ROOT" || fail "无法清理旧的 linode-tool Go 安装目录。"
	run_privileged tar -C "$(dirname "$GO_ROOT")" -xzf "$GO_ARCHIVE_PATH" || fail "Go 解压失败。"
	GO_BIN="$GO_ROOT/bin/go"
	go_meets_minimum "$GO_BIN" || fail "Go 安装完成但版本验证失败。"
	GO_VERSION=$($GO_BIN version | awk '{print $3}' | sed 's/^go//')
	printf '%s\n' "Go ${GO_VERSION} 安装完成。"
fi

SOURCE_ARCHIVE_PATH="${BUILD_DIR}/source.tar.gz"
SOURCE_DIR="${BUILD_DIR}/source"
printf '%s\n' "正在下载 ${REPOSITORY} main 源码并构建..."
if ! download "$SOURCE_URL" "$SOURCE_ARCHIVE_PATH"; then
	fail "源码下载失败: ${SOURCE_URL}"
fi
mkdir -p "$SOURCE_DIR" || fail "无法创建临时源码目录。"
if ! tar -xzf "$SOURCE_ARCHIVE_PATH" -C "$SOURCE_DIR" --strip-components=1; then
	fail "源码解压失败。"
fi
if ! (cd "$SOURCE_DIR" && GOBIN="$BUILD_DIR/bin" "$GO_BIN" install ./cmd/linode-tool); then
	fail "Go 构建失败，请检查网络、Go 环境和仓库源码。"
fi

[ -x "$BUILD_DIR/bin/linode-tool" ] || fail "Go 构建完成但未找到目标二进制。"
run_privileged install -d /usr/local/bin || fail "无法创建 /usr/local/bin。"
run_privileged install -m 0755 "$BUILD_DIR/bin/linode-tool" "$TARGET" || fail "无法安装到 ${TARGET}。"

[ -x "$TARGET" ] || fail "安装后未找到可执行文件 ${TARGET}。"
if ! "$TARGET" help >/dev/null 2>&1; then
	fail "已写入 ${TARGET}，但执行验证失败。"
fi

printf '%s\n' "安装成功: ${TARGET}"
printf '%s\n' "运行 linode-tool 时可通过环境变量提供 LINODE_TOKEN，或按提示交互输入。"
