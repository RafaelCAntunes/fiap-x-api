package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// Função auxiliar para gerar tokens válidos nos testes
func generateTestToken(userId string) string {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, _ := token.SignedString([]byte("chave-secreta-para-teste"))
	return t
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Erro: Token não fornecido", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token não fornecido")
	})

	t.Run("Erro: Token inválido", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer token-errado")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token inválido")
	})

	t.Run("Sucesso: Token válido", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/protected", func(c *gin.Context) {
			val, _ := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{"user_id": val})
		})

		token := generateTestToken("user-test-123")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user-test-123")
	})

	
}
