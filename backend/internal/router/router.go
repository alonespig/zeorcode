package router

import (
	"time"

	"zoj/internal/common/response"
	"zoj/internal/handler"
	"zoj/internal/infra/cache"
	"zoj/internal/middleware"

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
	auth                   *middleware.Auth
	cache                  *cache.Cache
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
	auth *middleware.Auth,
	c *cache.Cache) *HttpServer {
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
	h.initContestRouter(r)
	h.initSubmissionRouter(r)
	h.initPostRouter(r)
	h.initUploadRouter(r)
	h.initAdminRouter(r)
	h.initRemoteRouter(r)
	h.initNotificationRouter(r)
	h.initAgentRouter(r)
}

// Agent is a standalone administrator workspace. It deliberately does not use
// /api/admin so its API can remain resource-oriented while sharing the same
// site-admin authorization boundary.
func (h *HttpServer) initAgentRouter(r *gin.Engine) {
	agent := r.Group("/api/agent", h.auth.JWTAuth(), middleware.AdminRequired(),
		middleware.RateLimit(h.cache, "agent", 120, time.Hour))
	{
		agent.GET("/conversations", response.Wrap(h.agentController.ListConversations))
		agent.POST("/conversations", response.Wrap(h.agentController.CreateConversation))
		agent.GET("/conversations/:id", response.Wrap(h.agentController.ConversationDetail))
		agent.PATCH("/conversations/:id", response.Wrap(h.agentController.UpdateConversation))
		agent.POST("/conversations/:id/turns", h.agentController.Turn)
	}
}

// ==================== Notification(站内通知) ====================

func (h *HttpServer) initNotificationRouter(r *gin.Engine) {
	auth := r.Group("/api", h.auth.JWTAuth())
	{
		auth.GET("/notifications", response.Wrap(h.notificationController.List))
		auth.GET("/notifications/unread-count", response.Wrap(h.notificationController.UnreadCount))
		auth.POST("/notifications/read", response.Wrap(h.notificationController.MarkRead))
	}
	admin := r.Group("/api/admin", h.auth.JWTAuth(), middleware.AdminRequired())
	{
		admin.POST("/notifications", response.Wrap(h.notificationController.Broadcast))
	}
}

func (h *HttpServer) initLanguageRouter(r *gin.Engine) {
	pub := r.Group("/api")
	{
		pub.GET("/languages", response.Wrap(h.languageController.List))
	}
}

// ==================== Remote(远程拉题) ====================

func (h *HttpServer) initRemoteRouter(r *gin.Engine) {
	admin := r.Group("/api/remote", h.auth.JWTAuth(), middleware.AdminRequired())
	{
		admin.GET("/oj", response.Wrap(h.remoteController.ListOJ))
		admin.GET("/problem", response.Wrap(h.remoteController.GetRemoteProblem))
	}
}

// ==================== User ====================

func (h *HttpServer) initUserRouter(r *gin.Engine) {
	user := r.Group("/api")
	{
		user.POST("/user", response.Wrap(h.userController.CreateUser))
		user.GET("/captcha", middleware.RateLimit(h.cache, "captcha", 30, time.Minute), response.Wrap(h.userController.GenerateCaptcha))
		user.POST("/login", middleware.RateLimit(h.cache, "login", 10, time.Minute), h.userController.Login)
		user.GET("/session", h.auth.JWTAuthOptional(), response.Wrap(h.userController.Session))
		// 即使 token 已失效也允许退出，以确保浏览器中的 HttpOnly Cookie 能被清除。
		user.POST("/logout", h.auth.JWTAuthOptional(), h.userController.Logout)
		// 邮箱验证码：发码 + 找回密码按 IP 限流；换绑需登录
		user.POST("/verify-code", middleware.RateLimit(h.cache, "verifycode", 20, time.Hour), response.Wrap(h.userController.SendVerifyCode))
		user.POST("/reset-password", middleware.RateLimit(h.cache, "resetpwd", 20, time.Hour), response.Wrap(h.userController.ResetPassword))
		user.PUT("/user/email", h.auth.JWTAuth(), response.Wrap(h.userController.BindEmail))
		user.PUT("/user/password", h.auth.JWTAuth(), middleware.RateLimit(h.cache, "change-password", 10, time.Hour), h.userController.ChangePassword)
		user.GET("/user/rank", response.Wrap(h.userController.Rank))
		user.GET("/user/rating-rank", response.Wrap(h.userController.RatingRank))
		user.GET("/user/:id/profile", response.Wrap(h.userController.UserProfile))
		user.GET("/user/info", h.auth.JWTAuth(), response.Wrap(h.userController.GetUserInfo))
		user.PUT("/user/info", h.auth.JWTAuth(), response.Wrap(h.userController.UpdateUserInfo))
		user.GET("/user/:id/ac-stats", response.Wrap(h.userController.GetUserRecent7DaysPassCount))
		user.GET("/user/:id/rating", response.Wrap(h.userController.RatingHistory))
		user.GET("/user/:id/contest-history", response.Wrap(h.userController.ContestHistory))
	}
}

// ==================== Problem ====================

func (h *HttpServer) initTeamRouter(r *gin.Engine) {
	// 列表与详情用可选登录：未登录也能浏览团队，只是看不到「我的角色」和内部内容
	pub := r.Group("/api")
	{
		pub.GET("/team", h.auth.JWTAuthOptional(), response.Wrap(h.teamController.List))
		pub.GET("/team/:id", h.auth.JWTAuthOptional(), response.Wrap(h.teamController.Detail))
		pub.GET("/team/:id/member", h.auth.JWTAuthOptional(), response.Wrap(h.teamController.ListMembers))
	}

	// 写操作一律要登录；创建团队额外要求站点管理员，其他操作仍由 service 层
	// 按团队角色判定，不能统一使用全站 AdminRequired。
	auth := r.Group("/api", h.auth.JWTAuth())
	{
		auth.POST("/team", middleware.AdminRequired(), response.Wrap(h.teamController.Create))
		auth.PUT("/team/:id", response.Wrap(h.teamController.Update))
		auth.DELETE("/team/:id", response.Wrap(h.teamController.Delete))
		auth.POST("/team/:id/join", response.Wrap(h.teamController.Join))
		auth.POST("/team/:id/quit", response.Wrap(h.teamController.Quit))
		auth.GET("/team/:id/member/import/template", h.teamController.DownloadStudentImportTemplate)
		auth.POST("/team/:id/member/import",
			middleware.RateLimit(h.cache, "team-student-import", 10, time.Hour),
			response.Wrap(h.teamController.ImportStudents))
		auth.PUT("/team/:id/member/:uid/role", response.Wrap(h.teamController.SetMemberRole))
		auth.DELETE("/team/:id/member/:uid", response.Wrap(h.teamController.RemoveMember))
	}

	// 作业读接口用可选登录：service 里非团队成员一律按「作业不存在」处理，
	// 不泄露作业的存在，因此这里不必强制登录
	{
		pub.GET("/team/:id/homework", h.auth.JWTAuthOptional(), response.Wrap(h.homeworkController.ListByTeam))
		pub.GET("/homework/:hid", h.auth.JWTAuthOptional(), response.Wrap(h.homeworkController.Detail))
		pub.GET("/homework/:hid/problem/:problemID", h.auth.JWTAuthOptional(), response.Wrap(h.homeworkController.Problem))
		pub.GET("/homework/:hid/rank", h.auth.JWTAuthOptional(), response.Wrap(h.homeworkController.Rank))
		pub.GET("/homework/:hid/rank/export", h.auth.JWTAuthOptional(), h.homeworkController.ExportRank)
		pub.GET("/homework/:hid/submission", h.auth.JWTAuthOptional(), response.Wrap(h.homeworkController.Submissions))
		pub.GET("/homework/:hid/submission/:sid", h.auth.JWTAuthOptional(), response.Wrap(h.homeworkController.SubmissionDetail))
	}
	{
		auth.POST("/team/:id/homework", response.Wrap(h.homeworkController.Create))
		auth.PUT("/homework/:hid", response.Wrap(h.homeworkController.Update))
		auth.DELETE("/homework/:hid", response.Wrap(h.homeworkController.Delete))
		auth.POST("/homework/:hid/submit", response.Wrap(h.homeworkController.Submit))
	}
}

func (h *HttpServer) initProblemSetRouter(r *gin.Engine) {
	// 前台：列表和详情都用可选登录 —— 未登录也能看，只是不返回做题进度
	pub := r.Group("/api")
	{
		pub.GET("/problemset", h.auth.JWTAuthOptional(), response.Wrap(h.problemSetController.List))
		pub.GET("/problemset/:id", h.auth.JWTAuthOptional(), response.Wrap(h.problemSetController.Detail))
		// 解锁要记到当前用户名下，必须登录
		pub.POST("/problemset/:id/unlock", h.auth.JWTAuth(),
			middleware.RateLimit(h.cache, "problemset-unlock", 10, time.Minute),
			response.Wrap(h.problemSetController.Unlock))
	}

	// 后台：仅管理员，列表含草稿
	admin := r.Group("/api/admin", h.auth.JWTAuth(), middleware.AdminRequired())
	{
		admin.GET("/problemset", response.Wrap(h.problemSetController.AdminList))
		admin.POST("/problemset", response.Wrap(h.problemSetController.AdminCreate))
		admin.PUT("/problemset/:id", response.Wrap(h.problemSetController.AdminUpdate))
		admin.DELETE("/problemset/:id", response.Wrap(h.problemSetController.AdminDelete))
	}
}

func (h *HttpServer) initProblemRouter(r *gin.Engine) {
	// 公开接口
	pub := r.Group("/api")
	{
		pub.GET("/tags", response.Wrap(h.problemController.GetTagList))
		pub.GET("/problems", h.auth.JWTAuthOptional(), response.Wrap(h.problemController.List))
		pub.GET("/problems/:id", h.auth.JWTAuthOptional(), response.Wrap(h.problemController.GetProblemDetail))
	}

	// 管理员接口
	admin := r.Group("/api", h.auth.JWTAuth(), middleware.AdminRequired())
	{
		admin.POST("/problems", response.Wrap(h.problemController.CreateProblem))
		admin.PUT("/problems/:id", response.Wrap(h.problemController.UpdateProblem))
		admin.GET("/problems/:id/edit", response.Wrap(h.problemController.GetProblemForEdit))
		admin.GET("/problems/:id/testdata", response.Wrap(h.problemController.ListFiles))
		admin.GET("/problems/:id/testdata/content", response.Wrap(h.problemController.GetFileContent))
		admin.POST("/problems/:id/testdata", response.Wrap(h.problemController.UploadFile))
		admin.POST("/problems/:id/testdata/zip", response.Wrap(h.problemController.UploadTestcaseZip))
		admin.DELETE("/problems/:id/testdata", response.Wrap(h.problemController.DeleteFiles))
		admin.GET("/problems/:id/testdata/download", h.problemController.DownloadFiles)
	}

	tagAdmin := r.Group("/api/admin/tags", h.auth.JWTAuth(), middleware.AdminRequired())
	{
		tagAdmin.GET("", response.Wrap(h.problemController.GetAdminTagList))
		tagAdmin.POST("", response.Wrap(h.problemController.CreateTag))
		tagAdmin.PUT("/:id", response.Wrap(h.problemController.UpdateTag))
		tagAdmin.DELETE("/:id", response.Wrap(h.problemController.DeleteTag))
	}
}

// ==================== Contest ====================

func (h *HttpServer) initContestRouter(r *gin.Engine) {
	// 公开浏览：可匿名（列表 / 简介 / 排名 / 倒计时）。登录后自动带上 isRegistered / isSelf
	contestPub := r.Group("/api/contest", h.auth.JWTAuthOptional())
	{
		contestPub.GET("", response.Wrap(h.contestController.List))
		contestPub.GET("/:id", response.Wrap(h.contestController.GetContestDetail))
		contestPub.GET("/:id/desc", response.Wrap(h.contestController.GetContestDesc))
		contestPub.GET("/:id/rank", response.Wrap(h.contestController.GetContestRank))
		contestPub.GET("/:id/myrank", response.Wrap(h.contestController.GetMyContestRank))
		contestPub.GET("/:id/sse", h.contestController.SSEventStream)
	}

	// 需登录：报名 / 做题 / 提交 / 我的提交
	contest := r.Group("/api/contest", h.auth.JWTAuth())
	{
		contest.POST("/join", response.Wrap(h.contestController.JoinContest))
		contest.GET("/:id/problem", response.Wrap(h.contestController.GetContestProblemList))
		contest.GET("/:id/problem/:problemID", response.Wrap(h.contestController.GetContestProblem))
		contest.POST("/submit", response.Wrap(h.contestController.Submit))
		contest.GET("/:id/submission", response.Wrap(h.contestController.GetContestSubmissions))
		contest.GET("/:id/submit-info", response.Wrap(h.contestController.GetContestSubmitInfo))
	}

	// 管理员接口
	contestAdmin := r.Group("/api/contest", h.auth.JWTAuth(), middleware.AdminRequired())
	{
		contestAdmin.POST("", response.Wrap(h.contestController.Create))
		contestAdmin.GET("/:id/info", response.Wrap(h.contestController.EditContest))
		contestAdmin.PUT("/:id", response.Wrap(h.contestController.Update))
	}
}

// ==================== Submission ====================

func (h *HttpServer) initSubmissionRouter(r *gin.Engine) {
	api := r.Group("/api/submission")
	{
		api.POST("", h.auth.JWTAuth(),
			response.Wrap(h.submissionController.Create))

		api.GET("", h.auth.JWTAuthOptional(),
			response.Wrap(h.submissionController.GetSubmissionList))
		api.GET("/:id", h.auth.JWTAuthOptional(),
			response.Wrap(h.submissionController.GetSubmissionByID))
		api.GET("/:id/stream", h.auth.JWTAuthOptional(),
			h.submissionController.GetSubmissionStream)
	}
}

// ==================== Post(博客 / 讨论 / 题解) ====================

func (h *HttpServer) initPostRouter(r *gin.Engine) {
	pub := r.Group("/api")
	{
		pub.GET("/posts", h.auth.JWTAuthOptional(), response.Wrap(h.postController.ListPosts))
		pub.GET("/posts/:id", h.auth.JWTAuthOptional(), response.Wrap(h.postController.GetPost))
		pub.GET("/posts/:id/comments", h.auth.JWTAuthOptional(), response.Wrap(h.postController.ListComments))
	}

	auth := r.Group("/api", h.auth.JWTAuth())
	{
		auth.POST("/posts", response.Wrap(h.postController.CreatePost))
		auth.PUT("/posts/:id", response.Wrap(h.postController.UpdatePost))
		auth.DELETE("/posts/:id", response.Wrap(h.postController.DeletePost))
		auth.POST("/posts/:id/like", response.Wrap(h.postController.ToggleLike))
		auth.POST("/posts/:id/comments", response.Wrap(h.postController.CreateComment))
		auth.POST("/comments/:cid/like", response.Wrap(h.postController.ToggleCommentLike))
		auth.DELETE("/comments/:cid", response.Wrap(h.postController.DeleteComment))
	}
}

// ==================== Upload ====================

func (h *HttpServer) initUploadRouter(r *gin.Engine) {
	// 注册头像允许未登录上传，但独立限流；普通图片上传仍要求登录。
	publicUpload := r.Group("/api/upload")
	{
		publicUpload.POST("/register-avatar",
			middleware.RateLimit(h.cache, "register-avatar", 20, time.Hour),
			response.Wrap(h.uploadController.UploadImage))
	}
	upload := r.Group("/api/upload", h.auth.JWTAuth())
	{
		upload.POST("/image", response.Wrap(h.uploadController.UploadImage))
	}
}

// ==================== Admin ====================

func (h *HttpServer) initAdminRouter(r *gin.Engine) {
	admin := r.Group("/api/admin", h.auth.JWTAuth(), middleware.AdminRequired())
	{
		admin.GET("/users", response.Wrap(h.adminController.ListUsers))
		admin.GET("/users/import/template", h.adminController.DownloadUserImportTemplate)
		admin.POST("/users/import", response.Wrap(h.adminController.ImportUsers))
		admin.PUT("/users/:id/status", response.Wrap(h.adminController.SetUserStatus))
		admin.PUT("/users/:id/role", response.Wrap(h.adminController.SetUserRole))
		admin.GET("/judge/status", response.Wrap(h.adminController.JudgeStatus))
		admin.GET("/languages", response.Wrap(h.languageController.AdminList))
		admin.POST("/languages", response.Wrap(h.languageController.Create))
		admin.PUT("/languages/:id", response.Wrap(h.languageController.Update))
		admin.DELETE("/languages/:id", response.Wrap(h.languageController.Delete))
		admin.GET("/posts/pending", response.Wrap(h.postController.ReviewListPending))
		admin.PUT("/posts/:id/review", response.Wrap(h.postController.ReviewPost))

		// 比赛榜从 submissions 明细整场重算（榜单漂移时手动重建；路①下重判已不需要它）
		admin.POST("/contest/:id/recompute", response.Wrap(h.contestController.RecomputeContest))

		// 重判（超管）：单条 / 整场 / 某比赛某题
		admin.POST("/submission/:id/rejudge", response.Wrap(h.submissionController.Rejudge))
		admin.POST("/contest/:id/rejudge", response.Wrap(h.submissionController.RejudgeContest))
		admin.POST("/contest/:id/problem/:pid/rejudge", response.Wrap(h.submissionController.RejudgeContestProblem))

		// 远程账号管理
		admin.GET("/remote-account", response.Wrap(h.remoteController.ListAccounts))
		admin.POST("/remote-account", response.Wrap(h.remoteController.CreateAccount))
		admin.PUT("/remote-account/:id", response.Wrap(h.remoteController.UpdateAccount))
		admin.DELETE("/remote-account/:id", response.Wrap(h.remoteController.DeleteAccount))
	}
}
