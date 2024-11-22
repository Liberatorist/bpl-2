package repository

import (
	"time"

	"gorm.io/gorm"
)

type ObjectiveType string

const (
	ITEM ObjectiveType = "ITEM"
)

type Objective struct {
	ID             int           `gorm:"primaryKey"`
	Name           string        `gorm:"not null"`
	RequiredNumber int           `gorm:"not null"`
	Conditions     []*Condition  `gorm:"foreignKey:ObjectiveID;constraint:OnDelete:CASCADE"`
	CategoryID     int           `gorm:"not null"`
	ObjectiveType  ObjectiveType `gorm:"not null;type:bpl2.objective_type"`
	ValidFrom      *time.Time    `gorm:"null"`
	ValidTo        *time.Time    `gorm:"null"`
}

type ObjectiveRepository struct {
	DB *gorm.DB
}

func NewObjectiveRepository(db *gorm.DB) *ObjectiveRepository {
	return &ObjectiveRepository{DB: db}
}

func (r *ObjectiveRepository) SaveObjective(objective *Objective) (*Objective, error) {
	result := r.DB.Save(objective)
	if result.Error != nil {
		return nil, result.Error
	}
	return objective, nil
}

func (r *ObjectiveRepository) GetObjectiveById(objectiveId int, preloads ...string) (*Objective, error) {
	var objective Objective
	query := r.DB
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	result := query.First(&objective, "id = ?", objectiveId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &objective, nil
}

func (r *ObjectiveRepository) DeleteObjective(objectiveId int) error {
	result := r.DB.Delete(&Objective{}, "id = ?", objectiveId)
	return result.Error
}

func (r *ObjectiveRepository) GetObjectivesByCategoryId(categoryId int) ([]*Objective, error) {
	var objectives []*Objective
	result := r.DB.Preload("Conditions").Find(&objectives, "category_id = ?", categoryId)
	if result.Error != nil {
		return nil, result.Error
	}
	return objectives, nil
}
