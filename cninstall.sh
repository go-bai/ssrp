#!/bin/sh

arch=$(arch)
if [[ $arch == "x86_64" || $arch == "x64" || $arch == "amd64" ]]; then
  arch="amd64"
elif [[ $arch == "aarch64" || $arch == "arm64" ]]; then
  arch="arm64"
else
  echo -e "本软件不支持此系统"
  exit 2
fi

echo "架构: ${arch}"

last_version=$(curl -Ls "https://api.github.com/repos/go-bai/ssrp/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
echo -e "检测到 ssrp 最新版本：${last_version}，开始安装"
rm -rf /etc/ssrp
mkdir /etc/ssrp/ -p
cd /etc/ssrp

wget https://ghproxy.com/https://github.com/go-bai/ssrp/releases/download/${last_version}/ssrp-${last_version}-linux-${arch}.tar.gz
tar -xzf ssrp-${last_version}-linux-${arch}.tar.gz && rm -f ssrp-${last_version}-linux-${arch}.tar.gz
chmod +x ssrp

wget https://ghproxy.com/https://raw.githubusercontent.com/go-bai/ssrp/master/config.toml

cat > /etc/systemd/system/ssrp.service << EOF
[Unit]
Description=ssrp - simple small reverse proxy
After=network.target
[Service]
User=$INSTALL_USER
Type=simple
Restart=always
RestartSec=5s
ExecStart=/etc/ssrp/ssrp
WorkingDirectory=/etc/ssrp/
LimitNPROC=10000
LimitNOFILE=1000000
[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl restart ssrp
systemctl enable ssrp
systemctl status ssrp