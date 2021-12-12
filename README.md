# ssrp
simple small reverse proxy

```bash
bash <(curl -Ls https://raw.githubusercontent.com/go-bai/ssrp/master/install.sh)
```

cn install 

```bash
bash <(curl -Ls https://ghproxy.com/https://raw.githubusercontent.com/go-bai/ssrp/master/cninstall.sh)
```

配置文件在 `/etc/ssrp/config.toml`

操作命令

```bash
systemctl status ssrp
systemctl start ssrp
systemctl stop ssrp
systemctl restart ssrp
journalctl -f -u ssrp
```


todo

- 端口复用