// Package all 空导入所有 OJ 实现，触发其 init() 自注册。
// 新增 OJ 时在此加一行 import 即可。
package all

import (
	_ "github.com/alonespig/zeorcode/backend/pkg/remoteoj/atcoder"
	_ "github.com/alonespig/zeorcode/backend/pkg/remoteoj/hdu"
	_ "github.com/alonespig/zeorcode/backend/pkg/remoteoj/loj"
	_ "github.com/alonespig/zeorcode/backend/pkg/remoteoj/nowcoder"
)
