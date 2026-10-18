# ZeorCode 后端

完整的 Ubuntu 生产环境安装、Nginx、MySQL、Redis、go-judge、systemd、HTTPS、更新与备份说明见 [DEPLOYMENT.md](DEPLOYMENT.md)。

## 构建

```bash
go build -o zeorcode .
```

## 运行

```bash
# 启动 HTTP 服务
./zeorcode http

# 启动判题 worker（默认 4 并发）
./zeorcode judge

# 自定义判题并发数
./zeorcode judge -c 8
./zeorcode judge --concurrency 16

# 查看帮助
./zeorcode --help
./zeorcode judge --help
```
