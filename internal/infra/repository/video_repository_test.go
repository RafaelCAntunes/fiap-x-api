package repository

import (
	"fiap-x-api/internal/domain"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/glebarez/sqlite"
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

func TestVideoRepository_Queries(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Video{})

	repo := NewPostgresVideoRepository(db)
	videoID := "uuid-test"

	t.Run("Fluxo de busca e atualização", func(t *testing.T) {
		testVideo := &domain.Video{
			ID:     videoID,
			Status: "PENDING",
			UserID: "user-123",
		}
		err := db.Create(testVideo).Error
		assert.NoError(t, err)
		found, err := repo.FindByID(videoID)

		assert.NoError(t, err)
		assert.NotNil(t, found)

		if found != nil {
			assert.Equal(t, "PENDING", found.Status)

			// Testar atualização
			found.Status = "COMPLETED"
			err = repo.Update(found)
			assert.NoError(t, err)

			// Verificar se persistiu
			updated, _ := repo.FindByID(videoID)
			assert.Equal(t, "COMPLETED", updated.Status)
		}
	})
}
