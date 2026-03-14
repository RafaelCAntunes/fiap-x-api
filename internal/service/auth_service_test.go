package service

import (
	"fiap-x-api/internal/domain"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAuthService_Login(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Erro ao abrir banco de teste: %v", err)
	}

	db.AutoMigrate(&domain.User{})

	authService := &AuthService{DB: db}
	userService := &UserService{DB: db}

	userService.CreateUser("rafael", "senha123")
	assert.NoError(t, err)

	t.Run("Sucesso: Login válido", func(t *testing.T) {
		token, err := authService.Login("rafael", "senha123")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Erro: Credenciais inválidas", func(t *testing.T) {
		_, err := authService.Login("rafael", "senha_errada")
		assert.Error(t, err)
	})

	t.Run("Erro: Login com senha incorreta", func(t *testing.T) {

		userService.CreateUser("user_teste", "senha_certa")

		token, err := authService.Login("user_teste", "senha_errada")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "senha inválida")
	})

	t.Run("Erro: Hash inválido no banco", func(t *testing.T) {
		userQuebrado := &domain.User{Username: "quebrado", Password: "nao_e_bcrypt"}
		db.Create(userQuebrado)

		_, err := authService.Login("quebrado", "123")
		assert.Error(t, err)
	})
}
