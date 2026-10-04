package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Language ====================

func (h *HttpServer) initLanguageRouter(r *gin.Engine) {
	pub := r.Group("/api")
	{
		pub.GET("/languages", response.Wrap(h.languageController.List))
	}

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.GET("/languages", response.Wrap(h.languageController.AdminList))
		admin.POST("/languages", response.Wrap(h.languageController.Create))
		admin.PUT("/languages/:id", response.Wrap(h.languageController.Update))
		admin.DELETE("/languages/:id", response.Wrap(h.languageController.Delete))
	}
}
