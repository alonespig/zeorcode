package dto

import "time"

type LanguageItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	Sort      int       `json:"sort"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SaveLanguageReq struct {
	Name   string `json:"name" binding:"required,max=32"`
	Status int    `json:"status" binding:"oneof=0 1"`
	Sort   int    `json:"sort" binding:"gte=0"`
}
