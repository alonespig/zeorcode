package handler

import (
	"strconv"

	"zoj/internal/http/dto"
	"zoj/internal/model"
	"zoj/internal/service"
	"zoj/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type LanguageController struct {
	service *service.LanguageService
}

func NewLanguageController(languageService *service.LanguageService) *LanguageController {
	return &LanguageController{service: languageService}
}

func (h *LanguageController) List(c *gin.Context) (any, error) {
	langs, err := h.service.ListEnabled(c.Request.Context())
	if err != nil {
		return nil, err
	}
	return toLanguageItems(langs), nil
}

func (h *LanguageController) AdminList(c *gin.Context) (any, error) {
	langs, err := h.service.AdminList(c.Request.Context())
	if err != nil {
		return nil, err
	}
	return toLanguageItems(langs), nil
}

func (h *LanguageController) Create(c *gin.Context) (any, error) {
	var req dto.SaveLanguageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	language, err := h.service.Create(c.Request.Context(), service.SaveLanguageParams{
		Name: req.Name, Status: req.Status, Sort: req.Sort,
	})
	if err != nil {
		return nil, err
	}
	return toLanguageItem(language), nil
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
	return nil, h.service.Update(c.Request.Context(), id, service.SaveLanguageParams{
		Name: req.Name, Status: req.Status, Sort: req.Sort,
	})
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

func toLanguageItems(langs []model.Language) []dto.LanguageItem {
	items := make([]dto.LanguageItem, 0, len(langs))
	for _, l := range langs {
		items = append(items, dto.LanguageItem{
			ID:        l.ID,
			Name:      l.Name,
			Status:    l.Status,
			Sort:      l.Sort,
			CreatedAt: l.CreatedAt,
			UpdatedAt: l.UpdatedAt,
		})
	}
	return items
}

func toLanguageItem(l *model.Language) *dto.LanguageItem {
	return &dto.LanguageItem{
		ID:        l.ID,
		Name:      l.Name,
		Status:    l.Status,
		Sort:      l.Sort,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
