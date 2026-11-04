# Judge gRPC protocol

`judge.proto` 是中心后端与独立 judge-node 的唯一协议源。生成文件分别写入两个 Go Module，运行时不需要共享 Go 包。

在仓库根目录执行：

```bash
protoc -I . --go_out=backend --go_opt=module=github.com/alonespig/zeorcode/backend --go_opt=Mproto/judge/v1/judge.proto=github.com/alonespig/zeorcode/backend/internal/judgeproto --go-grpc_out=backend --go-grpc_opt=module=github.com/alonespig/zeorcode/backend --go-grpc_opt=Mproto/judge/v1/judge.proto=github.com/alonespig/zeorcode/backend/internal/judgeproto proto/judge/v1/judge.proto

protoc -I . --go_out=judge-node --go_opt=module=github.com/alonespig/zeorcode/judge-node --go_opt=Mproto/judge/v1/judge.proto=github.com/alonespig/zeorcode/judge-node/internal/judgeproto --go-grpc_out=judge-node --go-grpc_opt=module=github.com/alonespig/zeorcode/judge-node --go-grpc_opt=Mproto/judge/v1/judge.proto=github.com/alonespig/zeorcode/judge-node/internal/judgeproto proto/judge/v1/judge.proto
```
