package handler

import "github.com/google/wire"

var HandlerSet = wire.NewSet(
	NewContestController,
	NewProblemController,
	NewSubmissionController,
	NewPostController,
	NewUserController,
	NewUploadController,
	NewAdminController,
	NewRemoteController,
	NewNotificationController,
	NewProblemSetController,
	NewTeamController,
	NewHomeworkController,
	NewAgentController,
	NewLanguageController,
)
