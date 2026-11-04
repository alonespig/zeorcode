package handler

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"

	"github.com/alonespig/zeorcode/backend/internal/judgeartifact"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type JudgeArtifactController struct {
	store *judgeartifact.Store
	token string
}

func NewJudgeArtifactController() *JudgeArtifactController {
	return &JudgeArtifactController{
		store: judgeartifact.NewStore(viper.GetString("judge.data_dir")),
		token: strings.TrimSpace(viper.GetString("judge.artifact_token")),
	}
}

func (h *JudgeArtifactController) Download(c *gin.Context) {
	if h.token == "" {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "judge artifact service is not configured"})
		return
	}
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	provided := parts[1]
	providedHash := sha256.Sum256([]byte(provided))
	expectedHash := sha256.Sum256([]byte(h.token))
	if subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || problemID <= 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	want := judgeartifact.Descriptor{
		ProblemID: problemID,
		Version:   strings.TrimSpace(c.Query("version")),
		Digest:    strings.ToLower(strings.TrimSpace(c.Query("digest"))),
	}
	current, err := h.store.Describe(c.Request.Context(), problemID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "problem artifact unavailable"})
		return
	}
	if current.Version != want.Version || current.Digest != want.Digest {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "problem artifact version changed",
			"version": current.Version,
			"digest":  current.Digest,
		})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=problem-"+strconv.FormatInt(problemID, 10)+"-"+want.Version+".zip")
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Artifact-Digest", want.Digest)
	if err := h.store.WriteArchive(c.Request.Context(), c.Writer, want); err != nil {
		c.Error(err) // response may already be streaming; keep details in server logs only
	}
}
