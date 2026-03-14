package service

import (
	"errors"
	"fiap-x-api/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtKey = []byte("chave-secreta-para-teste") 

type AuthService struct {
	DB *gorm.DB
}

func (s *AuthService) Login(username, password string) (string, error) {
	var user domain.User
	
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return "", errors.New("usuário não encontrado")
	}

		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("senha inválida")
	}

	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // Expira em 1h
	})

	return token.SignedString(jwtKey)
}
