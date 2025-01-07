#!/bin/bash
#From https://github.com/ilychi/backtrace
#2024.01.07

# 安装 NextTrace
install_nexttrace() {
  echo "正在安装 NextTrace..."
  bash <(curl -Ls https://raw.githubusercontent.com/sjlleo/nexttrace/main/nt_install.sh)
  if ! command -v nexttrace &> /dev/null; then
    echo "NextTrace 安装失败，请手动安装"
    exit 1
  fi
  echo "NextTrace 安装成功"
}

# 检查 NextTrace 是否已安装
if ! command -v nexttrace &> /dev/null; then
  install_nexttrace
fi

rm -rf /usr/bin/backtrace
os=$(uname -s)
arch=$(uname -m)

check_cdn() {
  local o_url=$1
  for cdn_url in "${cdn_urls[@]}"; do
    if curl -sL -k "$cdn_url$o_url" --max-time 6 | grep -q "success" >/dev/null 2>&1; then
      export cdn_success_url="$cdn_url"
      return
    fi
    sleep 0.5
  done
  export cdn_success_url=""
}

check_cdn_file() {
  check_cdn "https://raw.githubusercontent.com/spiritLHLS/ecs/main/back/test"
  if [ -n "$cdn_success_url" ]; then
    echo "CDN available, using CDN"
  else
    echo "No CDN available, no use CDN"
  fi
}

cdn_urls=("https://cdn0.spiritlhl.top/" "http://cdn3.spiritlhl.net/" "http://cdn1.spiritlhl.net/" "http://cdn2.spiritlhl.net/")
check_cdn_file

case $os in
  Linux)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O backtrace "${cdn_success_url}https://github.com/ilychi/backtrace/releases/download/latest/backtrace-linux-amd64"
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O backtrace "${cdn_success_url}https://github.com/ilychi/backtrace/releases/download/latest/backtrace-linux-arm64"
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  Darwin)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O backtrace "${cdn_success_url}https://github.com/ilychi/backtrace/releases/download/latest/backtrace-darwin-amd64"
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O backtrace "${cdn_success_url}https://github.com/ilychi/backtrace/releases/download/latest/backtrace-darwin-arm64"
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "This script only supports Linux and macOS systems"
    exit 1
    ;;
esac

chmod 777 backtrace
mv backtrace /usr/bin/backtrace
echo "Installation completed. You can now use 'backtrace' command."
echo
echo "Usage:"
echo "  backtrace           - Run backtrace with default settings"
echo "  backtrace -h       - Show help information"
echo "  backtrace -v       - Show version information"
echo "  backtrace -e       - Enable logging"
echo "  backtrace -s=false - Disable IP information display"
echo
echo "Example:"
echo "  backtrace          - Trace route to multiple destinations"
echo "  backtrace -e       - Trace route with logging enabled"
