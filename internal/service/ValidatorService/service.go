package ValidatorService

import (
	"backend/internal/dto"
	"backend/internal/repository/ValidatorRepository"
	"backend/pkg/email"
	"errors"
	"fmt"
	"time"
)

var ErrCooldown = errors.New("验证码发送过于频繁")

type Service struct {
	validatorRepo ValidatorRepository.Interface
}

func New(validatorRepo ValidatorRepository.Interface) *Service {
	return &Service{validatorRepo: validatorRepo}
}

func (s *Service) SendCode(request dto.SendVerificationCodeRequest) error {
	ok, err := s.validatorRepo.AcquireCooldown(request.To, time.Minute)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCooldown
	}

	code, err := email.GenerateVerificationCode()
	if err != nil {
		s.validatorRepo.DeleteCooldown(request.To)
		return fmt.Errorf("generate verification code: %w", err)
	}

	if err := s.validatorRepo.StoreCode(request.To, code, 5*time.Minute); err != nil {
		return err
	}

	if err := email.SendVerificationCode(request.To, code); err != nil {
		s.validatorRepo.DeleteCooldown(request.To)
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}
