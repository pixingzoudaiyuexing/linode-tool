#!/bin/sh

set -eu

REPOSITORY="github.com/pixingzoudaiyuexing/linode-tool/cmd/linode-tool"
TARGET="/usr/local/bin/linode-tool"
MIN_GO_MAJOR=1
MIN_GO_MINOR=23

fail() {
	printf '%s\n' "错误: $*" >&2
	exit 1
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

command -v go >/dev/null 2>&1 || fail "未找到 Go，请先安装 Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR} 或更高版本。"
command -v install >/dev/null 2>&1 || fail "未找到 install 命令，无法写入 ${TARGET}。"

GO_VERSION=$(go version 2>/dev/null | awk '{print $3}' | sed 's/^go//')
[ -n "$GO_VERSION" ] || fail "无法读取 Go 版本，请检查 Go 安装是否正常。"

GO_MAJOR=$(printf '%s\n' "$GO_VERSION" | awk -F. '{print $1}')
GO_MINOR=$(printf '%s\n' "$GO_VERSION" | awk -F. '{print $2}')
case "$GO_MAJOR" in
	''|*[!0-9]*) fail "无法解析 Go 版本: ${GO_VERSION}" ;;
esac
case "$GO_MINOR" in
	''|*[!0-9]*) fail "无法解析 Go 版本: ${GO_VERSION}" ;;
esac

if [ "$GO_MAJOR" -lt "$MIN_GO_MAJOR" ] || {
	[ "$GO_MAJOR" -eq "$MIN_GO_MAJOR" ] && [ "$GO_MINOR" -lt "$MIN_GO_MINOR" ];
}; then
	fail "Go 版本 ${GO_VERSION} 过低，需要 Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR} 或更高版本。"
fi

if [ "$(id -u)" -eq 0 ]; then
	AS_ROOT=1
else
	AS_ROOT=0
	command -v sudo >/dev/null 2>&1 || fail "写入 ${TARGET} 需要 root 权限，且未找到 sudo。请使用 root 运行或安装 sudo。"
fi

BUILD_DIR=$(mktemp -d "${TMPDIR:-/tmp}/linode-tool.XXXXXX") || fail "无法创建临时构建目录。"
trap 'rm -rf "$BUILD_DIR"' EXIT HUP INT TERM
mkdir -p "$BUILD_DIR/bin" || fail "无法创建临时 Go 安装目录。"

printf '%s\n' "检测到 Debian/Ubuntu，Go ${GO_VERSION}。正在构建 ${REPOSITORY}@main..."
if ! GOBIN="$BUILD_DIR/bin" go install "${REPOSITORY}@main"; then
	fail "Go 构建失败，请检查网络、Go 环境和仓库可访问性。"
fi

[ -x "$BUILD_DIR/bin/linode-tool" ] || fail "Go 构建完成但未找到目标二进制。"

if [ "$AS_ROOT" -eq 1 ]; then
	install -d /usr/local/bin || fail "无法创建 /usr/local/bin。"
	install -m 0755 "$BUILD_DIR/bin/linode-tool" "$TARGET" || fail "无法安装到 ${TARGET}。"
else
	sudo install -d /usr/local/bin || fail "无法创建 /usr/local/bin，请检查 sudo 权限。"
	sudo install -m 0755 "$BUILD_DIR/bin/linode-tool" "$TARGET" || fail "无法安装到 ${TARGET}，请检查 sudo 权限。"
fi

[ -x "$TARGET" ] || fail "安装后未找到可执行文件 ${TARGET}。"
if ! "$TARGET" >/dev/null 2>&1; then
	fail "已写入 ${TARGET}，但执行验证失败。"
fi

printf '%s\n' "安装成功: ${TARGET}"
printf '%s\n' "请先设置 LINODE_TOKEN，再运行: linode-tool create"
