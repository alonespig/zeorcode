# GitHub Actions 自动部署

当代码推送到 `main` 后，`deploy.yml` 会测试后端、按目标架构构建 Linux 后端和 Vue 前端，并通过 SSH 更新生产服务器。生产服务器必须先完成首次安装，确保以下服务和路径已经存在：

- `zoj-http.service`
- `zoj-judge.service`
- Nginx 和 `/var/www/zoj`
- `/opt/zoj/backend/zoj`
- `/opt/zoj/releases`

## 1. 创建专用 SSH 密钥

在可信电脑上执行：

```bash
ssh-keygen -t ed25519 -C "zeorcode-github-actions" -f zeorcode-github-actions
```

把 `zeorcode-github-actions.pub` 的内容追加到服务器部署用户的 `~/.ssh/authorized_keys`。私钥 `zeorcode-github-actions` 的完整内容稍后保存为 GitHub Secret，不要提交到仓库。

## 2. 安装服务器发布脚本

把仓库中的 `ops/zeorcode-deploy` 复制到服务器，然后执行：

```bash
sudo install -o root -g root -m 0755 zeorcode-deploy /usr/local/sbin/zeorcode-deploy
```

假设 SSH 部署用户叫 `deploy`，使用 `sudo visudo -f /etc/sudoers.d/zeorcode-deploy` 创建：

```sudoers
deploy ALL=(root) NOPASSWD: /usr/local/sbin/zeorcode-deploy
```

将 `deploy` 换成实际用户名，然后验证：

```bash
sudo visudo -cf /etc/sudoers.d/zeorcode-deploy
```

## 3. 配置 GitHub Secrets

进入 GitHub 仓库的 **Settings → Secrets and variables → Actions → New repository secret**，添加：

| Secret | 内容 |
| --- | --- |
| `DEPLOY_HOST` | 服务器域名或公网 IP |
| `DEPLOY_USER` | SSH 部署用户名 |
| `DEPLOY_PORT` | SSH 端口；使用 22 时可以不填 |
| `DEPLOY_SSH_KEY` | 第 1 步生成的私钥完整内容 |
| `DEPLOY_KNOWN_HOSTS` | 服务器 SSH 主机公钥指纹 |

工作流默认构建 AMD64 后端。先在服务器执行 `uname -m`：返回 `x86_64` 时无需修改；返回 `aarch64` 或 `arm64` 时，在 **Settings → Secrets and variables → Actions → Variables** 中添加仓库变量 `DEPLOY_GOARCH=arm64`。

在可信网络中生成 `DEPLOY_KNOWN_HOSTS`。默认 22 端口：

```bash
ssh-keyscan -H your-server.example.com
```

非默认端口：

```bash
ssh-keyscan -p 2222 -H your-server.example.com
```

添加前，应在云服务器控制台或首次人工 SSH 连接时核对主机指纹，避免保存错误的服务器公钥。

## 4. 首次运行

进入 GitHub 仓库的 **Actions → Build and deploy production → Run workflow** 手动运行一次。成功后，每次推送 `main` 都会自动发布：

```bash
git push origin main
```

如果 `backend/database` 有变化，自动发布会停止。先备份数据库并在服务器手动执行、验证相应迁移，再从 Actions 页面手动运行工作流。手动运行用于确认迁移已经完成，不会再次拦截数据库变更。

发布失败时，服务器脚本会恢复上一个后端二进制和前端文件。生产配置、`uploads`、`ojdata` 和 `runtime` 不会被发布流程覆盖。
