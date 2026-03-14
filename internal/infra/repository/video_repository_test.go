package repository

import (
	"fiap-x-api/internal/domain"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestVideoEntity_Hooks(t *testing.T) {
	t.Run("Deve testar os hooks de UUID", func(t *testing.T) {
		v := &domain.Video{}
		u := &domain.User{}

		assert.NoError(t, v.BeforeCreate(nil))
		assert.NoError(t, u.BeforeCreate(nil))

		assert.NotEmpty(t, v.ID)
		assert.NotEmpty(t, u.ID)
	})
}
func TestPostgresVideoRepository_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	repo := NewPostgresVideoRepository(gormDB)

	video := &domain.Video{UserID: "123", FileName: "test.mp4"}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"videos\"").
		WithArgs(sqlmock.AnyArg(), "123", "test.mp4", "", "", "PENDING", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(video)
	assert.NoError(t, err)
}
