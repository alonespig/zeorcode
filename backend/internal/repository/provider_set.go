package repository

import (
	"github.com/google/wire"
)

var RepoSet = wire.NewSet(
	NewUserRepo,
	NewContestRepo,
	NewProblemRepo,
	NewSubmissionRepo,
	NewPostRepo,
	NewRemoteAccountRepo,
	NewNotificationRepo,
	NewProblemSetRepo,
	NewTeamRepo,
	NewHomeworkRepo,
	NewAgentRepo,
	NewLanguageRepo,
)
