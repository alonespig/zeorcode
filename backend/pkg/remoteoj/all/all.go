// Package all 空导入所有 OJ 实现，触发其 init() 自注册。
// 新增 OJ 时在此加一行 import 即可。
package all

import (
	_ "zoj/pkg/remoteoj/atcoder"
	_ "zoj/pkg/remoteoj/hdu"
	_ "zoj/pkg/remoteoj/loj"
	_ "zoj/pkg/remoteoj/nowcoder"
)
