package HitokotoRepository

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type Interface interface {
	Create(hitokoto *model.Hitokoto) error
	DeleteByID(id int) (int64, error)
	Count() (int64, error)
	List(limit, offset int) ([]model.Hitokoto, error)
	GetByID(id int) (model.Hitokoto, error)
	Random() (model.Hitokoto, error)
}

type HitokotoRepository struct {
	db *gorm.DB
}

func NewHitokotoRepository(db *gorm.DB) *HitokotoRepository {
	return &HitokotoRepository{db: db}
}

func (r *HitokotoRepository) Create(hitokoto *model.Hitokoto) error {
	return r.db.Create(hitokoto).Error
}

func (r *HitokotoRepository) DeleteByID(id int) (int64, error) {
	result := r.db.Where("id = ?", id).Delete(&model.Hitokoto{})
	return result.RowsAffected, result.Error
}

func (r *HitokotoRepository) Count() (int64, error) {
	var total int64
	err := r.db.Model(&model.Hitokoto{}).Count(&total).Error
	return total, err
}

func (r *HitokotoRepository) List(limit, offset int) ([]model.Hitokoto, error) {
	var list []model.Hitokoto
	err := r.db.Order("id asc").Limit(limit).Offset(offset).Find(&list).Error
	return list, err
}

func (r *HitokotoRepository) GetByID(id int) (model.Hitokoto, error) {
	var hitokoto model.Hitokoto
	err := r.db.Where("id = ?", id).First(&hitokoto).Error
	return hitokoto, err
}

func (r *HitokotoRepository) Random() (model.Hitokoto, error) {
	var hitokoto model.Hitokoto
	err := r.db.Order("RANDOM()").Limit(1).First(&hitokoto).Error
	return hitokoto, err
}
