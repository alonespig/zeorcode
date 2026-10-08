package router

import (
	"time"

	"zoj/internal/http/handler"
	"zoj/internal/http/middleware"
	"zoj/internal/infra/cache"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	problemController      *handler.ProblemController
	contestController      *handler.ContestController
	submissionController   *handler.SubmissionController
	postController         *handler.PostController
	userController         *handler.UserController
	uploadController       *handler.UploadController
	adminController        *handler.AdminController
	remoteController       *handler.RemoteController
	notificationController *handler.NotificationController
	problemSetController   *handler.ProblemSetController
	teamController         *handler.TeamController
	homeworkController     *handler.HomeworkController
	agentController        *handler.AgentController
	languageController     *handler.LanguageController
	auth                   *service.TokenService
	cache                  *cache.Cache
	audit                  *service.AuditService
}

func NewHttpServer(problemController *handler.ProblemController,
	contestController *handler.ContestController,
	submissionController *handler.SubmissionController,
	postController *handler.PostController,
	userController *handler.UserController,
	uploadController *handler.UploadController,
	adminController *handler.AdminController,
	remoteController *handler.RemoteController,
	notificationController *handler.NotificationController,
	problemSetController *handler.ProblemSetController,
	teamController *handler.TeamController,
	homeworkController *handler.HomeworkController,
	agentController *handler.AgentController,
	languageController *handler.LanguageController,
	auth *service.TokenService,
	c *cache.Cache,
	audit *service.AuditService) *HttpServer {
	return &HttpServer{
		problemController:      problemController,
		contestController:      contestController,
		submissionController:   submissionController,
		postController:         postController,
		userController:         userController,
		uploadController:       uploadController,
		adminController:        adminController,
		remoteController:       remoteController,
		notificationController: notificationController,
		problemSetController:   problemSetController,
		teamController:         teamController,
		homeworkController:     homeworkController,
		agentController:        agentController,
		languageController:     languageController,
		auth:                   auth,
		cache:                  c,
		audit:                  audit,
	}
}

func (h *HttpServer) Register(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())
	// 全局按 IP 限流：每 IP 每分钟最多 600 次，兜底防刷（Redis 挂了自动放行）
	r.Use(middleware.RateLimit(h.cache, "global", 600, time.Minute))
	r.Static("/uploads", "./uploads")

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code": 404,
			"msg":  "API not found",
			"data": nil,
		})
	})

	h.initUserRouter(r)
	h.initLanguageRouter(r)
	h.initProblemRouter(r)
	h.initProblemSetRouter(r)
	h.initTeamRouter(r)
	h.initHomeworkRouter(r)
	h.initContestRouter(r)
	h.initSubmissionRouter(r)
	h.initPostRouter(r)
	h.initNotificationRouter(r)
	h.initUploadRouter(r)
	h.initRemoteRouter(r)
	h.initAgentRouter(r)
	h.initAdminRouter(r)
}
