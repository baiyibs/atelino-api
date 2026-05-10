package UserService

import (
	"atelino/internal/auth"
	"atelino/internal/dto"
	"atelino/internal/model"
	"atelino/internal/repository/UserRepository"
	"atelino/internal/repository/ValidatorRepository"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const maxUserDevices = 3

var (
	ErrNotFound                = errors.New("没有找到指定的用户")
	ErrInvalidCredentials      = errors.New("用户名或密码错误")
	ErrVerificationCodeExpired = errors.New("验证码无效或已经过期，请重新获取。")
	ErrInvalidVerificationCode = errors.New("验证码无效")
	ErrEmailExists             = errors.New("该邮箱已注册")
	ErrInvalidToken            = errors.New("无效的刷新令牌")
	ErrTokenExpired            = errors.New("刷新令牌已失效")
)

type Service struct {
	userRepo      UserRepository.Interface
	txManager     UserRepository.TransactionManager
	validatorRepo ValidatorRepository.Interface
}

func New(
	userRepo UserRepository.Interface,
	txManager UserRepository.TransactionManager,
	validatorRepo ValidatorRepository.Interface,
) *Service {
	return &Service{
		userRepo:      userRepo,
		txManager:     txManager,
		validatorRepo: validatorRepo,
	}
}

func (s *Service) Register(request dto.RegisterRequest) error {
	storedCode, err := s.validatorRepo.GetCode(request.Email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrVerificationCodeExpired
		}
		return err
	}

	if storedCode != request.Code {
		return ErrInvalidVerificationCode
	}

	if _, err := s.userRepo.FindByEmail(request.Email); err == nil {
		return ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	s.validatorRepo.DeleteCode(request.Email)

	newUser := model.User{
		Email:    request.Email,
		Username: request.Username,
		Password: string(hashedPassword),
		Role:     "user",
	}
	if err := s.userRepo.Create(&newUser); err != nil {
		return err
	}

	return nil
}

func (s *Service) Login(request dto.LoginRequest) (dto.TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.TokenResponse{}, ErrInvalidCredentials
		}
		return dto.TokenResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return dto.TokenResponse{}, ErrInvalidCredentials
	}

	var tokens dto.TokenResponse
	err = s.txManager.Transaction(func(_ UserRepository.Interface, refreshTokenRepo UserRepository.RefreshTokenInterface) error {
		if err := limitUserDevices(refreshTokenRepo, user.ID); err != nil {
			return err
		}

		accessToken, err := auth.GenerateAccessToken(user.ID, user.Role)
		if err != nil {
			return fmt.Errorf("生成访问令牌失败: %w", err)
		}

		rawRefresh, refreshHash, err := auth.GenerateRefreshToken()
		if err != nil {
			return fmt.Errorf("生成刷新令牌失败: %w", err)
		}

		if err := refreshTokenRepo.Create(&model.RefreshToken{
			UserID:    user.ID,
			TokenHash: refreshHash,
			ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
			CreatedAt: time.Now().UTC(),
		}); err != nil {
			return fmt.Errorf("存储刷新令牌失败: %w", err)
		}

		tokens = dto.TokenResponse{AccessToken: accessToken, RefreshToken: rawRefresh}
		return nil
	})
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return tokens, nil
}

func (s *Service) GetByID(request dto.UserIDRequest) (dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserResponse{}, ErrNotFound
		}
		return dto.UserResponse{}, err
	}

	return dto.NewUserResponse(user), nil
}

func (s *Service) List(request dto.UserListRequest, pageSize int) ([]dto.UserResponse, int64, error) {
	page := request.Page
	if page == 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	total, err := s.userRepo.Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := s.userRepo.List(pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return dto.NewUserResponses(list), total, nil
}

func (s *Service) Refresh(request dto.RefreshTokenRequest) (dto.TokenResponse, error) {
	hash := auth.HashRefreshToken(request.RefreshToken)

	var tokens dto.TokenResponse
	err := s.txManager.Transaction(func(userRepo UserRepository.Interface, refreshTokenRepo UserRepository.RefreshTokenInterface) error {
		oldToken, err := refreshTokenRepo.FindByHashForUpdate(hash)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTokenExpired
			}
			return err
		}

		if oldToken.RevokedAt != nil || time.Now().UTC().After(oldToken.ExpiresAt) {
			return ErrTokenExpired
		}

		user, err := userRepo.FindByID(oldToken.UserID)
		if err != nil {
			return fmt.Errorf("查找用户失败: %w", err)
		}

		accessToken, err := auth.GenerateAccessToken(user.ID, user.Role)
		if err != nil {
			return fmt.Errorf("生成访问令牌失败: %w", err)
		}

		rawRefresh, newHash, err := auth.GenerateRefreshToken()
		if err != nil {
			return fmt.Errorf("生成刷新令牌失败: %w", err)
		}

		oldToken.RevokedAt = new(time.Now().UTC())
		if err := refreshTokenRepo.Save(&oldToken); err != nil {
			return fmt.Errorf("吊销刷新令牌失败: %w", err)
		}

		if err := refreshTokenRepo.Create(&model.RefreshToken{
			UserID:    oldToken.UserID,
			TokenHash: newHash,
			ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		}); err != nil {
			return fmt.Errorf("存储刷新令牌失败: %w", err)
		}

		tokens = dto.TokenResponse{AccessToken: accessToken, RefreshToken: rawRefresh}
		return nil
	})
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return tokens, nil
}

func (s *Service) Logout(request dto.LogoutRequest) error {
	return s.txManager.Transaction(func(_ UserRepository.Interface, refreshTokenRepo UserRepository.RefreshTokenInterface) error {
		tokens, err := refreshTokenRepo.FindActiveByUserIDForUpdate(request.UserID)
		if err != nil {
			return err
		}

		if len(tokens) == 0 {
			return nil
		}

		now := time.Now().UTC()
		for i := range tokens {
			tokens[i].RevokedAt = &now
		}
		return refreshTokenRepo.SaveAll(tokens)
	})
}

func limitUserDevices(refreshTokenRepo UserRepository.RefreshTokenInterface, userID uint64) error {
	validTokens, err := refreshTokenRepo.FindValidByUserIDForUpdate(userID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("查询用户的刷新令牌时失败: %w", err)
	}

	currentCount := len(validTokens)
	if currentCount < maxUserDevices {
		return nil
	}

	revokeCount := currentCount - maxUserDevices + 1
	now := time.Now().UTC()
	for i := 0; i < revokeCount && i < currentCount; i++ {
		validTokens[i].RevokedAt = &now
		if err := refreshTokenRepo.Save(&validTokens[i]); err != nil {
			return fmt.Errorf("吊销旧的刷新令牌时失败: %w", err)
		}
	}

	return nil
}
