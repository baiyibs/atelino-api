package UserRepository

import (
	"backend/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Interface interface {
	FindByEmail(email string) (model.User, error)
	FindByID(id uint) (model.User, error)
	Create(user *model.User) error
}

type RefreshTokenInterface interface {
	FindValidByUserIDForUpdate(userID uint, now time.Time) ([]model.RefreshToken, error)
	FindByHashForUpdate(hash string) (model.RefreshToken, error)
	FindActiveByUserIDForUpdate(userID string) ([]model.RefreshToken, error)
	Create(token *model.RefreshToken) error
	Save(token *model.RefreshToken) error
	SaveAll(tokens []model.RefreshToken) error
}

type TransactionManager interface {
	Transaction(fn func(Interface, RefreshTokenInterface) error) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *UserRepository) FindByID(id uint) (model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return user, err
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) FindValidByUserIDForUpdate(userID uint, now time.Time) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("created_at ASC").
		Find(&tokens).Error
	return tokens, err
}

func (r *RefreshTokenRepository) FindByHashForUpdate(hash string) (model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", hash).
		First(&token).Error
	return token, err
}

func (r *RefreshTokenRepository) FindActiveByUserIDForUpdate(userID string) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Find(&tokens).Error
	return tokens, err
}

func (r *RefreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) Save(token *model.RefreshToken) error {
	return r.db.Save(token).Error
}

func (r *RefreshTokenRepository) SaveAll(tokens []model.RefreshToken) error {
	return r.db.Save(&tokens).Error
}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func (m *GormTransactionManager) Transaction(fn func(Interface, RefreshTokenInterface) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		return fn(NewUserRepository(tx), NewRefreshTokenRepository(tx))
	})
}
