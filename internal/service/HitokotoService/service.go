package HitokotoService

import (
	"atelino/internal/dto"
	"atelino/internal/model"
	"atelino/internal/repository/HitokotoRepository"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("没有找到指定的一言")
	ErrDuplicate = errors.New("该一言已经存在")
)

type Service struct {
	hitokotoRepo HitokotoRepository.Interface
}

func New(hitokotoRepo HitokotoRepository.Interface) *Service {
	return &Service{hitokotoRepo: hitokotoRepo}
}

func (s *Service) Create(request dto.CreateHitokotoRequest) (dto.HitokotoResponse, error) {
	hitokoto := model.Hitokoto{Content: request.Content}
	if err := s.hitokotoRepo.Create(&hitokoto); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return dto.HitokotoResponse{}, ErrDuplicate
		}
		return dto.HitokotoResponse{}, err
	}
	return dto.NewHitokotoResponse(hitokoto), nil
}

func (s *Service) DeleteByID(request dto.HitokotoIDRequest) error {
	rowsAffected, err := s.hitokotoRepo.DeleteByID(request.ID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) List(request dto.HitokotoListRequest, pageSize int) ([]dto.HitokotoResponse, int64, error) {
	page := request.Page
	if page == 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	total, err := s.hitokotoRepo.Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := s.hitokotoRepo.List(pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return dto.NewHitokotoResponses(list), total, nil
}

func (s *Service) GetByID(request dto.HitokotoIDRequest) (dto.HitokotoResponse, error) {
	hitokoto, err := s.hitokotoRepo.GetByID(request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.HitokotoResponse{}, ErrNotFound
		}
		return dto.HitokotoResponse{}, err
	}
	return dto.NewHitokotoResponse(hitokoto), nil
}

func (s *Service) Random() (dto.HitokotoResponse, error) {
	hitokoto, err := s.hitokotoRepo.Random()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.HitokotoResponse{}, ErrNotFound
		}
		return dto.HitokotoResponse{}, err
	}
	return dto.NewHitokotoResponse(hitokoto), nil
}
