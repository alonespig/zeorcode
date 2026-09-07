# ZOJ 生产环境部署教程

本文档说明如何把当前 ZOJ 前后端部署到一台 Ubuntu 服务器。前后端均在本地 Windows 开发机完成编译，服务器只接收 Linux 后端二进制、Vue `dist` 和数据库脚本，不在服务器安装 Go、Node.js，也不上传完整源码。内容覆盖 MySQL、Redis、go-judge、构建产物上传、Nginx、域名与免费 SSL 证书、无域名访问、systemd 守护、更新、备份和排错。

> 推荐环境：Ubuntu 24.04/26.04 LTS、4 核 CPU、8 GB 内存、80 GB SSD、5～10 Mbps 带宽。2 核 4 GB 可用于测试或很小的班级，但本地评测并发应降到 1～2。

## 1. 部署结构

本教程使用以下结构：

```text
浏览器
  |
  | 80 / 443
  v
Nginx
  |-- /                 -> /var/www/zoj（Vue 构建产物）
  |-- /api/             -> 127.0.0.1:9090（ZOJ HTTP 服务）
  `-- /uploads/         -> 127.0.0.1:9090（上传图片）

ZOJ HTTP 服务
  |-- MySQL             -> 127.0.0.1:3306
  `-- Redis             -> 127.0.0.1:6379

ZOJ judge worker
  |-- MySQL / Redis
  `-- go-judge          -> 127.0.0.1:5050
```

发布链路如下：

```text
本地 Windows 开发机
  |-- Go 交叉编译 Linux 后端
  |-- npm 构建 Vue dist
  `-- 打包 zoj-release-linux-amd64.tar.gz
             |
             | scp / SSH
             v
Ubuntu 服务器
  |-- /opt/zoj/backend/zoj（运行中的后端）
  |-- /var/www/zoj（运行中的前端）
  `-- /opt/zoj/releases（保留上传的版本，用于检查和回滚）
```

后端二进制需要启动两个独立进程：

- `zoj http`：提供 HTTP API，默认由 Nginx 反向代理到 `127.0.0.1:9090`。
- `zoj judge -c 4`：消费 Redis Stream 中的评测任务并调用 go-judge。

Redis 在本项目中不仅是缓存，还保存登录 session、限流状态、评测队列和 SSE 评测通知，因此不能把它当成随时可以清空的普通缓存。

数据库不会自动执行 AutoMigrate。全新数据库应导入 `database/schema.sql`；已有数据库只能按顺序执行尚未应用的 `database/migrations/*.sql`。

## 2. 准备服务器

以下命令默认使用具有 `sudo` 权限的普通用户执行。

### 2.1 设置时区和主机名

```bash
sudo timedatectl set-timezone Asia/Shanghai
sudo hostnamectl set-hostname zoj-server
timedatectl
```

### 2.2 安装基础工具

```bash
sudo apt update
sudo apt upgrade -y
sudo apt install -y \
  ca-certificates curl file gnupg unzip xz-utils rsync dnsutils snapd \
  nginx mysql-server apache2-utils
```

启动基础服务：

```bash
sudo systemctl enable --now nginx
sudo systemctl enable --now mysql
```

### 2.3 创建运行用户和目录

```bash
sudo useradd --system --create-home \
  --home-dir /opt/zoj --shell /usr/sbin/nologin zoj

sudo install -d -o zoj -g zoj /opt/zoj/backend
sudo install -d -o zoj -g zoj /opt/zoj/backend/uploads
sudo install -d -o zoj -g zoj /opt/zoj/backend/ojdata/problems
sudo install -d -o zoj -g zoj /opt/zoj/backend/runtime
sudo install -d -o "$USER" -g "$USER" /opt/zoj/releases
sudo install -d -o root -g root -m 0755 /var/www/zoj
```

后端使用相对路径读取 `config.yaml`、`uploads`、`ojdata/problems` 和 `runtime`，因此 systemd 中的 `WorkingDirectory` 必须保持为 `/opt/zoj/backend`。

## 3. 配置安全组与防火墙

云服务器安全组只需要放行：

| 端口 | 用途 | 来源 |
| --- | --- | --- |
| `22/tcp` | SSH | 最好仅允许管理员固定 IP |
| `80/tcp` | HTTP、申请证书 | 公网 |
| `443/tcp` | HTTPS | 公网 |

不要向公网开放 `9090`、`5050`、`3306`、`6379`。

如果同时使用 UFW，先确认 SSH 端口正确再启用：

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status verbose
```

go-judge 必须明确监听 `127.0.0.1:5050`，不能监听 `0.0.0.0:5050`，也不要在腾讯云安全组中放行 5050。

## 4. 在本地编译并制作发布包

以下命令在本地 Windows PowerShell 中执行，不是在服务器执行。服务器不需要安装 Go、Node.js、npm，也不需要保存项目源码。

本地需要：

- Go 1.25 或更高版本，与后端 `go.mod` 一致。
- Node.js `^20.19.0 || >=22.12.0`，与前端 `package.json` 一致。
- Windows 自带的 OpenSSH `scp` 和 `tar`；如果命令不存在，可在“可选功能”中安装 OpenSSH 客户端。

先在服务器确认 CPU 架构：

```bash
uname -m
```

- 返回 `x86_64`：本地使用 `GOARCH=amd64`。
- 返回 `aarch64` 或 `arm64`：本地使用 `GOARCH=arm64`。

下面以常见的腾讯云 x86-64 服务器为例。

### 4.1 本地测试并交叉编译后端

```powershell
Set-Location E:\fastoj\oj-backend

go version
go mod download
go test ./...

if (Test-Path .\release\zoj) {
    Remove-Item -Recurse -Force .\release\zoj
}
New-Item -ItemType Directory -Force .\release\zoj\backend | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -trimpath -ldflags="-s -w" -o .\release\zoj\backend\zoj .

Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

go version -m .\release\zoj\backend\zoj
```

不能直接把 Windows 下普通执行 `go build` 产生的 `.exe` 上传到 Ubuntu；必须设置 `GOOS=linux`。本项目后端不依赖 CGO，使用 `CGO_ENABLED=0` 可以生成不依赖服务器 C 运行库的 Linux 二进制。

如果服务器是 ARM64，只把上面的这一行改为：

```powershell
$env:GOARCH = "arm64"
```

### 4.2 本地构建前端

前端生产配置应使用同源 API：

```dotenv
VITE_API_BASE_URL=/api
```

在本地执行：

```powershell
Set-Location E:\fastoj\oj-frontend

node --version
npm --version
npm ci
npm run build

New-Item -ItemType Directory -Force E:\fastoj\oj-backend\release\zoj\frontend | Out-Null
Copy-Item -Recurse -Force .\dist\* E:\fastoj\oj-backend\release\zoj\frontend\
```

构建成功后只需要上传 `dist` 的内容，不需要上传 `node_modules`、前端源码，也不要在生产服务器运行 `npm run dev` 或 `npm run preview`。

### 4.3 加入数据库脚本并打包

回到后端目录，把数据库初始化和迁移脚本放入同一个发布包：

```powershell
Set-Location E:\fastoj\oj-backend

New-Item -ItemType Directory -Force .\release\zoj\database | Out-Null
Copy-Item -Recurse -Force .\database\* .\release\zoj\database\
Copy-Item -Force .\DEPLOYMENT.md .\release\zoj\DEPLOYMENT.md

tar -czf .\release\zoj-release-linux-amd64.tar.gz -C .\release zoj
Get-FileHash .\release\zoj-release-linux-amd64.tar.gz -Algorithm SHA256
```

最终上传的文件只有：

```text
release/zoj-release-linux-amd64.tar.gz
```

发布包中不包含生产 `config.yaml`、上传图片、题目测试数据或运行日志，这些文件必须保留在服务器上。

## 5. 安装和配置 MySQL

项目 schema 使用 `utf8mb4_0900_ai_ci`，应使用 MySQL 8.0 或 8.4 LTS，不要使用旧版 MySQL 5.7。

Ubuntu 自带包可以直接用于当前规模。若希望固定到 Oracle MySQL 8.4 LTS，可按 [MySQL 官方 APT 仓库说明](https://dev.mysql.com/doc/refman/8.4/en/linux-installation-apt-repo.html)配置仓库后再安装。

### 5.1 基础加固

```bash
sudo mysql_secure_installation
sudo systemctl status mysql --no-pager
```

确认 MySQL 只监听本机。Ubuntu 默认配置通常已经如此：

```bash
sudo ss -lntp | grep 3306
```

如需手动设置，在 `/etc/mysql/mysql.conf.d/mysqld.cnf` 的 `[mysqld]` 下确认：

```ini
bind-address = 127.0.0.1
```

修改后执行：

```bash
sudo systemctl restart mysql
```

### 5.2 创建数据库和专用账号

先生成一个强密码并妥善保存：

```bash
openssl rand -base64 36
```

进入 MySQL：

```bash
sudo mysql
```

执行以下 SQL，把示例密码换成真实密码：

```sql
CREATE DATABASE `oj_db`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;

CREATE USER 'zoj'@'127.0.0.1'
  IDENTIFIED BY 'NuQhpeVyXm/qbBqNL1TvGHKBd91gZ/bU2BF2zEWSTXzwuhM/';

GRANT ALL PRIVILEGES ON `oj_db`.* TO 'zoj'@'127.0.0.1';
FLUSH PRIVILEGES;
EXIT;
```

不要让应用使用 MySQL `root` 账号，也不要创建 `'zoj'@'%'`。

## 6. 安装和配置 Redis

可以直接安装 Ubuntu 的 `redis-server`，也可以按 [Redis 官方 APT 安装说明](https://redis.io/docs/latest/operate/oss_and_stack/install/install-stack/)使用 Redis 官方仓库。对于几十名用户，Ubuntu 包已经足够。

```bash
sudo apt install -y redis-server
sudo systemctl enable --now redis-server
```

生成 Redis 密码：

```bash
openssl rand -base64 36
```

编辑 `/etc/redis/redis.conf`，确认或修改以下项目：

```conf
bind 127.0.0.1 -::1
protected-mode yes
port 6379

requirepass 9d0pCnIR/rnkHQ0675INsIX9dVfDjX4XkPA2t4+p8eqC6krE

appendonly yes
appendfsync everysec

# Redis 包含评测队列和登录态，不要使用会主动淘汰 key 的策略。
# 4 GB 服务器可改为 256mb，8 GB 服务器可改为 512mb。
maxmemory 512mb
maxmemory-policy noeviction
```

重启并验证：

```bash
sudo systemctl restart redis-server
redis-cli -a '替换为强密码' ping
```

应返回 `PONG`。命令行中的密码可能进入 shell history；正式操作可使用 `REDISCLI_AUTH` 临时环境变量：

```bash
export REDISCLI_AUTH='替换为强密码'
redis-cli ping
unset REDISCLI_AUTH
```

## 7. 上传并解压发布包

### 7.1 从本地上传

在本地 Windows PowerShell 中执行，把用户名和公网 IP 替换为服务器的实际值：

```powershell
scp E:\fastoj\oj-backend\release\zoj-release-linux-amd64.tar.gz `
  服务器用户@服务器公网IP:/tmp/
```

第一次连接会询问是否信任服务器指纹，核对腾讯云服务器信息后输入 `yes`，再输入 SSH 密码。SSH 只负责登录和传输文件，不是网站使用的 SSL 证书。

也可以先配置 SSH 密钥，然后使用相同的 `scp` 命令免密码上传。

### 7.2 在服务器校验并解压

先用 SSH 登录服务器：

```powershell
ssh 服务器用户@服务器公网IP
```

如果上传前记录了 SHA-256，可在服务器核对：

```bash
sha256sum /tmp/zoj-release-linux-amd64.tar.gz
```

创建一个带时间的版本目录并解压：

```bash
ZOJ_RELEASE_ID="$(date +%Y%m%d%H%M%S)"
ZOJ_RELEASE_DIR="/opt/zoj/releases/${ZOJ_RELEASE_ID}"

mkdir -p "${ZOJ_RELEASE_DIR}"
tar -xzf /tmp/zoj-release-linux-amd64.tar.gz \
  -C "${ZOJ_RELEASE_DIR}" --strip-components=1

find "${ZOJ_RELEASE_DIR}" -maxdepth 2 -type f | sort
```

应至少看到：

```text
backend/zoj
frontend/index.html
database/schema.sql
```

确认无误后，把当前发布包指针切换到这个目录：

```bash
sudo ln -sfn "${ZOJ_RELEASE_DIR}" /opt/zoj/current
rm -f /tmp/zoj-release-linux-amd64.tar.gz
```

`/opt/zoj/current` 只指向当前上传的构建产物。生产配置、上传文件、测试数据始终保存在 `/opt/zoj/backend`，切换发布包不会覆盖这些持久数据。

## 8. 初始化数据库

### 8.1 全新数据库

`schema.sql` 会删除并重建表，只能用于全新数据库：

```bash
mysql -h 127.0.0.1 -u zoj -p oj_db \
  < /opt/zoj/current/database/schema.sql
```

验证：

```bash
mysql -h 127.0.0.1 -u zoj -p -D oj_db \
  -e "SELECT COUNT(*) AS table_count FROM information_schema.tables WHERE table_schema='oj_db';"
```

### 8.2 已有数据库升级

先备份数据库，然后只执行尚未应用的迁移，例如：

```bash
mysql -h 127.0.0.1 -u zoj -p oj_db \
  < /opt/zoj/current/database/migrations/20260829_add_public_resource_ids.sql
```

迁移目录目前没有自动迁移器，也没有自动记录执行历史。运维人员必须维护已经执行的文件清单。不要循环执行整个 `database/migrations` 目录；部分迁移包含 `ADD COLUMN`，重复执行会失败。

## 9. 安装和配置后端

### 9.1 安装本地编译的 Linux 二进制

再次确认服务器和发布包架构一致：

```bash
uname -m
file /opt/zoj/current/backend/zoj
```

安装二进制：

```bash
sudo install -o root -g zoj -m 0750 \
  /opt/zoj/current/backend/zoj /opt/zoj/backend/zoj
```

如果启动时出现 `Exec format error`，说明本地构建时的 `GOARCH` 与服务器架构不一致，需要重新交叉编译并上传。

### 9.2 创建生产配置

生成两个独立密钥：

```bash
openssl rand -hex 32
openssl rand -base64 48
```

创建 `/opt/zoj/backend/config.yaml`：

```yaml
server:
  # 只监听本机，由 Nginx 对外提供服务。
  port: "127.0.0.1:9090"
  # 有域名时使用 https://oj.example.com；无域名时使用 http://服务器公网IP。
  # base_url: "https://oj.example.com"
  base_url: "http://43.128.30.98"

cors:
  # 必须包含协议，末尾不要加 /。
  allowed_origins:
    # - "https://oj.example.com"
    base_url: "http://43.128.30.98"

db:
  user: "zoj"
  password: "替换为 MySQL 强密码"
  host: "127.0.0.1"
  port: "3306"
  name: "oj_db"
  pool:
    max_open_conns: 50
    max_idle_conns: 10
    conn_max_lifetime_seconds: 3600
    conn_max_idle_time_seconds: 600

redis:
  addr: "127.0.0.1:6379"
  password: "替换为 Redis 强密码"
  db: 0

jwt:
  secret: "a1fbd05d063702c6015a7783697b0205a79e3aa9d2068d881eed7a9630b42c17"
  expire: 24

security:
  # 用于加密远程 OJ 账号、Cookie 等敏感信息。
  # 上线后不能随意更换，否则已有密文无法解密。
  aes_key: "RGQ0COoKKx1PmyI1J9KHerQsJZ/5dRjTZmEcFL9oUVloCGpMf5JDPQhR03NnqUky"

judge:
  multi_instance: false
  urls:
    - "http://127.0.0.1:5050"
  concurrency: 4
  data_dir: "ojdata/problems"

# 可选。未配置时，注册验证码、找回密码和换绑邮箱不可用。
mail:
  host: "smtp.example.com"
  port: 465
  username: "noreply@example.com"
  password: "替换为 SMTP 授权码，不是网页登录密码"
  from_name: "ZOJ"

# 可选。后台远程账号管理通常优先于这里的全局 Cookie。
remote:
  nowcoder:
    cookie: ""
```

无域名时需要改两处：

```yaml
server:
  base_url: "http://43.128.30.98"

cors:
  allowed_origins:
    - "http://43.128.30.98"
```

设置权限并检查目录：

```bash
sudo chown root:zoj /opt/zoj/backend/config.yaml
sudo chmod 0640 /opt/zoj/backend/config.yaml
sudo chown -R zoj:zoj \
  /opt/zoj/backend/uploads \
  /opt/zoj/backend/ojdata \
  /opt/zoj/backend/runtime
```

不要把生产 `config.yaml` 提交到 Git，也不要把它放进 `/var/www/zoj`。

## 10. 安装官方 go-judge 预编译包

go-judge 官方 Release 提供已经编译好的 Linux 二进制，无需安装 Go、无需自己编译，也不需要 Docker。Linux 版本默认把宿主机的 `/lib`、`/lib64`、`/usr` 和 `/bin` 只读挂载进沙箱，因此 C++、Java、Python 编译器需要安装在宿主机。

go-judge 依赖 Linux cgroup。Ubuntu 由 systemd 挂载和管理 cgroup；本教程让 go-judge 以 root 身份运行，以确保 CPU、内存和进程数限制生效，但服务只监听本机 `127.0.0.1:5050`，不能向公网暴露。

### 10.1 安装编译器

当前项目使用：

- `/usr/bin/g++` 编译 C++17。
- `javac`、`jar` 和 `java` 编译、运行 Java。
- `python3` 检查并运行 Python。

安装：

```bash
sudo apt update
sudo apt install -y g++ default-jdk-headless python3
```

确认每个命令都存在：

```bash
/usr/bin/g++ --version
javac -version
jar --version
java -version
python3 --version
```

### 10.2 下载并校验官方二进制

下面固定使用已经核对过的 `v1.12.2`。先确认服务器架构：

```bash
uname -m
```

如果输出 `x86_64`，执行：

```bash
GO_JUDGE_VERSION="1.12.2"
GO_JUDGE_ASSET="go-judge_${GO_JUDGE_VERSION}_linux_amd64v2"

curl -fL \
  "https://github.com/criyle/go-judge/releases/download/v${GO_JUDGE_VERSION}/${GO_JUDGE_ASSET}" \
  -o "/tmp/${GO_JUDGE_ASSET}"
curl -fL \
  "https://github.com/criyle/go-judge/releases/download/v${GO_JUDGE_VERSION}/checksums.txt" \
  -o /tmp/go-judge-checksums.txt

(cd /tmp && grep " ${GO_JUDGE_ASSET}$" go-judge-checksums.txt | sha256sum -c -)
sudo install -o root -g root -m 0755 \
  "/tmp/${GO_JUDGE_ASSET}" /usr/local/bin/go-judge
rm -f "/tmp/${GO_JUDGE_ASSET}" /tmp/go-judge-checksums.txt
```

校验必须显示 `OK`。如果 `grep` 没有输出或 `sha256sum` 失败，不要安装该文件，重新核对版本和架构。

如果 `uname -m` 输出 `aarch64` 或 `arm64`，只把资源名称改为：

```bash
GO_JUDGE_ASSET="go-judge_${GO_JUDGE_VERSION}_linux_arm64"
```

`amd64v3` 只适用于支持更高 x86-64 指令集的 CPU；不确定时使用兼容性更好的 `amd64v2`。

### 10.3 使用 systemd 守护 go-judge

创建 `/etc/systemd/system/go-judge.service`：

```ini
[Unit]
Description=go-judge Sandbox Service
Documentation=https://docs.goj.ac
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/zoj
ExecStart=/usr/local/bin/go-judge \
  -http-addr 127.0.0.1:5050 \
  -parallelism 4 \
  -release
Restart=always
RestartSec=3s
LimitNOFILE=1048576
TasksMax=infinity
Delegate=yes

[Install]
WantedBy=multi-user.target
```

`-parallelism` 是 go-judge 同时运行的沙箱数，建议与后面的 `zoj judge -c` 保持一致：2 核使用 `1`～`2`，4 核使用 `2`～`4`，8 核使用 `4`～`8`。

加载并启动：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now go-judge
sudo systemctl status go-judge --no-pager
curl --fail http://127.0.0.1:5050/version
```

确认监听地址只能是本机：

```bash
sudo ss -lntp | grep 5050
```

正常应看到 `127.0.0.1:5050`，不能是 `0.0.0.0:5050`。查看日志：

```bash
sudo journalctl -u go-judge -n 100 --no-pager
```

### 10.4 升级 go-judge

升级时在官方 Release 页面选择新的稳定版本和相同架构，重复第 10.2 节的下载与校验步骤，然后执行：

```bash
sudo systemctl restart go-judge
curl --fail http://127.0.0.1:5050/version
```

不要把版本变量直接改成未经验证的 `latest`；固定版本、校验 SHA-256 后再替换，出现问题时才能明确回退到旧二进制。

## 11. 使用 systemd 守护后端

### 11.1 HTTP 服务

创建 `/etc/systemd/system/zoj-http.service`：

```ini
[Unit]
Description=ZOJ HTTP API
Wants=network-online.target
After=network-online.target mysql.service redis-server.service

[Service]
Type=simple
User=zoj
Group=zoj
WorkingDirectory=/opt/zoj/backend
Environment=GIN_MODE=release
ExecStart=/opt/zoj/backend/zoj http
Restart=on-failure
RestartSec=3s
TimeoutStopSec=20s
LimitNOFILE=65535
UMask=0027

NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/opt/zoj/backend/uploads
ReadWritePaths=/opt/zoj/backend/ojdata
ReadWritePaths=/opt/zoj/backend/runtime

[Install]
WantedBy=multi-user.target
```

### 11.2 judge worker

创建 `/etc/systemd/system/zoj-judge.service`：

```ini
[Unit]
Description=ZOJ Judge Worker
Wants=network-online.target go-judge.service
After=network-online.target mysql.service redis-server.service go-judge.service

[Service]
Type=simple
User=zoj
Group=zoj
WorkingDirectory=/opt/zoj/backend
ExecStart=/opt/zoj/backend/zoj judge -c 4
Restart=always
RestartSec=5s
TimeoutStopSec=30s
LimitNOFILE=65535
UMask=0027

NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/opt/zoj/backend/ojdata
ReadWritePaths=/opt/zoj/backend/runtime

[Install]
WantedBy=multi-user.target
```

并发建议：

| 服务器 CPU | `-c` 建议值 |
| --- | --- |
| 2 核 | `1`～`2` |
| 4 核 | `2`～`4` |
| 8 核 | `4`～`8` |

不要因为在线人数是 40 就把并发直接设成 40。评测并发主要受 CPU、内存和题目资源限制影响。

加载并启动：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now zoj-http zoj-judge
sudo systemctl status zoj-http zoj-judge --no-pager
```

查看日志：

```bash
sudo journalctl -u zoj-http -f
sudo journalctl -u zoj-judge -f
```

项目还会在 `/opt/zoj/backend/runtime/` 生成按日期命名的日志文件，应定期清理或归档。例如保留 30 天：

```bash
sudo find /opt/zoj/backend/runtime -type f \
  -name 'oj-*.log' -mtime +30 -delete
```

## 12. 发布前端构建产物

前端已经在本地构建。服务器只需要把发布包内的静态文件同步到 Nginx 目录：

```bash
sudo rsync -a --delete /opt/zoj/current/frontend/ /var/www/zoj/
sudo chown -R root:root /var/www/zoj
sudo find /var/www/zoj -type d -exec chmod 0755 {} \;
sudo find /var/www/zoj -type f -exec chmod 0644 {} \;
```

确认首页存在：

```bash
test -f /var/www/zoj/index.html
```

服务器不需要安装 Node.js，也不要在生产环境运行 `npm run dev` 或 `npm run preview`。

## 13. 配置 Nginx

本项目需要：

- SPA 回退：刷新 `/team/...`、`/contest/...` 等前端路由时回到 `index.html`。
- `/api/` 反向代理到后端，并关闭响应缓冲以支持 SSE。
- `/uploads/` 反向代理到后端。
- 至少允许 220 MB 请求体，因为测试数据 ZIP 上限为 200 MB。

### 13.1 有域名

先把域名 A 记录指向服务器公网 IPv4。确认解析：

```bash
dig +short oj.example.com
```

创建 `/etc/nginx/sites-available/zoj`：

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name oj.zeorcoder.top;

    root /var/www/zoj;
    index index.html;

    client_max_body_size 220m;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:9090;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 提交详情和比赛事件使用 SSE。
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    location /uploads/ {
        proxy_pass http://127.0.0.1:9090;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

启用站点：

```bash
sudo ln -sfn /etc/nginx/sites-available/zoj /etc/nginx/sites-enabled/zoj
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

### 13.2 无域名，仅使用公网 IP

创建相同文件，但把开头改成默认站点：

```nginx
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;

    root /var/www/zoj;
    index index.html;
    client_max_body_size 220m;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:9090;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    location /uploads/ {
        proxy_pass http://127.0.0.1:9090;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

无域名时可通过 `http://服务器公网IP` 访问，但登录密码和 Cookie 都是明文传输，不建议长期公网使用。正式使用应尽快绑定域名并启用 HTTPS，或者只允许校园网/VPN 访问。

## 14. 申请并启用免费 SSL 证书

这里配置的是网站 HTTPS 使用的 **SSL/TLS 证书**，与登录服务器使用的 SSH 密钥不是一回事。推荐使用免费的 Let's Encrypt 证书，并由 Certbot 自动配置 Nginx 和续期，不需要先在腾讯云 SSL 控制台购买证书。

### 14.1 申请前检查

必须同时满足：

1. 域名实名认证已通过。
2. 腾讯云 DNS 已添加 `oj` 的 `A` 记录，并指向服务器公网 IPv4。
3. 腾讯云安全组和 UFW 已放行 `80/tcp`、`443/tcp`。
4. Nginx 的 `server_name` 已改成真实域名。
5. `http://oj.example.com` 已经可以从公网访问。

在服务器检查 DNS 和 HTTP：

```bash
dig +short A oj.example.com
curl -I http://oj.example.com
sudo nginx -t
```

`dig` 返回值必须是服务器公网 IP。域名尚未生效、80 端口未开放或 Nginx 配错时，不要继续申请，否则域名验证会失败。

免费证书不能签发给裸 IP，因此使用 `http://服务器公网IP` 部署时跳过本节。

### 14.2 安装 Certbot 并申请证书

Ubuntu 使用 Snap 安装 Certbot：

```bash
sudo systemctl enable --now snapd.socket
sudo snap install core
sudo snap refresh core
sudo snap install --classic certbot
sudo ln -sfn /snap/bin/certbot /usr/local/bin/certbot
```

让 Certbot 申请证书、修改 Nginx，并把 HTTP 自动跳转到 HTTPS：

```bash
sudo certbot --nginx --redirect -d oj.example.com
```

按提示填写用于接收过期和安全通知的邮箱，并同意服务条款。成功后，Certbot 会把证书保存在 `/etc/letsencrypt/` 并修改 Nginx 配置；不要手动移动证书文件，也不要再同时安装腾讯云控制台下载的另一套证书。

### 14.3 验证 HTTPS 和自动续期

```bash
sudo nginx -t
curl -I https://oj.example.com
sudo certbot certificates
sudo certbot renew --dry-run
```

Certbot 的 Snap 安装会配置自动续期。以后不需要定期手动重新申请，但应在系统更新或修改 Nginx 后再次运行 `sudo certbot renew --dry-run`，确认续期链路仍然正常。

最后确认 `/opt/zoj/backend/config.yaml` 使用 HTTPS 地址：

```yaml
server:
  base_url: "https://oj.example.com"
cors:
  allowed_origins:
    - "https://oj.example.com"
```

修改配置后重启后端：

```bash
sudo systemctl restart zoj-http zoj-judge
```

浏览器最终应使用：

```text
https://oj.example.com
```

访问 `http://oj.example.com` 时应自动跳转到 HTTPS。

## 15. 首次启动检查

### 15.1 检查所有服务

```bash
sudo systemctl is-active mysql redis-server go-judge nginx zoj-http zoj-judge
```

### 15.2 检查后端和 Nginx

项目当前没有独立 `/health` 接口，可使用公开语言接口检查：

```bash
curl --fail http://127.0.0.1:9090/api/languages
curl --fail http://127.0.0.1:5050/version
curl --fail http://127.0.0.1/
curl --fail https://oj.example.com/api/languages
```

无域名时将最后一个地址替换为公网 IP。

### 15.3 立即修改初始管理员密码

全新 `schema.sql` 会创建初始管理员：

- 用户名：`admin`
- 密码：`admin123`

在开放网站前必须修改。下面使用 `htpasswd` 生成兼容 bcrypt 的密码哈希：

```bash
read -rsp '新的 admin 密码: ' ADMIN_PASSWORD; echo
ADMIN_HASH=$(htpasswd -bnBC 12 admin "$ADMIN_PASSWORD" | cut -d: -f2)

mysql -h 127.0.0.1 -u zoj -p oj_db <<SQL
UPDATE users
SET password = '${ADMIN_HASH}', email = '你的真实管理员邮箱'
WHERE username = 'admin';
SQL

unset ADMIN_PASSWORD ADMIN_HASH
```

然后使用新密码登录，检查：

1. 题库与团队页面是否能打开。
2. 图片能否上传并通过 `/uploads/...` 访问。
3. 管理后台的 go-judge 状态是否正常。
4. 新建一道简单题并上传成对的 `.in/.out` 测试点。
5. 分别提交 C++、Java 和 Python 程序，确认编译与测试点结果正常。

## 16. 日常更新

每次更新都在本地重新执行第 4 节：测试后端、交叉编译 Linux 二进制、构建前端并生成新的 `zoj-release-linux-amd64.tar.gz`，然后上传。服务器始终不拉源码、不运行 `go build` 或 `npm run build`。

更新前先备份数据库和文件。不要在生产数据库执行 `schema.sql`。

### 16.1 上传并展开新版本

本地 PowerShell：

```powershell
scp E:\fastoj\oj-backend\release\zoj-release-linux-amd64.tar.gz `
  服务器用户@服务器公网IP:/tmp/
```

服务器：

```bash
ZOJ_RELEASE_ID="$(date +%Y%m%d%H%M%S)"
ZOJ_RELEASE_DIR="/opt/zoj/releases/${ZOJ_RELEASE_ID}"

mkdir -p "${ZOJ_RELEASE_DIR}"
tar -xzf /tmp/zoj-release-linux-amd64.tar.gz \
  -C "${ZOJ_RELEASE_DIR}" --strip-components=1

test -x "${ZOJ_RELEASE_DIR}/backend/zoj"
test -f "${ZOJ_RELEASE_DIR}/frontend/index.html"
```

先查看该版本携带的迁移文件，并只执行尚未应用的迁移：

```bash
find "${ZOJ_RELEASE_DIR}/database/migrations" -maxdepth 1 -type f | sort
```

例如：

```bash
mysql -h 127.0.0.1 -u zoj -p oj_db \
  < "${ZOJ_RELEASE_DIR}/database/migrations/20260829_add_public_resource_ids.sql"
```

迁移失败时停止发布，不要安装新二进制。生产迁移前必须先备份数据库。

### 16.2 发布后端和前端

先保留当前版本，再短暂停止后端进程并替换构建产物：

```bash
sudo cp /opt/zoj/backend/zoj /opt/zoj/backend/zoj.previous
sudo install -d -o root -g root -m 0755 /var/www/zoj.previous
sudo rsync -a --delete /var/www/zoj/ /var/www/zoj.previous/

sudo systemctl stop zoj-http zoj-judge
sudo install -o root -g zoj -m 0750 \
  "${ZOJ_RELEASE_DIR}/backend/zoj" /opt/zoj/backend/zoj
sudo rsync -a --delete "${ZOJ_RELEASE_DIR}/frontend/" /var/www/zoj/
sudo chown -R root:root /var/www/zoj
sudo ln -sfn "${ZOJ_RELEASE_DIR}" /opt/zoj/current

sudo nginx -t
sudo systemctl start zoj-http zoj-judge
sudo systemctl reload nginx
sudo systemctl status zoj-http zoj-judge --no-pager
```

发布后立即检查：

```bash
curl --fail http://127.0.0.1:9090/api/languages
curl --fail https://oj.example.com/api/languages
```

确认成功后删除临时上传文件。`/opt/zoj/releases` 建议保留最近 3～5 个发布目录，确认不再需要时再手工清理明确的旧版本目录。

```bash
rm -f /tmp/zoj-release-linux-amd64.tar.gz
```

### 16.3 回滚

后端二进制回滚：

```bash
sudo systemctl stop zoj-http zoj-judge
sudo cp /opt/zoj/backend/zoj.previous /opt/zoj/backend/zoj
sudo systemctl start zoj-http zoj-judge
```

前端回滚：

```bash
sudo rsync -a --delete /var/www/zoj.previous/ /var/www/zoj/
sudo systemctl reload nginx
```

数据库迁移不保证可以自动回滚。涉及删列、改类型或数据回填的版本，必须根据迁移内容制定单独回滚方案。

## 17. 备份与恢复

至少备份：

- MySQL 数据库。
- `/opt/zoj/backend/ojdata`：题目测试数据。
- `/opt/zoj/backend/uploads`：用户上传图片和团队封面。
- `/opt/zoj/backend/config.yaml`：包含密钥，只能加密保存并严格限制权限。

### 17.1 手工备份

```bash
BACKUP_DIR="/var/backups/zoj/$(date +%F-%H%M%S)"
sudo install -d -m 0700 "$BACKUP_DIR"

mysqldump -h 127.0.0.1 -u zoj -p \
  --single-transaction --routines --triggers oj_db \
  | sudo tee "$BACKUP_DIR/oj_db.sql" >/dev/null

sudo tar -C /opt/zoj/backend -czf "$BACKUP_DIR/files.tar.gz" \
  ojdata uploads
sudo install -m 0600 /opt/zoj/backend/config.yaml \
  "$BACKUP_DIR/config.yaml"
```

备份完成后把文件复制到另一台机器或对象存储。同一块云硬盘上的备份不能应对磁盘损坏、误删整机或账号问题。

### 17.2 恢复

```bash
sudo systemctl stop zoj-http zoj-judge

mysql -h 127.0.0.1 -u zoj -p oj_db < /备份路径/oj_db.sql
sudo tar -C /opt/zoj/backend -xzf /备份路径/files.tar.gz
sudo chown -R zoj:zoj /opt/zoj/backend/ojdata /opt/zoj/backend/uploads

sudo systemctl start zoj-http zoj-judge
```

恢复后至少执行一次 C++ 测试提交，不能只检查首页是否能打开。

## 18. 常见问题

### 18.1 Nginx 返回 502

```bash
sudo systemctl status zoj-http --no-pager
sudo journalctl -u zoj-http -n 200 --no-pager
curl -v http://127.0.0.1:9090/api/languages
```

重点检查 `config.yaml`、MySQL、Redis，以及 `server.port` 是否为 `127.0.0.1:9090`。

### 18.2 刷新前端子页面后 404

确认 Nginx 的 `location /` 中存在：

```nginx
try_files $uri $uri/ /index.html;
```

### 18.3 上传测试数据返回 413

确认 Nginx 配置包含：

```nginx
client_max_body_size 220m;
```

修改后执行 `sudo nginx -t && sudo systemctl reload nginx`。

### 18.4 图片上传成功但无法显示

检查：

- `server.base_url` 是否与浏览器访问地址完全一致。
- Nginx 是否代理 `/uploads/`。
- `/opt/zoj/backend/uploads` 是否属于 `zoj:zoj`。
- 从 HTTP 切换到 HTTPS 后是否仍返回旧的 `http://` 图片地址。

### 18.5 登录后马上变成未登录

Redis 保存登录 session。检查：

```bash
sudo systemctl status redis-server --no-pager
export REDISCLI_AUTH='Redis 密码'
redis-cli ping
redis-cli DBSIZE
unset REDISCLI_AUTH
```

前后端应使用同一域名，通过 Nginx 的同源 `/api` 访问。`cors.allowed_origins` 必须包含准确的协议和域名，末尾不能多一个 `/`。

### 18.6 提交一直 Pending

依次检查：

```bash
sudo systemctl status zoj-judge --no-pager
sudo journalctl -u zoj-judge -n 200 --no-pager
sudo systemctl status go-judge --no-pager
sudo journalctl -u go-judge -n 200 --no-pager
curl --fail http://127.0.0.1:5050/version
```

还要确认题目已经上传成对的 `.in/.out` 测试点，并且 `judge.data_dir` 指向 `ojdata/problems`。

### 18.7 C++、Java 或 Python 提示找不到命令

说明宿主机没有安装相应编译器，或者命令路径不正确。检查：

```bash
/usr/bin/g++ --version
javac -version
jar --version
java -version
python3 --version
```

缺少命令时重新执行：

```bash
sudo apt install -y g++ default-jdk-headless python3
sudo systemctl restart go-judge zoj-judge
```

### 18.8 SSE 评测进度延迟或一次性出现

确认 `/api/` 代理关闭缓冲并延长读取超时：

```nginx
proxy_buffering off;
proxy_cache off;
proxy_read_timeout 3600s;
```

### 18.9 后端提示目录没有写权限

```bash
sudo chown -R zoj:zoj \
  /opt/zoj/backend/uploads \
  /opt/zoj/backend/ojdata \
  /opt/zoj/backend/runtime
sudo systemctl restart zoj-http zoj-judge
```

### 18.10 数据库迁移提示 Duplicate column

通常表示迁移被重复执行。不要继续尝试整个迁移目录。先查看表结构和迁移文件，确认该文件是否已经执行：

```bash
mysql -h 127.0.0.1 -u zoj -p -D oj_db -e "SHOW CREATE TABLE teams\G"
```

生产环境修复迁移前必须先备份。

## 19. 上线检查清单

- [ ] 安全组只开放 SSH、80、443。
- [ ] MySQL、Redis、后端和 go-judge 都只监听本机。
- [ ] MySQL、Redis、JWT、AES 使用四组不同的强密码或密钥。
- [ ] `config.yaml` 权限为 `0640`，且未提交到 Git。
- [ ] 已修改 `admin/admin123` 默认密码和默认邮箱。
- [ ] 已启用 HTTPS，并通过 `certbot renew --dry-run`。
- [ ] `zoj-http`、`zoj-judge`、go-judge、MySQL、Redis、Nginx 已设置开机启动。
- [ ] C++、Java、Python 各完成一次真实评测。
- [ ] 测试过图片上传、测试数据上传、Excel 下载和 SSE 评测进度。
- [ ] 已做一次数据库与文件备份，并验证备份可以读取。
- [ ] 已记录执行过的数据库迁移文件。

## 20. 官方参考资料

- [Go：Linux 下载与安装](https://go.dev/doc/install)
- [Node.js 发布周期与 LTS](https://nodejs.org/en/about/previous-releases)
- [MySQL 8.4 APT 仓库安装](https://dev.mysql.com/doc/refman/8.4/en/linux-installation-apt-repo.html)
- [Redis Linux 安装](https://redis.io/docs/latest/operate/oss_and_stack/install/install-stack/)
- [go-judge 安装](https://docs.goj.ac/install)
- [go-judge GitHub Releases](https://github.com/criyle/go-judge/releases)
- [Nginx Linux 软件包](https://nginx.org/en/linux_packages.html)
- [Certbot + Nginx](https://certbot.eff.org/instructions?ws=nginx&os=snap)
