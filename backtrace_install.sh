#!/bin/bash
# From https://github.com/ilychi/backtrace
# 2025.01.05

set -e

# 清理旧版本
rm -rf /usr/bin/backtrace

# 获取系统信息
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

# 获取最新版本
REPO="ilychi/backtrace"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
[ -z "$LATEST_VERSION" ] && LATEST_VERSION="latest"

# 映射架构名称
case $arch in
  "x86_64" | "amd64")
    arch="amd64"
    ;;
  "aarch64" | "arm64")
    arch="arm64"
    ;;
  *)
    echo "不支持的架构: $arch"
    exit 1
    ;;
esac

# 下载对应版本
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/backtrace-${os}-${arch}"
if [ "$os" = "windows" ]; then
    DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
fi

echo "正在下载 backtrace..."
echo "下载地址: $DOWNLOAD_URL"
curl -L "$DOWNLOAD_URL" -o backtrace

# 设置权限并安装
chmod +x backtrace
mv backtrace /usr/bin/

echo "backtrace 已成功安装！"
echo "版本: $LATEST_VERSION"
echo "系统: $os"
echo "架构: $arch"
