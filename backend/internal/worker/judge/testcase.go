package judge

import (
	judgeapi "zoj/pkg/judge"
)

// judgeOutput 用清单哈希判定用户输出：AC / PE / WA
func judgeOutput(stdout, strippedMd5, allStrippedMd5 string) int {
	if judgeapi.MD5(judgeapi.RtrimOutput(stdout)) == strippedMd5 {
		return judgeapi.Accepted
	}
	if judgeapi.MD5(judgeapi.StripAllSpace(stdout)) == allStrippedMd5 {
		return judgeapi.PresentationError
	}
	return judgeapi.WrongAnswer
}
