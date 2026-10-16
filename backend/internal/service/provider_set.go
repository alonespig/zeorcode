package service

import "github.com/google/wire"

var ServerSet = wire.NewSet(
	NewContestService,
	NewProblemService,
	NewSubmissionService,
	NewCodeRunService,
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
	NewDashboardService,
	NewAuditService,
)
