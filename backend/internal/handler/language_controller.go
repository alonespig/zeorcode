package handler

import (
	"strconv"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

type LanguageController struct {
	service *service.LanguageService
}

func NewLanguageController(languageService *service.LanguageService) *LanguageController {
	return &LanguageController{service: languageService}
}

func (h *LanguageController) List(c *gin.Context) (any, error) {
	return h.service.ListEnabled(c.Request.Context())
}

func (h *LanguageController) AdminList(c *gin.Context) (any, error) {
	return h.service.AdminList(c.Request.Context())
}

func (h *LanguageController) Create(c *gin.Context) (any, error) {
	var req dto.SaveLanguageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return h.service.Create(c.Request.Context(), &req)
}

func (h *LanguageController) Update(c *gin.Context) (any, error) {
	id, err := languageID(c)
	if err != nil {
		return nil, err
	}
	var req dto.SaveLanguageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return nil, h.service.Update(c.Request.Context(), id, &req)
}

func (h *LanguageController) Delete(c *gin.Context) (any, error) {
	id, err := languageID(c)
	if err != nil {
		return nil, err
	}
	return nil, h.service.Delete(c.Request.Context(), id)
}

func languageID(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return 0, errcode.ErrInvalidParams.WithMsg("非法语言编号")
	}
	return id, nil
}
