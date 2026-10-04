package router

import (
	"time"

	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Team ====================

func (h *HttpServer) initTeamRouter(r *gin.Engine) {
	// 列表与详情用可选登录：未登录也能浏览团队，只是看不到「我的角色」和内部内容
	pub := r.Group("/api")
	{
		pub.GET("/team", middleware.JWTAuthOptional(h.auth), response.Wrap(h.teamController.List))
		pub.GET("/team/:id", middleware.JWTAuthOptional(h.auth), response.Wrap(h.teamController.Detail))
		pub.GET("/team/:id/member", middleware.JWTAuthOptional(h.auth), response.Wrap(h.teamController.ListMembers))
	}

	// 写操作一律要登录；创建团队额外要求站点管理员，其他操作仍由 service 层
	// 按团队角色判定，不能统一使用全站 AdminRequired。
	auth := r.Group("/api", middleware.JWTAuth(h.auth))
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
		auth.POST("/team/:id/member/import/manual",
			middleware.RateLimit(h.cache, "team-student-import-manual", 10, time.Hour),
			response.Wrap(h.teamController.ImportStudentsManual))
		auth.PUT("/team/:id/member/:uid/role", response.Wrap(h.teamController.SetMemberRole))
		auth.DELETE("/team/:id/member/:uid", response.Wrap(h.teamController.RemoveMember))
	}
}
