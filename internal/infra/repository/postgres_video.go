package repository

import (
	"fiap-x-api/internal/domain"
	"gorm.io/gorm"
)

type PostgresVideoRepository struct {
	db *gorm.DB
}

func NewPostgresVideoRepository(db *gorm.DB) *PostgresVideoRepository {
	return &PostgresVideoRepository{db: db}
}

func (r *PostgresVideoRepository) Create(video *domain.Video) error {
	return r.db.Create(video).Error
}

func (r *PostgresVideoRepository) Update(video *domain.Video) error {
	return r.db.Save(video).Error
}

func (r *PostgresVideoRepository) FindByID(id string) (*domain.Video, error) {
	var video domain.Video
	err := r.db.First(&video, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *PostgresVideoRepository) FindByUserID(userID string) ([]domain.Video, error) {
	var videos []domain.Video
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&videos).Error
	return videos, err
}