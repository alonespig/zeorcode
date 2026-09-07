package service

import "github.com/google/wire"

var ServerSet = wire.NewSet(
	NewContestService,
	NewProblemService,
	NewSubmissionService,
	NewPostService,
	NewUserService,
	NewRemoteService,
	NewNotificationService,
	NewVerifyService,
	NewCaptchaService,
	NewProblemSetService,
	NewTeamService,
	NewHomeworkService,
	NewAgentService,
	NewLanguageService,
)
