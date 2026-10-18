# ZeorCode 后端

项目架构、本地启动和配置说明见[根目录 README](../README.md)，GitHub Actions、Nginx 与 systemd 发布流程见[自动部署说明](../.github/AUTOMATIC_DEPLOYMENT.md)。

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
