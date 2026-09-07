package cache

import "fmt"

func ContestRank(contestID int64) string {
	return fmt.Sprintf("contest:%d:rank", contestID)
}

func ContestProblemSubmittedUsers(contestID, problemID int64) string {
	return fmt.Sprintf("contest:%d:problem:%d:submitted_users", contestID, problemID)
}

func ContestProblemAcceptedUsers(contestID, problemID int64) string {
	return fmt.Sprintf("contest:%d:problem:%d:accepted_users", contestID, problemID)
}

func ContestProblemStats(contestID, problemID int64) string {
	return fmt.Sprintf("contest:%d:problem:%d:stats", contestID, problemID)
}

func ContestUserAcceptedProblems(contestID, userID int64) string {
	return fmt.Sprintf("contest:%d:user:%d:accepted_problems", contestID, userID)
}

func ContestUserStats(contestID, userID int64) string {
	return fmt.Sprintf("contest:%d:user:%d:stats", contestID, userID)
}

func ContestUserProblem(contestID, userID, problemID int64) string {
	return fmt.Sprintf("contest:%d:user:%d:problem:%d", contestID, userID, problemID)
}

func SubmitLock(userID int64) string {
	return fmt.Sprintf("submit:%d:lock", userID)
}

func UserRankGen() string {
	return "user:rank:gen"
}

func UserRank(gen string, page, pageSize int) string {
	return fmt.Sprintf("user:rank:%s:%d:%d", gen, page, pageSize)
}

func ProblemDetail(problemID int64) string {
	return fmt.Sprintf("problem:%d:detail", problemID)
}

func ProblemTags() string {
	return "problem:tags"
}
