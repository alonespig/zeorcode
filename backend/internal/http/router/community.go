package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Post(博客 / 讨论 / 题解) ====================

func (h *HttpServer) initPostRouter(r *gin.Engine) {
	pub := r.Group("/api")
	{
		pub.GET("/posts", middleware.JWTAuthOptional(h.auth), response.Wrap(h.postController.ListPosts))
		pub.GET("/posts/:id", middleware.JWTAuthOptional(h.auth), response.Wrap(h.postController.GetPost))
		pub.GET("/posts/:id/comments", middleware.JWTAuthOptional(h.auth), response.Wrap(h.postController.ListComments))
	}

	auth := r.Group("/api", middleware.JWTAuth(h.auth))
	{
		auth.POST("/posts", response.Wrap(h.postController.CreatePost))
		auth.PUT("/posts/:id", response.Wrap(h.postController.UpdatePost))
		auth.DELETE("/posts/:id", response.Wrap(h.postController.DeletePost))
		auth.POST("/posts/:id/like", response.Wrap(h.postController.ToggleLike))
		auth.POST("/posts/:id/comments", response.Wrap(h.postController.CreateComment))
		auth.POST("/comments/:cid/like", response.Wrap(h.postController.ToggleCommentLike))
		auth.DELETE("/comments/:cid", response.Wrap(h.postController.DeleteComment))
	}

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.GET("/posts/pending", response.Wrap(h.postController.ReviewListPending))
		admin.PUT("/posts/:id/review", response.Wrap(h.postController.ReviewPost))
	}
}
