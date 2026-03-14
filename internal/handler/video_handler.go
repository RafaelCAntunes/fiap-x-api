package handler

import (
	"context"
	"fiap-x-api/internal/domain"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type S3ClientInterface interface {
	UploadVideo(file io.Reader, fileName string) (string, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
}

type SQSClientInterface interface {
	PublishVideoJob(videoID string) error
}

type VideoHandler struct {
	Repo       domain.VideoRepository
	S3Adapter  S3ClientInterface  
	SQSAdapter SQSClientInterface 
}


func (h *VideoHandler) Upload(c *gin.Context) {
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		log.Printf("Erro no FormFile: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "O campo 'video' é obrigatório"})
		return
	}
	defer file.Close()

	video := &domain.Video{
		UserID:   userID.(string),
		FileName: header.Filename,
		Status:   domain.StatusPending,
	}

	if err := h.Repo.Create(video); err != nil {
		log.Printf("Erro no FormFile2: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao registrar processamento"})
		return
	}

	s3Key, err := h.S3Adapter.UploadVideo(file, video.ID+"_"+header.Filename)
	if err != nil {
		log.Printf("Erro no FormFile3: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo no storage"})
		return
	}

	video.S3InputKey = s3Key
	if err := h.Repo.Update(video); err != nil {
		log.Printf("Erro no FormFile4: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar dados do vídeo"})
		return
	}

	if err := h.SQSAdapter.PublishVideoJob(video.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar para fila de processamento"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Vídeo recebido! O processamento iniciará em breve.",
		"video_id": video.ID,
		"status":   video.Status,
	})
}

func (h *VideoHandler) ListStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	videos, err := h.Repo.FindByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar histórico"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"total":   len(videos),
		"data":    videos,
	})
}

func (h *VideoHandler) DownloadZip(c *gin.Context) {
	videoID := c.Param("id")

	video, err := h.Repo.FindByID(videoID)
	if err != nil || video.S3OutputKey == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado ou processamento pendente"})
		return
	}
	log.Printf("id do video no S3: %v", video.S3OutputKey)

	body, err := h.S3Adapter.GetObject(context.TODO(), video.S3OutputKey)
	if err != nil {
		log.Printf("erro ao buscar no s3: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar arquivo"})
		return
	}
	defer body.Close()

	//Força o navegador a baixar o arquivo
	c.Header("Content-Disposition", "attachment; filename="+videoID+".zip")
	c.Header("Content-Type", "application/zip")

	io.Copy(c.Writer, body)
}
