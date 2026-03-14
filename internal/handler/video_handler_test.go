package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"fiap-x-api/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


type MockVideoRepo struct {
	mock.Mock
}

func (m *MockVideoRepo) FindByID(id string) (*domain.Video, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Video), args.Error(1)
}

func (m *MockVideoRepo) Update(v *domain.Video) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockVideoRepo) FindByUserID(uid string) ([]domain.Video, error) {
	args := m.Called(uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Video), args.Error(1)
}

func (m *MockVideoRepo) Create(v *domain.Video) error {
	args := m.Called(v)
	return args.Error(0)
}

type MockS3Adapter struct {
	mock.Mock
}

func (m *MockS3Adapter) UploadVideo(file io.Reader, fileName string) (string, error) {
	args := m.Called(file, fileName)
	return args.String(0), args.Error(1)
}

func (m *MockS3Adapter) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockS3Adapter) EnsureBucketExists(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockSQSAdapter struct {
	mock.Mock
}

func (m *MockSQSAdapter) PublishVideoJob(videoID string) error {
	args := m.Called(videoID)
	return args.Error(0)
}


func TestDownloadZip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Erro: Vídeo não encontrado no Banco", func(t *testing.T) {
		mockRepo := new(MockVideoRepo)
		mockRepo.On("FindByID", "123").Return((*domain.Video)(nil), errors.New("not found"))

		h := &VideoHandler{
			Repo:      mockRepo,
			S3Adapter: nil,
		}

		r := gin.New()
		r.GET("/download/:id", h.DownloadZip)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/download/123", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Erro: Processamento incompleto", func(t *testing.T) {
		video := &domain.Video{ID: "123", Status: domain.StatusProcessing, S3OutputKey: ""}
		mockRepo := new(MockVideoRepo)
		mockRepo.On("FindByID", "123").Return(video, nil)

		h := &VideoHandler{Repo: mockRepo}

		r := gin.New()
		r.GET("/download/:id", h.DownloadZip)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/download/123", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "pendente")
	})
}

func TestDownloadZip_S3Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockVideoRepo)
	mockS3 := new(MockS3Adapter)

	t.Run("Erro: Falha na busca do arquivo no S3", func(t *testing.T) {
		video := &domain.Video{ID: "123", S3OutputKey: "path.zip"}
		mockRepo.On("FindByID", "123").Return(video, nil)

		mockS3.On("GetObject", mock.Anything, "path.zip").Return(nil, errors.New("s3 connection error"))

		h := &VideoHandler{
			Repo:       mockRepo,
			S3Adapter:  mockS3,
			SQSAdapter: new(MockSQSAdapter),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "123"}}

		h.DownloadZip(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Erro ao buscar arquivo")
	})
}

func TestUploadVideo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Caminho Feliz: Upload e SQS ok", func(t *testing.T) {
		mockRepo := new(MockVideoRepo)
		mockS3 := new(MockS3Adapter)
		mockSQS := new(MockSQSAdapter)

		mockRepo.On("Create", mock.Anything).Return(nil)
		mockS3.On("UploadVideo", mock.Anything, mock.Anything).Return("video_final.mp4", nil)
		mockRepo.On("Update", mock.Anything).Return(nil)
		mockSQS.On("PublishVideoJob", mock.Anything).Return(nil)

		h := &VideoHandler{
			Repo:       mockRepo,
			S3Adapter:  mockS3,
			SQSAdapter: mockSQS,
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("video", "test.mp4")
		part.Write([]byte("fake content"))
		writer.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		c.Request = req
		c.Set("user_id", "user-uuid-123")

		h.Upload(c)

		assert.Equal(t, http.StatusAccepted, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Erro: Falha ao atualizar dados do vídeo após upload", func(t *testing.T) {
		mockRepo := new(MockVideoRepo)
		mockS3 := new(MockS3Adapter)

		mockRepo.On("Create", mock.Anything).Return(nil)
		mockS3.On("UploadVideo", mock.Anything, mock.Anything).Return("s3-key", nil)
		// Simulando erro no Update final
		mockRepo.On("Update", mock.Anything).Return(errors.New("erro banco"))

		h := &VideoHandler{Repo: mockRepo, S3Adapter: mockS3}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("video", "test.mp4")
		part.Write([]byte("data"))
		writer.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		c.Request = req
		c.Set("user_id", "user-123")

		h.Upload(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Erro ao atualizar dados")
	})

	t.Run("Erro: Falha ao enviar para o SQS", func(t *testing.T) {
		mockRepo := new(MockVideoRepo)
		mockS3 := new(MockS3Adapter)
		mockSQS := new(MockSQSAdapter)

		mockRepo.On("Create", mock.Anything).Return(nil)
		mockS3.On("UploadVideo", mock.Anything, mock.Anything).Return("key", nil)
		mockRepo.On("Update", mock.Anything).Return(nil)
		mockSQS.On("PublishVideoJob", mock.Anything).Return(errors.New("erro fila"))

		h := &VideoHandler{Repo: mockRepo, S3Adapter: mockS3, SQSAdapter: mockSQS}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("video", "test.mp4")
		part.Write([]byte("data"))
		writer.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		c.Request = req
		c.Set("user_id", "user-123")

		h.Upload(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Erro ao enviar para fila")
	})

	t.Run("Erro: Falha no FormFile2 (Erro de Banco no Create)", func(t *testing.T) {
		mockRepo := new(MockVideoRepo)
		mockRepo.On("Create", mock.Anything).Return(errors.New("db error on create"))

		h := &VideoHandler{Repo: mockRepo}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("video", "t.mp4")
		part.Write([]byte("data"))
		writer.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		c.Request = req
		c.Set("user_id", "user-123")

		h.Upload(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestListStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Sucesso: Retorna lista", func(t *testing.T) {
		mockRepo := new(MockVideoRepo)
		videos := []domain.Video{
			{ID: "1", UserID: "user-123", Status: domain.StatusCompleted},
		}

		mockRepo.On("FindByUserID", "user-123").Return(videos, nil)

		h := &VideoHandler{Repo: mockRepo}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("user_id", "user-123")

		h.ListStatus(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Erro: Falha ao buscar no Banco em ListStatus", func(t *testing.T) {
		mockRepo := new(MockVideoRepo)
		mockRepo.On("FindByUserID", "user-123").Return([]domain.Video(nil), errors.New("db error"))

		h := &VideoHandler{Repo: mockRepo}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", "user-123")

		h.ListStatus(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandlers_NoAuth(t *testing.T) {
	h := &VideoHandler{}

	t.Run("Upload sem user_id no contexto", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.Upload(c) // Contexto vazio
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ListStatus sem user_id no contexto", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.ListStatus(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
