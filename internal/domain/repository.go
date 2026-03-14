package domain

// VideoRepository define o contrato para persistência de vídeos
type VideoRepository interface {
    Create(video *Video) error
    Update(video *Video) error
    FindByID(id string) (*Video, error)
    FindByUserID(userID string) ([]Video, error)
} 

