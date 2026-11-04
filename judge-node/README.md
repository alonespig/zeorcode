# ZeorCode Judge Node

独立的分布式评测节点。它通过 gRPC 接收中心 Judge Worker 的任务，按版本缓存题目输入数据，并调用同机的 go-judge 完成编译和运行。

该模块不连接 ZeorCode 的 MySQL 或 Redis，也不导入 `backend` 模块。它只依赖 gRPC 协议、中心题目下载接口和本机 go-judge。

```bash
go test ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/zeorcode-judge-node ./cmd/judge-node
```

运行配置见 `../ops/systemd/zeorcode-judge-node.env.example`。
