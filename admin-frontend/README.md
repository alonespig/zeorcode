# ZeorCode 管理后台

ZeorCode 的独立管理端，基于 Vue 3、TypeScript、Vite、Element Plus 与 Pure Admin Thin 构建。

## 本地开发

```bash
pnpm install
pnpm dev
```

默认访问地址为 `http://localhost:8848`，开发服务器会将 `/api` 请求代理到 `http://localhost:1001`。

## 验证

```bash
pnpm typecheck
pnpm build
```

登录态由后端写入 HttpOnly Cookie。管理端不会在浏览器存储 JWT，启动时通过 `/api/session` 校验管理员身份。

本项目基于 Pure Admin Thin 修改，原项目的 MIT 许可证见 [LICENSE](./LICENSE)。
