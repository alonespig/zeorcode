package judge

import (
	"zoj/internal/repository"
	judgeapi "zoj/pkg/judge"
	"zoj/pkg/util"
)

// judgeOutput 用清单哈希判定用户输出：AC / PE / WA
func judgeOutput(stdout, strippedMd5, allStrippedMd5 string) int {
	if util.MD5(repository.RtrimOutput(stdout)) == strippedMd5 {
		return judgeapi.Accepted
	}
	if util.MD5(repository.StripAllSpace(stdout)) == allStrippedMd5 {
		return judgeapi.PresentationError
	}
	return judgeapi.WrongAnswer
}
